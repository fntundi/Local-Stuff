package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditLog represents an audit log document in MongoDB
type AuditLog struct {
	ID           string                 `bson:"id"`
	UserID       string                 `bson:"user_id"`
	Username     string                 `bson:"username"`
	Action       string                 `bson:"action"`
	ResourceType string                 `bson:"resource_type"`
	ResourceID   string                 `bson:"resource_id,omitempty"`
	Details      map[string]interface{} `bson:"details,omitempty"`
	IPAddress    string                 `bson:"ip_address"`
	CreatedAt    time.Time              `bson:"created_at"`
}

// AuditRepository handles audit log data operations
type AuditRepository struct {
	collection *mongo.Collection
}

// NewAuditRepository creates a new AuditRepository
func NewAuditRepository(db *mongo.Database) *AuditRepository {
	return &AuditRepository{
		collection: db.Collection("audit_logs"),
	}
}

// Create inserts a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *AuditLog) error {
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// AuditFilter represents filter options for listing audit logs
type AuditFilter struct {
	Action       string
	ResourceType string
	UserID       string
	Limit        int64
}

// FindAll retrieves audit logs with optional filters
func (r *AuditRepository) FindAll(ctx context.Context, filter AuditFilter) ([]*AuditLog, error) {
	query := bson.M{}

	if filter.Action != "" {
		query["action"] = filter.Action
	}
	if filter.ResourceType != "" {
		query["resource_type"] = filter.ResourceType
	}
	if filter.UserID != "" {
		query["user_id"] = filter.UserID
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

	var logs []*AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
