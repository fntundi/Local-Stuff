package services

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"

	"github.com/gin-gonic/gin"
)

// AuditService handles audit logging operations
type AuditService struct {
	repo *repository.AuditRepository
}

// NewAuditService creates a new AuditService
func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log creates an audit log entry
func (s *AuditService) Log(ctx context.Context, userID, username, action, resourceType, resourceID, ipAddress string, details map[string]interface{}) error {
	log := &repository.AuditLog{
		ID:           GenerateUUID(),
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now().UTC(),
	}
	return s.repo.Create(ctx, log)
}

// LogFromContext creates an audit log entry using context values
func (s *AuditService) LogFromContext(c *gin.Context, action, resourceType, resourceID string, details map[string]interface{}) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	userIDStr, _ := userID.(string)
	usernameStr, _ := username.(string)

	_ = s.Log(c.Request.Context(), userIDStr, usernameStr, action, resourceType, resourceID, c.ClientIP(), details)
}

// ListHandler is an HTTP handler for listing audit logs
func (s *AuditService) ListHandler(c *gin.Context) {
	filter := repository.AuditFilter{
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		UserID:       c.Query("user_id"),
		Limit:        100,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			filter.Limit = limit
		}
	}

	logs, err := s.repo.FindAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve audit logs"})
		return
	}

	// Convert to response format
	response := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = models.AuditLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			CreatedAt:    log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}
