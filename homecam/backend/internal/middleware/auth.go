// Package middleware provides HTTP middleware for the application
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims stored in JWT tokens
type JWTClaims struct {
	UserID   string `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

// JWTAuth middleware validates JWT tokens and extracts user information
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Token expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token claims"})
			return
		}

		// Verify token type is "access"
		if claims.Type != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token type"})
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole middleware checks if the user has one of the required roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"detail": "Invalid role format"})
			return
		}

		// Check if user's role is in the allowed roles
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Insufficient permissions"})
	}
}

// GenerateAccessToken creates a new access token
func GenerateAccessToken(userID, username, role, secret string, expireMinutes int) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a new refresh token
func GenerateRefreshToken(userID, secret string, expireDays int) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseRefreshToken validates and extracts claims from a refresh token
func ParseRefreshToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.Type != "refresh" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
