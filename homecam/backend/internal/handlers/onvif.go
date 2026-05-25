// Package handlers provides HTTP handlers for ONVIF operations
package handlers

import (
	"net/http"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ONVIFHandler handles ONVIF-related operations
type ONVIFHandler struct {
	cameraRepo    *repository.CameraRepository
	cryptoService *services.CryptoService
	onvifService  *services.ONVIFService
	alarmService  *services.AlarmService
	auditService  *services.AuditService
}

// NewONVIFHandler creates a new ONVIFHandler
func NewONVIFHandler(
	cameraRepo *repository.CameraRepository,
	cryptoService *services.CryptoService,
	onvifService *services.ONVIFService,
	alarmService *services.AlarmService,
	auditService *services.AuditService,
) *ONVIFHandler {
	return &ONVIFHandler{
		cameraRepo:    cameraRepo,
		cryptoService: cryptoService,
		onvifService:  onvifService,
		alarmService:  alarmService,
		auditService:  auditService,
	}
}

// DetectCapabilities probes a camera for ONVIF capabilities
func (h *ONVIFHandler) DetectCapabilities(c *gin.Context) {
	cameraID := c.Param("id")

	camera, err := h.cameraRepo.FindByIDWithCredentials(c.Request.Context(), cameraID)
	if err != nil || camera == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	// Check if ONVIF is configured
	if camera.ONVIFUsernameEncrypted == "" || camera.ONVIFPasswordEncrypted == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "ONVIF credentials not configured for this camera"})
		return
	}

	// Decrypt ONVIF credentials
	username, err := h.cryptoService.Decrypt(camera.ONVIFUsernameEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt ONVIF credentials"})
		return
	}
	password, err := h.cryptoService.Decrypt(camera.ONVIFPasswordEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt ONVIF credentials"})
		return
	}

	// Use ONVIF port or default to 80
	port := camera.ONVIFPort
	if port == 0 {
		port = 80
	}

	// Detect capabilities
	caps, err := h.onvifService.DetectCapabilities(c.Request.Context(), camera.IPAddress, port, username, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail":  "Failed to detect ONVIF capabilities",
			"error":   err.Error(),
			"message": "Ensure ONVIF is enabled on the camera and credentials are correct",
		})
		return
	}

	// Update camera with detected capabilities
	hasAlarm := caps.HasRelayOutputs || caps.HasAudioOutputs
	updateData := bson.M{
		"has_relay_outputs":    caps.HasRelayOutputs,
		"has_audio_outputs":    caps.HasAudioOutputs,
		"has_alarm_capability": hasAlarm,
		"relay_count":          caps.RelayCount,
		"ptz_capable":          caps.HasPTZ,
	}

	if err := h.cameraRepo.Update(c.Request.Context(), cameraID, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to update camera capabilities"})
		return
	}

	h.auditService.LogFromContext(c, "onvif_capabilities_detected", "camera", cameraID, map[string]interface{}{
		"capabilities": caps,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":      "Capabilities detected successfully",
		"capabilities": caps,
		"has_alarm":    hasAlarm,
	})
}

// TestConnection tests ONVIF connectivity
func (h *ONVIFHandler) TestConnection(c *gin.Context) {
	cameraID := c.Param("id")

	camera, err := h.cameraRepo.FindByIDWithCredentials(c.Request.Context(), cameraID)
	if err != nil || camera == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	if camera.ONVIFUsernameEncrypted == "" || camera.ONVIFPasswordEncrypted == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "ONVIF credentials not configured"})
		return
	}

	username, err := h.cryptoService.Decrypt(camera.ONVIFUsernameEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt credentials"})
		return
	}
	password, err := h.cryptoService.Decrypt(camera.ONVIFPasswordEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt credentials"})
		return
	}

	port := camera.ONVIFPort
	if port == 0 {
		port = 80
	}

	success, info, err := h.onvifService.TestConnection(c.Request.Context(), camera.IPAddress, port, username, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     success,
		"device_info": info,
	})
}

// TriggerAlarm manually triggers the camera alarm
func (h *ONVIFHandler) TriggerAlarm(c *gin.Context) {
	cameraID := c.Param("id")
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	err := h.alarmService.ManualTriggerAlarm(c.Request.Context(), cameraID, userID.(string), username.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alarm triggered successfully"})
}

// StopAlarm manually stops the camera alarm
func (h *ONVIFHandler) StopAlarm(c *gin.Context) {
	cameraID := c.Param("id")

	err := h.alarmService.StopAlarm(c.Request.Context(), cameraID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	h.auditService.LogFromContext(c, "alarm_stopped", "camera", cameraID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Alarm stopped successfully"})
}

// UpdateONVIFCredentials updates ONVIF credentials for a camera
func (h *ONVIFHandler) UpdateONVIFCredentials(c *gin.Context) {
	cameraID := c.Param("id")

	var req struct {
		ONVIFPort     int    `json:"onvif_port"`
		ONVIFUsername string `json:"onvif_username" binding:"required"`
		ONVIFPassword string `json:"onvif_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Encrypt credentials
	usernameEncrypted, err := h.cryptoService.Encrypt(req.ONVIFUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt credentials"})
		return
	}
	passwordEncrypted, err := h.cryptoService.Encrypt(req.ONVIFPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt credentials"})
		return
	}

	port := req.ONVIFPort
	if port == 0 {
		port = 80
	}

	updateData := bson.M{
		"onvif_port":               port,
		"onvif_username_encrypted": usernameEncrypted,
		"onvif_password_encrypted": passwordEncrypted,
		"onvif_configured":         true,
	}

	if err := h.cameraRepo.Update(c.Request.Context(), cameraID, updateData); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	h.auditService.LogFromContext(c, "onvif_credentials_updated", "camera", cameraID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "ONVIF credentials updated successfully"})
}

// GetCameraWithEffectiveMode returns camera with computed effective mode
func GetCameraResponseWithMode(camera *repository.Camera, systemMode string) models.CameraResponse {
	effectiveMode := systemMode
	if camera.ModeOverride == "home" {
		effectiveMode = "home"
	} else if camera.ModeOverride == "away" {
		effectiveMode = "away"
	}

	var caps *models.ONVIFCapabilities
	if camera.ONVIFConfigured {
		caps = &models.ONVIFCapabilities{
			Supported:       true,
			HasRelayOutputs: camera.HasRelayOutputs,
			HasAudioOutputs: camera.HasAudioOutputs,
			RelayCount:      camera.RelayCount,
		}
	}

	return models.CameraResponse{
		ID:                     camera.ID,
		Name:                   camera.Name,
		IPAddress:              camera.IPAddress,
		Port:                   camera.Port,
		RTSPPort:               camera.RTSPPort,
		RTSPPath:               camera.RTSPPath,
		Protocol:               camera.Protocol,
		Manufacturer:           camera.Manufacturer,
		Model:                  camera.Model,
		Location:               camera.Location,
		PTZCapable:             camera.PTZCapable,
		MotionDetectionEnabled: camera.MotionDetectionEnabled,
		RecordingEnabled:       camera.RecordingEnabled,
		IsOnline:               camera.IsOnline,
		LastSeen:               camera.LastSeen,
		CreatedAt:              camera.CreatedAt,
		ONVIFPort:              camera.ONVIFPort,
		ONVIFConfigured:        camera.ONVIFConfigured,
		ONVIFCapabilities:      caps,
		HasAlarmCapability:     camera.HasAlarmCapability,
		ModeOverride:           camera.ModeOverride,
		EffectiveMode:          effectiveMode,
	}
}
