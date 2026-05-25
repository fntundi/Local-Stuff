package services

import (
	"context"
	"errors"
	"time"

	"sentinel-noc/internal/config"
	"sentinel-noc/internal/middleware"
	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"

	"github.com/pquerna/otp/totp"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrRequires2FA        = errors.New("2FA required")
	ErrInvalid2FACode     = errors.New("invalid 2FA code")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already registered")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo      *repository.UserRepository
	cryptoService *CryptoService
	auditService  *AuditService
	config        *config.Config
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	cryptoService *CryptoService,
	auditService *AuditService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		cryptoService: cryptoService,
		auditService:  auditService,
		config:        cfg,
	}
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	User         *models.UserResponse
	AccessToken  string
	RefreshToken string
	Requires2FA  bool
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.UserRegisterRequest, ipAddress string) (*models.UserResponse, error) {
	// Validate password strength
	if err := s.cryptoService.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Check if username exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Determine role
	role := req.Role
	if role == "" {
		role = string(models.RoleViewer)
	}

	// First user is always admin
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if userCount == 0 {
		role = string(models.RoleAdmin)
	} else if role == string(models.RoleAdmin) {
		// Only allow admin creation by existing admins (handled by handler)
		return nil, errors.New("only admins can create admin users")
	}

	// Hash password
	hashedPassword, err := s.cryptoService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &repository.User{
		ID:                  GenerateUUID(),
		Username:            req.Username,
		Email:               req.Email,
		PasswordHash:        hashedPassword,
		Role:                role,
		TOTPEnabled:         false,
		CreatedAt:           time.Now().UTC(),
		FailedLoginAttempts: 0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Audit log
	_ = s.auditService.Log(ctx, user.ID, user.Username, "user_registered", "user", user.ID, ipAddress, map[string]interface{}{"role": role})

	return &models.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		TOTPEnabled: user.TOTPEnabled,
		CreatedAt:   user.CreatedAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.UserLoginRequest, ipAddress string) (*AuthResult, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().UTC().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}

	// Clear lock if expired
	if user.LockedUntil != nil && time.Now().UTC().After(*user.LockedUntil) {
		_ = s.userRepo.ResetFailedAttempts(ctx, user.ID)
	}

	// Verify password
	if !s.cryptoService.VerifyPassword(req.Password, user.PasswordHash) {
		// Increment failed attempts
		var lockUntil *time.Time
		if user.FailedLoginAttempts >= 4 { // Will be 5 after increment
			lockTime := time.Now().UTC().Add(15 * time.Minute)
			lockUntil = &lockTime
		}
		_ = s.userRepo.IncrementFailedAttempts(ctx, user.ID, lockUntil)

		_ = s.auditService.Log(ctx, user.ID, user.Username, "login_failed", "auth", "", ipAddress, map[string]interface{}{"reason": "invalid_password"})

		return nil, ErrInvalidCredentials
	}

	// Check 2FA if enabled
	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			return &AuthResult{Requires2FA: true}, nil
		}

		// Decrypt and verify TOTP
		secret, err := s.cryptoService.Decrypt(user.TOTPSecret)
		if err != nil {
			return nil, err
		}

		if !totp.Validate(req.TOTPCode, secret) {
			_ = s.auditService.Log(ctx, user.ID, user.Username, "login_failed", "auth", "", ipAddress, map[string]interface{}{"reason": "invalid_2fa"})
			return nil, ErrInvalid2FACode
		}
	}

	// Reset failed attempts and update last login
	_ = s.userRepo.ResetFailedAttempts(ctx, user.ID)

	// Generate tokens
	accessToken, err := middleware.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
		s.config.JWTSecret,
		s.config.AccessTokenExpireMin,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := middleware.GenerateRefreshToken(
		user.ID,
		s.config.JWTSecret,
		s.config.RefreshTokenExpireDays,
	)
	if err != nil {
		return nil, err
	}

	// Audit log
	_ = s.auditService.Log(ctx, user.ID, user.Username, "login_success", "auth", "", ipAddress, nil)

	return &AuthResult{
		User: &models.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        user.Role,
			TOTPEnabled: user.TOTPEnabled,
			CreatedAt:   user.CreatedAt,
			LastLogin:   user.LastLogin,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Parse refresh token
	claims, err := middleware.ParseRefreshToken(refreshToken, s.config.JWTSecret)
	if err != nil {
		return "", err
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	// Generate new access token
	return middleware.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
		s.config.JWTSecret,
		s.config.AccessTokenExpireMin,
	)
}

// Setup2FA generates a new TOTP secret for a user
func (s *AuthService) Setup2FA(ctx context.Context, userID string) (*models.TOTPSetupResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.TOTPEnabled {
		return nil, errors.New("2FA already enabled")
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Sentinel NOC",
		AccountName: user.Username,
	})
	if err != nil {
		return nil, err
	}

	// Encrypt and store pending secret
	encryptedSecret, err := s.cryptoService.Encrypt(key.Secret())
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Update(ctx, userID, map[string]interface{}{
		"totp_secret_pending": encryptedSecret,
	})
	if err != nil {
		return nil, err
	}

	return &models.TOTPSetupResponse{
		Secret: key.Secret(),
		URI:    key.URL(),
		QRData: key.URL(),
	}, nil
}

// Verify2FA verifies a TOTP code and enables 2FA
func (s *AuthService) Verify2FA(ctx context.Context, userID, code, ipAddress string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.TOTPSecretPending == "" {
		return errors.New("no pending 2FA setup")
	}

	// Decrypt pending secret
	secret, err := s.cryptoService.Decrypt(user.TOTPSecretPending)
	if err != nil {
		return err
	}

	// Verify code
	if !totp.Validate(code, secret) {
		return ErrInvalid2FACode
	}

	// Enable 2FA
	err = s.userRepo.UpdateWithUnset(ctx, userID,
		map[string]interface{}{
			"totp_enabled": true,
			"totp_secret":  user.TOTPSecretPending,
		},
		map[string]interface{}{
			"totp_secret_pending": "",
		},
	)
	if err != nil {
		return err
	}

	_ = s.auditService.Log(ctx, userID, user.Username, "2fa_enabled", "user", userID, ipAddress, nil)

	return nil
}

// Disable2FA disables 2FA for a user
func (s *AuthService) Disable2FA(ctx context.Context, userID, code, ipAddress string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if !user.TOTPEnabled {
		return errors.New("2FA not enabled")
	}

	// Decrypt and verify code
	secret, err := s.cryptoService.Decrypt(user.TOTPSecret)
	if err != nil {
		return err
	}

	if !totp.Validate(code, secret) {
		return ErrInvalid2FACode
	}

	// Disable 2FA
	err = s.userRepo.UpdateWithUnset(ctx, userID,
		map[string]interface{}{
			"totp_enabled": false,
		},
		map[string]interface{}{
			"totp_secret": "",
		},
	)
	if err != nil {
		return err
	}

	_ = s.auditService.Log(ctx, userID, user.Username, "2fa_disabled", "user", userID, ipAddress, nil)

	return nil
}
