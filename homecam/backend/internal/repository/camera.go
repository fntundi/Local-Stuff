package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Camera represents a camera document in MongoDB
type Camera struct {
	ID                     string     `bson:"id"`
	Name                   string     `bson:"name"`
	IPAddress              string     `bson:"ip_address"`
	Port                   int        `bson:"port"`
	RTSPPort               int        `bson:"rtsp_port"`
	RTSPPath               string     `bson:"rtsp_path"`
	UsernameEncrypted      string     `bson:"username_encrypted"`
	PasswordEncrypted      string     `bson:"password_encrypted"`
	Protocol               string     `bson:"protocol"`
	Manufacturer           string     `bson:"manufacturer,omitempty"`
	Model                  string     `bson:"model,omitempty"`
	Location               string     `bson:"location,omitempty"`
	PTZCapable             bool       `bson:"ptz_capable"`
	MotionDetectionEnabled bool       `bson:"motion_detection_enabled"`
	RecordingEnabled       bool       `bson:"recording_enabled"`
	IsOnline               bool       `bson:"is_online"`
	LastSeen               *time.Time `bson:"last_seen,omitempty"`
	CreatedAt              time.Time  `bson:"created_at"`
	// ONVIF fields
	ONVIFPort              int    `bson:"onvif_port"`
	ONVIFUsernameEncrypted string `bson:"onvif_username_encrypted,omitempty"`
	ONVIFPasswordEncrypted string `bson:"onvif_password_encrypted,omitempty"`
	ONVIFConfigured        bool   `bson:"onvif_configured"`
	// ONVIF Capabilities (detected)
	HasRelayOutputs bool `bson:"has_relay_outputs"`
	HasAudioOutputs bool `bson:"has_audio_outputs"`
	HasAlarmCapability bool `bson:"has_alarm_capability"`
	RelayCount      int  `bson:"relay_count"`
	// Mode override
	ModeOverride string `bson:"mode_override"` // "none", "home", "away"
}

// CameraRepository handles camera data operations
type CameraRepository struct {
	collection *mongo.Collection
}

// NewCameraRepository creates a new CameraRepository
func NewCameraRepository(db *mongo.Database) *CameraRepository {
	return &CameraRepository{
		collection: db.Collection("cameras"),
	}
}

// Create inserts a new camera
func (r *CameraRepository) Create(ctx context.Context, camera *Camera) error {
	_, err := r.collection.InsertOne(ctx, camera)
	return err
}

// FindByID retrieves a camera by ID
func (r *CameraRepository) FindByID(ctx context.Context, id string) (*Camera, error) {
	var camera Camera
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&camera)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &camera, err
}

// FindByIPAddress retrieves a camera by IP address
func (r *CameraRepository) FindByIPAddress(ctx context.Context, ip string) (*Camera, error) {
	var camera Camera
	err := r.collection.FindOne(ctx, bson.M{"ip_address": ip}).Decode(&camera)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &camera, err
}

// FindAll retrieves all cameras (without encrypted credentials)
func (r *CameraRepository) FindAll(ctx context.Context, limit int64) ([]*Camera, error) {
	opts := options.Find().SetLimit(limit)
	// Exclude encrypted credentials
	opts.SetProjection(bson.M{
		"username_encrypted":       0,
		"password_encrypted":       0,
		"onvif_username_encrypted": 0,
		"onvif_password_encrypted": 0,
	})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cameras []*Camera
	if err := cursor.All(ctx, &cameras); err != nil {
		return nil, err
	}
	return cameras, nil
}

// Update updates a camera document
func (r *CameraRepository) Update(ctx context.Context, id string, update bson.M) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// Delete removes a camera by ID
func (r *CameraRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// CountTotal returns the total number of cameras
func (r *CameraRepository) CountTotal(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountOnline returns the number of online cameras
func (r *CameraRepository) CountOnline(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_online": true})
}

// UpdateStatus updates the online status of a camera
func (r *CameraRepository) UpdateStatus(ctx context.Context, id string, isOnline bool) error {
	now := time.Now().UTC()
	return r.Update(ctx, id, bson.M{
		"is_online":  isOnline,
		"last_seen": now,
	})
}

// CountAlarmCapable returns the number of cameras with alarm capability
func (r *CameraRepository) CountAlarmCapable(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"has_alarm_capability": true})
}

// FindAllWithCredentials retrieves a camera with all fields including encrypted credentials
func (r *CameraRepository) FindByIDWithCredentials(ctx context.Context, id string) (*Camera, error) {
	var camera Camera
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&camera)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &camera, err
}
