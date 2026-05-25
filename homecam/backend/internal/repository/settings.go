package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Settings represents system settings document in MongoDB
type Settings struct {
	ID                       string    `bson:"id"`
	StoragePath              string    `bson:"storage_path"`
	RetentionDays            int       `bson:"retention_days"`
	MotionSensitivity        int       `bson:"motion_sensitivity"`
	AlarmNotificationEnabled bool      `bson:"alarm_notification_enabled"`
	EmailNotifications       bool      `bson:"email_notifications"`
	SMTPServer               string    `bson:"smtp_server,omitempty"`
	SMTPPort                 int       `bson:"smtp_port,omitempty"`
	UpdatedAt                time.Time `bson:"updated_at"`
	// System mode
	SystemMode    string    `bson:"system_mode"`
	ModeChangedAt time.Time `bson:"mode_changed_at,omitempty"`
	ModeChangedBy string    `bson:"mode_changed_by,omitempty"`
	// Webhook configuration
	WebhookEnabled      bool              `bson:"webhook_enabled"`
	WebhookURL          string            `bson:"webhook_url,omitempty"`
	WebhookSecret       string            `bson:"webhook_secret,omitempty"`
	WebhookHeaders      map[string]string `bson:"webhook_headers,omitempty"`
	WebhookRetryCount   int               `bson:"webhook_retry_count"`
	WebhookTimeoutSecs  int               `bson:"webhook_timeout_secs"`
	WebhookOnMotion     bool              `bson:"webhook_on_motion"`
	WebhookOnAlarm      bool              `bson:"webhook_on_alarm"`
	WebhookOnModeChange bool              `bson:"webhook_on_mode_change"`
	WebhookOnConnection bool              `bson:"webhook_on_connection"`
}

// SettingsRepository handles settings data operations
type SettingsRepository struct {
	collection *mongo.Collection
}

// NewSettingsRepository creates a new SettingsRepository
func NewSettingsRepository(db *mongo.Database) *SettingsRepository {
	return &SettingsRepository{
		collection: db.Collection("settings"),
	}
}

const systemSettingsID = "system_settings"

// Get retrieves the system settings
func (r *SettingsRepository) Get(ctx context.Context) (*Settings, error) {
	var settings Settings
	err := r.collection.FindOne(ctx, bson.M{"id": systemSettingsID}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &settings, err
}

// Upsert creates or updates the system settings
func (r *SettingsRepository) Upsert(ctx context.Context, settings *Settings) error {
	settings.ID = systemSettingsID
	settings.UpdatedAt = time.Now().UTC()

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": systemSettingsID},
		bson.M{"$set": settings},
		opts,
	)
	return err
}

// UpdateMode updates only the system mode
func (r *SettingsRepository) UpdateMode(ctx context.Context, mode, changedBy string) error {
	now := time.Now().UTC()
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": systemSettingsID},
		bson.M{"$set": bson.M{
			"system_mode":     mode,
			"mode_changed_at": now,
			"mode_changed_by": changedBy,
			"updated_at":      now,
		}},
		opts,
	)
	return err
}

// GetMode retrieves just the system mode
func (r *SettingsRepository) GetMode(ctx context.Context) (string, time.Time, string, error) {
	settings, err := r.Get(ctx)
	if err != nil {
		return "home", time.Time{}, "", err
	}
	if settings == nil || settings.SystemMode == "" {
		return "home", time.Time{}, "", nil
	}
	return settings.SystemMode, settings.ModeChangedAt, settings.ModeChangedBy, nil
}
