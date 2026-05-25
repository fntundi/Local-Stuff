// Package handlers provides HTTP handlers for system mode operations
package handlers

import (
	"net/http"
	"time"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// SystemModeHandler handles system mode operations
type SystemModeHandler struct {
	settingsRepo   *repository.SettingsRepository
	cameraRepo     *repository.CameraRepository
	eventRepo      *repository.EventRepository
	auditService   *services.AuditService
	webhookService *services.WebhookService
}

// NewSystemModeHandler creates a new SystemModeHandler
func NewSystemModeHandler(
	settingsRepo *repository.SettingsRepository,
	cameraRepo *repository.CameraRepository,
	eventRepo *repository.EventRepository,
	auditService *services.AuditService,
	webhookService *services.WebhookService,
) *SystemModeHandler {
	return &SystemModeHandler{
		settingsRepo:   settingsRepo,
		cameraRepo:     cameraRepo,
		eventRepo:      eventRepo,
		auditService:   auditService,
		webhookService: webhookService,
	}
}

// GetMode returns the current system mode
func (h *SystemModeHandler) GetMode(c *gin.Context) {
	mode, changedAt, changedBy, err := h.settingsRepo.GetMode(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get system mode"})
		return
	}

	c.JSON(http.StatusOK, models.SystemModeResponse{
		Mode:      models.SystemMode(mode),
		ChangedAt: changedAt,
		ChangedBy: changedBy,
	})
}

// SetMode changes the system mode
func (h *SystemModeHandler) SetMode(c *gin.Context) {
	var req models.SystemModeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Validate mode
	if req.Mode != models.ModeHome && req.Mode != models.ModeAway {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid mode. Must be 'home' or 'away'"})
		return
	}

	username, _ := c.Get("username")
	usernameStr, _ := username.(string)
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Update mode
	err := h.settingsRepo.UpdateMode(c.Request.Context(), string(req.Mode), usernameStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to update system mode"})
		return
	}

	// Create mode change event
	event := &repository.Event{
		ID:           services.GenerateUUID(),
		EventType:    "mode_change",
		Severity:     "info",
		Message:      "System mode changed to " + string(req.Mode),
		Details:      map[string]interface{}{"new_mode": req.Mode, "changed_by": usernameStr},
		Acknowledged: true,
		AcknowledgedBy: usernameStr,
		CreatedAt:    time.Now().UTC(),
	}
	_ = h.eventRepo.Create(c.Request.Context(), event)

	// Audit log
	h.auditService.LogFromContext(c, "mode_changed", "system", "mode", map[string]interface{}{
		"new_mode": req.Mode,
	})

	// Send webhook
	go func() {
		_ = h.webhookService.SendModeChangeEvent(c.Request.Context(), string(req.Mode), usernameStr)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "System mode updated successfully",
		"mode":    req.Mode,
		"changed_by": usernameStr,
		"changed_at": time.Now().UTC(),
	})

	// Log for operators
	_ = h.auditService.Log(c.Request.Context(), userIDStr, usernameStr, "system_mode_changed", "system", "", c.ClientIP(), map[string]interface{}{
		"new_mode": req.Mode,
	})
}
