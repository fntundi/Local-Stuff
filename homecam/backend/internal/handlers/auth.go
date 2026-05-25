// Package handlers provides HTTP request handlers
package handlers

import (
	"net/http"

	"sentinel-noc/internal/config"
	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService   *services.AuthService
	userRepo      *repository.UserRepository
	cryptoService *services.CryptoService
	auditService  *services.AuditService
	config        *config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	authService *services.AuthService,
	userRepo *repository.UserRepository,
	cryptoService *services.CryptoService,
	auditService *services.AuditService,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		userRepo:      userRepo,
		cryptoService: cryptoService,
		auditService:  auditService,
		config:        cfg,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		switch err {
		case services.ErrUsernameExists:
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Username already exists"})
		case services.ErrEmailExists:
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Email already registered"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid credentials"})
		case services.ErrAccountLocked:
			c.JSON(http.StatusLocked, gin.H{"detail": "Account temporarily locked"})
		case services.ErrInvalid2FACode:
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid 2FA code"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Login failed"})
		}
		return
	}

	if result.Requires2FA {
		c.JSON(http.StatusOK, models.Requires2FAResponse{
			Requires2FA: true,
			Message:     "2FA code required",
		})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "bearer",
		User:         *result.User,
	})
}

// RefreshToken handles token refresh.
// The refresh token must be sent in the JSON body only — never in the URL,
// to prevent it from being captured in access logs, browser history, or referrer headers.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "refresh_token required in request body"})
		return
	}
	refreshToken := body.RefreshToken

	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Refresh token required"})
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "bearer",
	})
}

// GetCurrentUser returns the current user's information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	user, err := h.userRepo.FindByID(c.Request.Context(), userIDStr)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		TOTPEnabled: user.TOTPEnabled,
		CreatedAt:   user.CreatedAt,
		LastLogin:   user.LastLogin,
	})
}

// Setup2FA initiates 2FA setup
func (h *AuthHandler) Setup2FA(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	result, err := h.authService.Setup2FA(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Verify2FA verifies 2FA setup
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req models.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	err := h.authService.Verify2FA(c.Request.Context(), userIDStr, req.Code, c.ClientIP())
	if err != nil {
		if err == services.ErrInvalid2FACode {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid verification code"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable2FA disables 2FA
func (h *AuthHandler) Disable2FA(c *gin.Context) {
	var req models.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	err := h.authService.Disable2FA(c.Request.Context(), userIDStr, req.Code, c.ClientIP())
	if err != nil {
		if err == services.ErrInvalid2FACode {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid verification code"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}
