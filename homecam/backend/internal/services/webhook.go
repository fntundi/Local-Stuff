// Package services provides the webhook integration service
package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sentinel-noc/internal/models"
	"sentinel-noc/internal/repository"
)

// WebhookService handles webhook notifications
type WebhookService struct {
	settingsRepo *repository.SettingsRepository
	httpClient   *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(settingsRepo *repository.SettingsRepository) *WebhookService {
	return &WebhookService{
		settingsRepo: settingsRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookPayload represents the data sent to webhooks
type WebhookPayload struct {
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	CameraID    string                 `json:"camera_id,omitempty"`
	CameraName  string                 `json:"camera_name,omitempty"`
	SystemMode  string                 `json:"system_mode,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SendMotionEvent sends a motion detection webhook
func (s *WebhookService) SendMotionEvent(ctx context.Context, cameraID, cameraName, message string, details map[string]interface{}) error {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil || settings == nil {
		return nil // Silently skip if no settings
	}

	if !settings.WebhookEnabled || !settings.WebhookOnMotion || settings.WebhookURL == "" {
		return nil
	}

	payload := WebhookPayload{
		EventType:  "motion",
		Timestamp:  time.Now().UTC(),
		CameraID:   cameraID,
		CameraName: cameraName,
		SystemMode: settings.SystemMode,
		Severity:   "warning",
		Message:    message,
		Details:    details,
	}

	return s.sendWebhook(ctx, settings, payload)
}

// SendAlarmEvent sends an alarm triggered webhook
func (s *WebhookService) SendAlarmEvent(ctx context.Context, cameraID, cameraName, message string, details map[string]interface{}) error {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil || settings == nil {
		return nil
	}

	if !settings.WebhookEnabled || !settings.WebhookOnAlarm || settings.WebhookURL == "" {
		return nil
	}

	payload := WebhookPayload{
		EventType:  "alarm_triggered",
		Timestamp:  time.Now().UTC(),
		CameraID:   cameraID,
		CameraName: cameraName,
		SystemMode: settings.SystemMode,
		Severity:   "critical",
		Message:    message,
		Details:    details,
	}

	return s.sendWebhook(ctx, settings, payload)
}

// SendModeChangeEvent sends a system mode change webhook
func (s *WebhookService) SendModeChangeEvent(ctx context.Context, newMode, changedBy string) error {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil || settings == nil {
		return nil
	}

	if !settings.WebhookEnabled || !settings.WebhookOnModeChange || settings.WebhookURL == "" {
		return nil
	}

	payload := WebhookPayload{
		EventType:  "mode_change",
		Timestamp:  time.Now().UTC(),
		SystemMode: newMode,
		Message:    fmt.Sprintf("System mode changed to %s by %s", newMode, changedBy),
		Details: map[string]interface{}{
			"new_mode":   newMode,
			"changed_by": changedBy,
		},
	}

	return s.sendWebhook(ctx, settings, payload)
}

// SendConnectionEvent sends a camera connection status webhook
func (s *WebhookService) SendConnectionEvent(ctx context.Context, cameraID, cameraName string, isOnline bool) error {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil || settings == nil {
		return nil
	}

	if !settings.WebhookEnabled || !settings.WebhookOnConnection || settings.WebhookURL == "" {
		return nil
	}

	status := "offline"
	eventType := "connection_lost"
	if isOnline {
		status = "online"
		eventType = "connection_restored"
	}

	payload := WebhookPayload{
		EventType:  eventType,
		Timestamp:  time.Now().UTC(),
		CameraID:   cameraID,
		CameraName: cameraName,
		Message:    fmt.Sprintf("Camera %s is now %s", cameraName, status),
		Details: map[string]interface{}{
			"status": status,
		},
	}

	return s.sendWebhook(ctx, settings, payload)
}

// sendWebhook sends the webhook with retry logic
func (s *WebhookService) sendWebhook(ctx context.Context, settings *repository.Settings, payload WebhookPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	retryCount := settings.WebhookRetryCount
	if retryCount <= 0 {
		retryCount = 3
	}

	timeout := settings.WebhookTimeoutSecs
	if timeout <= 0 {
		timeout = 10
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", settings.WebhookURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Sentinel-NOC/1.0")
		req.Header.Set("X-Sentinel-Event", payload.EventType)

		// Add signature if secret is configured
		if settings.WebhookSecret != "" {
			signature := computeHMAC(jsonPayload, settings.WebhookSecret)
			req.Header.Set("X-Sentinel-Signature", signature)
		}

		// Add custom headers
		for key, value := range settings.WebhookHeaders {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("Webhook attempt %d failed: %v", attempt+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook sent successfully: %s", payload.EventType)
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		log.Printf("Webhook attempt %d failed: status %d", attempt+1, resp.StatusCode)
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", retryCount, lastErr)
}

// computeHMAC generates HMAC-SHA256 signature
func computeHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// ValidateWebhookConfig validates webhook configuration
func ValidateWebhookConfig(config *models.WebhookConfig) error {
	if config == nil {
		return nil
	}

	if config.Enabled && config.URL == "" {
		return fmt.Errorf("webhook URL is required when enabled")
	}

	if config.RetryCount < 0 || config.RetryCount > 10 {
		return fmt.Errorf("retry count must be between 0 and 10")
	}

	if config.TimeoutSecs < 1 || config.TimeoutSecs > 60 {
		return fmt.Errorf("timeout must be between 1 and 60 seconds")
	}

	return nil
}
