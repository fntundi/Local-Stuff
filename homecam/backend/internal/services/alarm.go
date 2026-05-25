// Package services provides the alarm orchestration service
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"sentinel-noc/internal/repository"
)

// AlarmService handles alarm triggering logic
type AlarmService struct {
	cameraRepo     *repository.CameraRepository
	eventRepo      *repository.EventRepository
	settingsRepo   *repository.SettingsRepository
	cryptoService  *CryptoService
	onvifService   *ONVIFService
	webhookService *WebhookService
	auditService   *AuditService
}

// NewAlarmService creates a new alarm service
func NewAlarmService(
	cameraRepo *repository.CameraRepository,
	eventRepo *repository.EventRepository,
	settingsRepo *repository.SettingsRepository,
	cryptoService *CryptoService,
	onvifService *ONVIFService,
	webhookService *WebhookService,
	auditService *AuditService,
) *AlarmService {
	return &AlarmService{
		cameraRepo:     cameraRepo,
		eventRepo:      eventRepo,
		settingsRepo:   settingsRepo,
		cryptoService:  cryptoService,
		onvifService:   onvifService,
		webhookService: webhookService,
		auditService:   auditService,
	}
}

// ProcessMotionEvent handles a motion detection event
// This is the main entry point for motion-triggered alarm logic
func (s *AlarmService) ProcessMotionEvent(ctx context.Context, cameraID, message string, details map[string]interface{}) error {
	// Get camera
	camera, err := s.cameraRepo.FindByIDWithCredentials(ctx, cameraID)
	if err != nil || camera == nil {
		return fmt.Errorf("camera not found: %w", err)
	}

	// Determine effective mode for this camera
	effectiveMode, err := s.GetEffectiveMode(ctx, camera)
	if err != nil {
		return err
	}

	log.Printf("Motion detected on camera %s (%s), effective mode: %s", camera.Name, cameraID, effectiveMode)

	// Always create the motion event
	motionEvent := &repository.Event{
		ID:           GenerateUUID(),
		CameraID:     cameraID,
		CameraName:   camera.Name,
		EventType:    "motion",
		Severity:     "warning",
		Message:      message,
		Details:      details,
		Acknowledged: false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.eventRepo.Create(ctx, motionEvent); err != nil {
		log.Printf("Failed to create motion event: %v", err)
	}

	// Send motion webhook
	go func() {
		_ = s.webhookService.SendMotionEvent(context.Background(), cameraID, camera.Name, message, details)
	}()

	// If in AWAY mode, trigger alarm
	if effectiveMode == "away" {
		return s.triggerAlarm(ctx, camera, message, details)
	}

	return nil
}

// triggerAlarm triggers the camera alarm and creates critical event
func (s *AlarmService) triggerAlarm(ctx context.Context, camera *repository.Camera, message string, details map[string]interface{}) error {
	log.Printf("Triggering alarm for camera %s", camera.Name)

	alarmTriggered := false
	alarmError := ""

	// Try to trigger ONVIF alarm if camera has capability
	if camera.HasAlarmCapability && camera.ONVIFConfigured {
		// Decrypt ONVIF credentials
		username, err := s.cryptoService.Decrypt(camera.ONVIFUsernameEncrypted)
		if err != nil {
			alarmError = "failed to decrypt ONVIF credentials"
		} else {
			password, err := s.cryptoService.Decrypt(camera.ONVIFPasswordEncrypted)
			if err != nil {
				alarmError = "failed to decrypt ONVIF credentials"
			} else {
				// Trigger alarm for 30 seconds
				err = s.onvifService.TriggerAlarm(ctx, camera.IPAddress, camera.ONVIFPort, username, password, "", 30)
				if err != nil {
					alarmError = fmt.Sprintf("ONVIF alarm trigger failed: %v", err)
					log.Printf("Failed to trigger ONVIF alarm: %v", err)
				} else {
					alarmTriggered = true
					log.Printf("ONVIF alarm triggered successfully for camera %s", camera.Name)
				}
			}
		}
	} else if !camera.HasAlarmCapability {
		alarmError = "camera does not have alarm capability"
	} else if !camera.ONVIFConfigured {
		alarmError = "ONVIF not configured for this camera"
	}

	// Create critical alarm event
	eventDetails := map[string]interface{}{
		"original_message":   message,
		"alarm_triggered":    alarmTriggered,
		"camera_has_alarm":   camera.HasAlarmCapability,
		"onvif_configured":   camera.ONVIFConfigured,
	}
	if alarmError != "" {
		eventDetails["alarm_error"] = alarmError
	}
	if details != nil {
		for k, v := range details {
			eventDetails[k] = v
		}
	}

	alarmEvent := &repository.Event{
		ID:           GenerateUUID(),
		CameraID:     camera.ID,
		CameraName:   camera.Name,
		EventType:    "alarm_triggered",
		Severity:     "critical",
		Message:      fmt.Sprintf("ALARM: %s (Mode: Away)", message),
		Details:      eventDetails,
		Acknowledged: false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.eventRepo.Create(ctx, alarmEvent); err != nil {
		log.Printf("Failed to create alarm event: %v", err)
	}

	// Send alarm webhook
	go func() {
		_ = s.webhookService.SendAlarmEvent(context.Background(), camera.ID, camera.Name, alarmEvent.Message, eventDetails)
	}()

	return nil
}

// GetEffectiveMode determines the effective security mode for a camera
func (s *AlarmService) GetEffectiveMode(ctx context.Context, camera *repository.Camera) (string, error) {
	// Check camera override first
	if camera.ModeOverride == "home" {
		return "home", nil
	}
	if camera.ModeOverride == "away" {
		return "away", nil
	}

	// Fall back to system mode
	systemMode, _, _, err := s.settingsRepo.GetMode(ctx)
	if err != nil {
		return "home", err // Default to home (safe) on error
	}

	return systemMode, nil
}

// ManualTriggerAlarm allows manual alarm triggering
func (s *AlarmService) ManualTriggerAlarm(ctx context.Context, cameraID, userID, username string) error {
	camera, err := s.cameraRepo.FindByIDWithCredentials(ctx, cameraID)
	if err != nil || camera == nil {
		return fmt.Errorf("camera not found")
	}

	if !camera.HasAlarmCapability {
		return fmt.Errorf("camera does not have alarm capability")
	}

	if !camera.ONVIFConfigured {
		return fmt.Errorf("ONVIF not configured for this camera")
	}

	// Decrypt credentials
	onvifUsername, err := s.cryptoService.Decrypt(camera.ONVIFUsernameEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}
	onvifPassword, err := s.cryptoService.Decrypt(camera.ONVIFPasswordEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}

	// Trigger alarm
	err = s.onvifService.TriggerAlarm(ctx, camera.IPAddress, camera.ONVIFPort, onvifUsername, onvifPassword, "", 10)
	if err != nil {
		return fmt.Errorf("failed to trigger alarm: %w", err)
	}

	// Audit log
	_ = s.auditService.Log(ctx, userID, username, "manual_alarm_trigger", "camera", cameraID, "", map[string]interface{}{
		"camera_name": camera.Name,
	})

	return nil
}

// StopAlarm manually stops an alarm
func (s *AlarmService) StopAlarm(ctx context.Context, cameraID string) error {
	camera, err := s.cameraRepo.FindByIDWithCredentials(ctx, cameraID)
	if err != nil || camera == nil {
		return fmt.Errorf("camera not found")
	}

	if !camera.ONVIFConfigured {
		return fmt.Errorf("ONVIF not configured")
	}

	// Decrypt credentials
	onvifUsername, err := s.cryptoService.Decrypt(camera.ONVIFUsernameEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}
	onvifPassword, err := s.cryptoService.Decrypt(camera.ONVIFPasswordEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}

	return s.onvifService.DeactivateAlarm(ctx, camera.IPAddress, camera.ONVIFPort, onvifUsername, onvifPassword, "")
}
