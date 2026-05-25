package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// EventHandler handles event-related requests
type EventHandler struct {
	eventRepo    *repository.EventRepository
	cameraRepo   *repository.CameraRepository
	auditService *services.AuditService
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(
	eventRepo *repository.EventRepository,
	cameraRepo *repository.CameraRepository,
	auditService *services.AuditService,
) *EventHandler {
	return &EventHandler{
		eventRepo:    eventRepo,
		cameraRepo:   cameraRepo,
		auditService: auditService,
	}
}

// Create handles event creation
func (h *EventHandler) Create(c *gin.Context) {
	var req models.EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Get camera name
	var cameraName string
	camera, err := h.cameraRepo.FindByID(c.Request.Context(), req.CameraID)
	if err == nil && camera != nil {
		cameraName = camera.Name
	} else {
		cameraName = "Unknown"
	}

	// Set default severity
	severity := req.Severity
	if severity == "" {
		severity = string(models.SeverityInfo)
	}

	event := &repository.Event{
		ID:           services.GenerateUUID(),
		CameraID:     req.CameraID,
		CameraName:   cameraName,
		EventType:    req.EventType,
		Severity:     severity,
		Message:      req.Message,
		Details:      req.Details,
		Acknowledged: false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.eventRepo.Create(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create event"})
		return
	}

	c.JSON(http.StatusOK, models.EventResponse{
		ID:           event.ID,
		CameraID:     event.CameraID,
		CameraName:   event.CameraName,
		EventType:    event.EventType,
		Severity:     event.Severity,
		Message:      event.Message,
		Details:      event.Details,
		Acknowledged: event.Acknowledged,
		CreatedAt:    event.CreatedAt,
	})
}

// List returns events with optional filters
func (h *EventHandler) List(c *gin.Context) {
	filter := repository.EventFilter{
		CameraID:  c.Query("camera_id"),
		EventType: c.Query("event_type"),
		Severity:  c.Query("severity"),
		Limit:     100,
	}

	// Parse acknowledged filter
	if ackStr := c.Query("acknowledged"); ackStr != "" {
		ack := ackStr == "true"
		filter.Acknowledged = &ack
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			filter.Limit = limit
		}
	}

	events, err := h.eventRepo.FindAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to retrieve events"})
		return
	}

	response := make([]models.EventResponse, len(events))
	for i, event := range events {
		response[i] = models.EventResponse{
			ID:             event.ID,
			CameraID:       event.CameraID,
			CameraName:     event.CameraName,
			EventType:      event.EventType,
			Severity:       event.Severity,
			Message:        event.Message,
			Details:        event.Details,
			Acknowledged:   event.Acknowledged,
			AcknowledgedBy: event.AcknowledgedBy,
			CreatedAt:      event.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Acknowledge marks an event as acknowledged
func (h *EventHandler) Acknowledge(c *gin.Context) {
	eventID := c.Param("id")
	username, _ := c.Get("username")
	usernameStr, _ := username.(string)

	err := h.eventRepo.Acknowledge(c.Request.Context(), eventID, usernameStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Event not found"})
		return
	}

	h.auditService.LogFromContext(c, "event_acknowledged", "event", eventID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Event acknowledged"})
}
