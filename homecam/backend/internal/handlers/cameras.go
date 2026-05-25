package handlers

import (
	"fmt"
	"net/http"
	"time"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// CameraHandler handles camera-related requests
type CameraHandler struct {
	cameraRepo    *repository.CameraRepository
	cryptoService *services.CryptoService
	auditService  *services.AuditService
}

// NewCameraHandler creates a new CameraHandler
func NewCameraHandler(
	cameraRepo *repository.CameraRepository,
	cryptoService *services.CryptoService,
	auditService *services.AuditService,
) *CameraHandler {
	return &CameraHandler{
		cameraRepo:    cameraRepo,
		cryptoService: cryptoService,
		auditService:  auditService,
	}
}

// Create handles camera creation
func (h *CameraHandler) Create(c *gin.Context) {
	var req models.CameraCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Validate IP address
	if !services.ValidateIPAddress(req.IPAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid IP address format"})
		return
	}

	// Check for existing camera with same IP
	existing, err := h.cameraRepo.FindByIPAddress(c.Request.Context(), req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to check existing camera"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Camera with this IP already exists"})
		return
	}

	// Encrypt credentials
	usernameEncrypted, err := h.cryptoService.Encrypt(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt credentials"})
		return
	}
	passwordEncrypted, err := h.cryptoService.Encrypt(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt credentials"})
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 80
	}
	if req.RTSPPort == 0 {
		req.RTSPPort = 554
	}
	if req.RTSPPath == "" {
		req.RTSPPath = "/stream1"
	}
	if req.Protocol == "" {
		req.Protocol = "http"
	}

	camera := &repository.Camera{
		ID:                     services.GenerateUUID(),
		Name:                   req.Name,
		IPAddress:              req.IPAddress,
		Port:                   req.Port,
		RTSPPort:               req.RTSPPort,
		RTSPPath:               req.RTSPPath,
		UsernameEncrypted:      usernameEncrypted,
		PasswordEncrypted:      passwordEncrypted,
		Protocol:               req.Protocol,
		Manufacturer:           req.Manufacturer,
		Model:                  req.Model,
		Location:               req.Location,
		PTZCapable:             req.PTZCapable,
		MotionDetectionEnabled: false,
		RecordingEnabled:       false,
		IsOnline:               false,
		CreatedAt:              time.Now().UTC(),
		// ONVIF defaults
		ONVIFPort:       req.ONVIFPort,
		ONVIFConfigured: false,
		ModeOverride:    "none",
	}

	// Set ONVIF port default
	if camera.ONVIFPort == 0 {
		camera.ONVIFPort = 80
	}

	// Encrypt ONVIF credentials if provided
	if req.ONVIFUsername != "" && req.ONVIFPassword != "" {
		onvifUsernameEnc, err := h.cryptoService.Encrypt(req.ONVIFUsername)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt ONVIF credentials"})
			return
		}
		onvifPasswordEnc, err := h.cryptoService.Encrypt(req.ONVIFPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to encrypt ONVIF credentials"})
			return
		}
		camera.ONVIFUsernameEncrypted = onvifUsernameEnc
		camera.ONVIFPasswordEncrypted = onvifPasswordEnc
		camera.ONVIFConfigured = true
	}

	// Set mode override if provided
	if req.ModeOverride != "" {
		if req.ModeOverride == "none" || req.ModeOverride == "home" || req.ModeOverride == "away" {
			camera.ModeOverride = req.ModeOverride
		}
	}

	if err := h.cameraRepo.Create(c.Request.Context(), camera); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create camera"})
		return
	}

	h.auditService.LogFromContext(c, "camera_created", "camera", camera.ID, map[string]interface{}{
		"name": camera.Name,
		"ip":   camera.IPAddress,
	})

	c.JSON(http.StatusOK, models.CameraResponse{
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
		HasAlarmCapability:     camera.HasAlarmCapability,
		ModeOverride:           camera.ModeOverride,
		EffectiveMode:          "home", // Default system mode
	})
}

// List returns all cameras
func (h *CameraHandler) List(c *gin.Context) {
	cameras, err := h.cameraRepo.FindAll(c.Request.Context(), 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve cameras"})
		return
	}

	response := make([]models.CameraResponse, len(cameras))
	for i, camera := range cameras {
		response[i] = models.CameraResponse{
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
			HasAlarmCapability:     camera.HasAlarmCapability,
			ModeOverride:           camera.ModeOverride,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Get returns a specific camera
func (h *CameraHandler) Get(c *gin.Context) {
	cameraID := c.Param("id")

	camera, err := h.cameraRepo.FindByID(c.Request.Context(), cameraID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve camera"})
		return
	}
	if camera == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	c.JSON(http.StatusOK, models.CameraResponse{
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
	})
}

// Update modifies a camera
func (h *CameraHandler) Update(c *gin.Context) {
	cameraID := c.Param("id")

	var req models.CameraUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Location != nil {
		updateData["location"] = *req.Location
	}
	if req.MotionDetectionEnabled != nil {
		updateData["motion_detection_enabled"] = *req.MotionDetectionEnabled
	}
	if req.RecordingEnabled != nil {
		updateData["recording_enabled"] = *req.RecordingEnabled
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "No update data provided"})
		return
	}

	err := h.cameraRepo.Update(c.Request.Context(), cameraID, updateData)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	h.auditService.LogFromContext(c, "camera_updated", "camera", cameraID, updateData)

	c.JSON(http.StatusOK, gin.H{"message": "Camera updated successfully"})
}

// Delete removes a camera
func (h *CameraHandler) Delete(c *gin.Context) {
	cameraID := c.Param("id")

	err := h.cameraRepo.Delete(c.Request.Context(), cameraID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	h.auditService.LogFromContext(c, "camera_deleted", "camera", cameraID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Camera deleted successfully"})
}

// GetStreamURL returns stream access URLs for a camera.
// Credentials are never included in the response; the backend holds them
// server-side and injects them when configuring MediaMTX via StartStream.
// For live playback, call POST /cameras/:id/stream/start first to obtain the HLS URL.
func (h *CameraHandler) GetStreamURL(c *gin.Context) {
	cameraID := c.Param("id")

	camera, err := h.cameraRepo.FindByID(c.Request.Context(), cameraID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve camera"})
		return
	}
	if camera == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	streamPath := fmt.Sprintf("cam-%s", cameraID)
	c.JSON(http.StatusOK, models.CameraStreamURLResponse{
		CameraID:    cameraID,
		StreamType:  "hls",
		RTSPURL:     fmt.Sprintf("rtsp://localhost:8554/%s", streamPath),
		SnapshotURL: fmt.Sprintf("http://%s:%d/snapshot.jpg", camera.IPAddress, camera.Port),
		HLSPath:     fmt.Sprintf("/hls/%s/index.m3u8", streamPath),
	})
}

// UpdateStatus updates a camera's online status
func (h *CameraHandler) UpdateStatus(c *gin.Context) {
	cameraID := c.Param("id")

	var req struct {
		IsOnline bool `json:"is_online"`
	}
	// Try query param first
	if isOnlineStr := c.Query("is_online"); isOnlineStr != "" {
		req.IsOnline = isOnlineStr == "true"
	} else if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "is_online field required"})
		return
	}

	err := h.cameraRepo.UpdateStatus(c.Request.Context(), cameraID, req.IsOnline)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}
