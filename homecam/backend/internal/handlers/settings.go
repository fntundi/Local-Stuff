package handlers

import (
	"net/http"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct {
	settingsRepo *repository.SettingsRepository
	auditService *services.AuditService
}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler(settingsRepo *repository.SettingsRepository, auditService *services.AuditService) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: settingsRepo,
		auditService: auditService,
	}
}

// Get returns the current system settings
func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.settingsRepo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve settings"})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, models.DefaultSettings())
		return
	}

	c.JSON(http.StatusOK, models.SystemSettings{
		StoragePath:              settings.StoragePath,
		RetentionDays:            settings.RetentionDays,
		MotionSensitivity:        settings.MotionSensitivity,
		AlarmNotificationEnabled: settings.AlarmNotificationEnabled,
		EmailNotifications:       settings.EmailNotifications,
		SMTPServer:               settings.SMTPServer,
		SMTPPort:                 settings.SMTPPort,
	})
}

// Update modifies the system settings
func (h *SettingsHandler) Update(c *gin.Context) {
	var req models.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	settings := &repository.Settings{
		StoragePath:              req.StoragePath,
		RetentionDays:            req.RetentionDays,
		MotionSensitivity:        req.MotionSensitivity,
		AlarmNotificationEnabled: req.AlarmNotificationEnabled,
		EmailNotifications:       req.EmailNotifications,
		SMTPServer:               req.SMTPServer,
		SMTPPort:                 req.SMTPPort,
	}

	if err := h.settingsRepo.Upsert(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to update settings"})
		return
	}

	h.auditService.LogFromContext(c, "settings_updated", "settings", "system_settings", map[string]interface{}{
		"storage_path":    req.StoragePath,
		"retention_days":  req.RetentionDays,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}
