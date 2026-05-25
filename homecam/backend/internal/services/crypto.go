// Package services provides business logic implementations
package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"regexp"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CryptoService handles encryption and hashing operations
type CryptoService struct {
	encryptionKey []byte
}

// NewCryptoService creates a new CryptoService
func NewCryptoService(key string) *CryptoService {
	// Decode base64 key or use as-is
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		// If not base64, use the key directly (padded/truncated to 32 bytes)
		decodedKey = []byte(key)
	}

	// Ensure key is 32 bytes for AES-256
	if len(decodedKey) < 32 {
		padded := make([]byte, 32)
		copy(padded, decodedKey)
		decodedKey = padded
	} else if len(decodedKey) > 32 {
		decodedKey = decodedKey[:32]
	}

	return &CryptoService{
		encryptionKey: decodedKey,
	}
}

// HashPassword creates a bcrypt hash of the password
func (s *CryptoService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches a hash
func (s *CryptoService) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks if password meets security requirements.
// Requirements: 12–128 chars, at least one uppercase, lowercase, digit, and special character.
func (s *CryptoService) ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if len(password) > 128 {
		return errors.New("password must be at most 128 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// Encrypt encrypts plaintext using AES-GCM
func (s *CryptoService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (s *CryptoService) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateUUID creates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// hostnamePattern validates RFC-1123 hostnames used as camera addresses.
var hostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// ValidateIPAddress accepts a valid IPv4, IPv6, or RFC-1123 hostname.
// It uses net.ParseIP for IP validation to reject octets like 999 that pass naive regexes.
func ValidateIPAddress(ip string) bool {
	if net.ParseIP(ip) != nil {
		return true
	}
	return hostnamePattern.MatchString(ip)
}
