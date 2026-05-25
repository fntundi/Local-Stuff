package handlers

import (
	"net/http"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user management requests
type UserHandler struct {
	userRepo     *repository.UserRepository
	auditService *services.AuditService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo *repository.UserRepository, auditService *services.AuditService) *UserHandler {
	return &UserHandler{
		userRepo:     userRepo,
		auditService: auditService,
	}
}

// List returns all users
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userRepo.FindAll(c.Request.Context(), 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve users"})
		return
	}

	response := make([]models.UserResponse, len(users))
	for i, user := range users {
		response[i] = models.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        user.Role,
			TOTPEnabled: user.TOTPEnabled,
			CreatedAt:   user.CreatedAt,
			LastLogin:   user.LastLogin,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRole updates a user's role
func (h *UserHandler) UpdateRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	// Try query param first, then body
	role := c.Query("role")
	if role == "" {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Role is required"})
			return
		}
		role = req.Role
	}

	if !models.IsValidRole(role) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid role"})
		return
	}

	err := h.userRepo.Update(c.Request.Context(), userID, map[string]interface{}{"role": role})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
		return
	}

	h.auditService.LogFromContext(c, "role_changed", "user", userID, map[string]interface{}{"new_role": role})

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// Delete removes a user
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot delete your own account"})
		return
	}

	err := h.userRepo.Delete(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
		return
	}

	h.auditService.LogFromContext(c, "user_deleted", "user", userID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
