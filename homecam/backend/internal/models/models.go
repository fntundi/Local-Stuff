// Package models defines all data structures used throughout the application
package models

import (
	"time"
)

// =============================================================================
// User Models
// =============================================================================

// UserRole defines available user roles
type UserRole string

const (
	RoleAdmin            UserRole = "admin"
	RoleSecurityOperator UserRole = "security_operator"
	RoleViewer           UserRole = "viewer"
)

// ValidRoles returns all valid user roles
func ValidRoles() []UserRole {
	return []UserRole{RoleAdmin, RoleSecurityOperator, RoleViewer}
}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if string(r) == role {
			return true
		}
	}
	return false
}

// UserRegisterRequest represents user registration input
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	Role     string `json:"role"`
}

// UserLoginRequest represents user login input
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totp_code"`
}

// UserResponse represents user data returned to clients
type UserResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	TOTPEnabled bool       `json:"totp_enabled"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
}

// TokenResponse represents authentication tokens returned after login
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// Requires2FAResponse indicates 2FA is required
type Requires2FAResponse struct {
	Requires2FA bool   `json:"requires_2fa"`
	Message     string `json:"message"`
}

// TOTPSetupResponse contains 2FA setup information
type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
	QRData string `json:"qr_data"`
}

// TOTPVerifyRequest represents 2FA verification input
type TOTPVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// =============================================================================
// System Mode Models
// =============================================================================

// SystemMode defines the system security mode
type SystemMode string

const (
	ModeHome SystemMode = "home"
	ModeAway SystemMode = "away"
)

// CameraModeOverride defines per-camera mode override options
type CameraModeOverride string

const (
	ModeOverrideNone   CameraModeOverride = "none"   // Follow system mode
	ModeOverrideHome   CameraModeOverride = "home"   // Always home (no alarms)
	ModeOverrideAway   CameraModeOverride = "away"   // Always away (trigger alarms)
)

// SystemModeResponse represents the current system mode
type SystemModeResponse struct {
	Mode        SystemMode `json:"mode"`
	ChangedAt   time.Time  `json:"changed_at"`
	ChangedBy   string     `json:"changed_by,omitempty"`
}

// SystemModeUpdateRequest represents a mode change request
type SystemModeUpdateRequest struct {
	Mode SystemMode `json:"mode" binding:"required"`
}

// =============================================================================
// Camera Models
// =============================================================================

// ONVIFCapabilities represents detected ONVIF capabilities
type ONVIFCapabilities struct {
	Supported       bool `json:"supported"`
	HasRelayOutputs bool `json:"has_relay_outputs"`
	HasAudioOutputs bool `json:"has_audio_outputs"`
	HasPTZ          bool `json:"has_ptz"`
	HasAnalytics    bool `json:"has_analytics"`
	RelayCount      int  `json:"relay_count"`
	ProfilesCount   int  `json:"profiles_count"`
}

// CameraCreateRequest represents camera creation input
type CameraCreateRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	IPAddress    string `json:"ip_address" binding:"required"`
	Port         int    `json:"port" binding:"gte=1,lte=65535"`
	RTSPPort     int    `json:"rtsp_port" binding:"gte=1,lte=65535"`
	RTSPPath     string `json:"rtsp_path"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Protocol     string `json:"protocol"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Location     string `json:"location"`
	PTZCapable   bool   `json:"ptz_capable"`
	// ONVIF credentials (separate from streaming credentials)
	ONVIFPort     int    `json:"onvif_port"`
	ONVIFUsername string `json:"onvif_username"`
	ONVIFPassword string `json:"onvif_password"`
	// Mode override
	ModeOverride string `json:"mode_override"`
}

// CameraUpdateRequest represents camera update input
type CameraUpdateRequest struct {
	Name                   *string `json:"name"`
	Location               *string `json:"location"`
	MotionDetectionEnabled *bool   `json:"motion_detection_enabled"`
	RecordingEnabled       *bool   `json:"recording_enabled"`
	// ONVIF credentials
	ONVIFPort     *int    `json:"onvif_port"`
	ONVIFUsername *string `json:"onvif_username"`
	ONVIFPassword *string `json:"onvif_password"`
	// Mode override
	ModeOverride *string `json:"mode_override"`
}

// CameraResponse represents camera data returned to clients
type CameraResponse struct {
	ID                     string             `json:"id"`
	Name                   string             `json:"name"`
	IPAddress              string             `json:"ip_address"`
	Port                   int                `json:"port"`
	RTSPPort               int                `json:"rtsp_port"`
	RTSPPath               string             `json:"rtsp_path"`
	Protocol               string             `json:"protocol"`
	Manufacturer           string             `json:"manufacturer,omitempty"`
	Model                  string             `json:"model,omitempty"`
	Location               string             `json:"location,omitempty"`
	PTZCapable             bool               `json:"ptz_capable"`
	MotionDetectionEnabled bool               `json:"motion_detection_enabled"`
	RecordingEnabled       bool               `json:"recording_enabled"`
	IsOnline               bool               `json:"is_online"`
	LastSeen               *time.Time         `json:"last_seen,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	// ONVIF fields
	ONVIFPort          int                `json:"onvif_port"`
	ONVIFConfigured    bool               `json:"onvif_configured"`
	ONVIFCapabilities  *ONVIFCapabilities `json:"onvif_capabilities,omitempty"`
	HasAlarmCapability bool               `json:"has_alarm_capability"`
	// Mode fields
	ModeOverride   string `json:"mode_override"`
	EffectiveMode  string `json:"effective_mode"` // Computed based on system mode + override
}

// CameraStreamURLResponse contains camera streaming information.
// RTSPURL is the MediaMTX proxy path (no credentials embedded).
// Call POST /cameras/:id/stream/start before accessing RTSPURL or HLSPath.
type CameraStreamURLResponse struct {
	CameraID    string `json:"camera_id"`
	StreamType  string `json:"stream_type"`
	RTSPURL     string `json:"rtsp_url"`
	SnapshotURL string `json:"snapshot_url"`
	HLSPath     string `json:"hls_path"`
}

// =============================================================================
// Event Models
// =============================================================================

// EventType defines available event types
type EventType string

const (
	EventTypeMotion             EventType = "motion"
	EventTypeAlarm              EventType = "alarm"
	EventTypeAlarmTriggered     EventType = "alarm_triggered"
	EventTypeConnectionLost     EventType = "connection_lost"
	EventTypeConnectionRestored EventType = "connection_restored"
	EventTypeModeChange         EventType = "mode_change"
)

// EventSeverity defines event severity levels
type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityCritical EventSeverity = "critical"
)

// EventCreateRequest represents event creation input
type EventCreateRequest struct {
	CameraID  string                 `json:"camera_id" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message" binding:"required"`
	Details   map[string]interface{} `json:"details"`
}

// EventResponse represents event data returned to clients
type EventResponse struct {
	ID             string                 `json:"id"`
	CameraID       string                 `json:"camera_id"`
	CameraName     string                 `json:"camera_name,omitempty"`
	EventType      string                 `json:"event_type"`
	Severity       string                 `json:"severity"`
	Message        string                 `json:"message"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Acknowledged   bool                   `json:"acknowledged"`
	AcknowledgedBy string                 `json:"acknowledged_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// =============================================================================
// Audit Log Models
// =============================================================================

// AuditLogResponse represents audit log data returned to clients
type AuditLogResponse struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	CreatedAt    time.Time              `json:"created_at"`
}

// =============================================================================
// Settings Models
// =============================================================================

// WebhookConfig represents webhook integration configuration
type WebhookConfig struct {
	Enabled       bool              `json:"enabled"`
	URL           string            `json:"url"`
	Secret        string            `json:"secret,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	RetryCount    int               `json:"retry_count"`
	TimeoutSecs   int               `json:"timeout_secs"`
	// Event types to send
	OnMotion      bool `json:"on_motion"`
	OnAlarm       bool `json:"on_alarm"`
	OnModeChange  bool `json:"on_mode_change"`
	OnConnection  bool `json:"on_connection"`
}

// SystemSettings represents system configuration
type SystemSettings struct {
	StoragePath              string        `json:"storage_path"`
	RetentionDays            int           `json:"retention_days"`
	MotionSensitivity        int           `json:"motion_sensitivity"`
	AlarmNotificationEnabled bool          `json:"alarm_notification_enabled"`
	EmailNotifications       bool          `json:"email_notifications"`
	SMTPServer               string        `json:"smtp_server,omitempty"`
	SMTPPort                 int           `json:"smtp_port,omitempty"`
	// System mode
	SystemMode    string    `json:"system_mode"`
	ModeChangedAt time.Time `json:"mode_changed_at,omitempty"`
	ModeChangedBy string    `json:"mode_changed_by,omitempty"`
	// Webhook integration
	Webhook       *WebhookConfig `json:"webhook,omitempty"`
}

// DefaultSettings returns default system settings
func DefaultSettings() *SystemSettings {
	return &SystemSettings{
		StoragePath:              "/recordings",
		RetentionDays:            30,
		MotionSensitivity:        50,
		AlarmNotificationEnabled: true,
		EmailNotifications:       false,
		SMTPPort:                 587,
		SystemMode:               string(ModeHome),
		Webhook: &WebhookConfig{
			Enabled:     false,
			RetryCount:  3,
			TimeoutSecs: 10,
			OnMotion:    true,
			OnAlarm:     true,
			OnModeChange: true,
			OnConnection: false,
		},
	}
}

// =============================================================================
// Dashboard Models
// =============================================================================

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalCameras          int64  `json:"total_cameras"`
	OnlineCameras         int64  `json:"online_cameras"`
	OfflineCameras        int64  `json:"offline_cameras"`
	TotalEvents           int64  `json:"total_events"`
	UnacknowledgedEvents  int64  `json:"unacknowledged_events"`
	CriticalEvents        int64  `json:"critical_events"`
	SystemMode            string `json:"system_mode"`
	AlarmCapableCameras   int64  `json:"alarm_capable_cameras"`
}

// =============================================================================
// Generic Response Models
// =============================================================================

// ErrorResponse represents an error returned to clients
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success message
type SuccessResponse struct {
	Message string `json:"message"`
}
