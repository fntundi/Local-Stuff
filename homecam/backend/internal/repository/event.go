package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Event represents an event document in MongoDB
type Event struct {
	ID             string                 `bson:"id"`
	CameraID       string                 `bson:"camera_id"`
	CameraName     string                 `bson:"camera_name,omitempty"`
	EventType      string                 `bson:"event_type"`
	Severity       string                 `bson:"severity"`
	Message        string                 `bson:"message"`
	Details        map[string]interface{} `bson:"details,omitempty"`
	Acknowledged   bool                   `bson:"acknowledged"`
	AcknowledgedBy string                 `bson:"acknowledged_by,omitempty"`
	CreatedAt      time.Time              `bson:"created_at"`
}

// EventRepository handles event data operations
type EventRepository struct {
	collection *mongo.Collection
}

// NewEventRepository creates a new EventRepository
func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
	}
}

// Create inserts a new event
func (r *EventRepository) Create(ctx context.Context, event *Event) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

// FindByID retrieves an event by ID
func (r *EventRepository) FindByID(ctx context.Context, id string) (*Event, error) {
	var event Event
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&event)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &event, err
}

// EventFilter represents filter options for listing events
type EventFilter struct {
	CameraID     string
	EventType    string
	Severity     string
	Acknowledged *bool
	Limit        int64
}

// FindAll retrieves events with optional filters
func (r *EventRepository) FindAll(ctx context.Context, filter EventFilter) ([]*Event, error) {
	query := bson.M{}

	if filter.CameraID != "" {
		query["camera_id"] = filter.CameraID
	}
	if filter.EventType != "" {
		query["event_type"] = filter.EventType
	}
	if filter.Severity != "" {
		query["severity"] = filter.Severity
	}
	if filter.Acknowledged != nil {
		query["acknowledged"] = *filter.Acknowledged
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// Acknowledge marks an event as acknowledged
func (r *EventRepository) Acknowledge(ctx context.Context, id, acknowledgedBy string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{
			"acknowledged":    true,
			"acknowledged_by": acknowledgedBy,
		}},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// CountTotal returns the total number of events
func (r *EventRepository) CountTotal(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountUnacknowledged returns the number of unacknowledged events
func (r *EventRepository) CountUnacknowledged(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"acknowledged": false})
}

// CountCriticalUnacknowledged returns the number of critical unacknowledged events
func (r *EventRepository) CountCriticalUnacknowledged(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"severity":     "critical",
		"acknowledged": false,
	})
}
