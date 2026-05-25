package handlers

import (
	"net/http"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	cameraRepo   *repository.CameraRepository
	eventRepo    *repository.EventRepository
	settingsRepo *repository.SettingsRepository
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(cameraRepo *repository.CameraRepository, eventRepo *repository.EventRepository, settingsRepo *repository.SettingsRepository) *DashboardHandler {
	return &DashboardHandler{
		cameraRepo:   cameraRepo,
		eventRepo:    eventRepo,
		settingsRepo: settingsRepo,
	}
}

// GetStats returns dashboard statistics
func (h *DashboardHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalCameras, err := h.cameraRepo.CountTotal(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get camera count"})
		return
	}

	onlineCameras, err := h.cameraRepo.CountOnline(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get online camera count"})
		return
	}

	totalEvents, err := h.eventRepo.CountTotal(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get event count"})
		return
	}

	unacknowledgedEvents, err := h.eventRepo.CountUnacknowledged(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get unacknowledged event count"})
		return
	}

	criticalEvents, err := h.eventRepo.CountCriticalUnacknowledged(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to get critical event count"})
		return
	}

	// Get system mode
	systemMode, _, _, err := h.settingsRepo.GetMode(ctx)
	if err != nil {
		systemMode = "home"
	}

	// Get alarm capable camera count
	alarmCapable, err := h.cameraRepo.CountAlarmCapable(ctx)
	if err != nil {
		alarmCapable = 0
	}

	c.JSON(http.StatusOK, models.DashboardStats{
		TotalCameras:         totalCameras,
		OnlineCameras:        onlineCameras,
		OfflineCameras:       totalCameras - onlineCameras,
		TotalEvents:          totalEvents,
		UnacknowledgedEvents: unacknowledgedEvents,
		CriticalEvents:       criticalEvents,
		SystemMode:           systemMode,
		AlarmCapableCameras:  alarmCapable,
	})
}
