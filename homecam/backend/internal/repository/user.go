// Package repository provides data access layer implementations
package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents a user document in MongoDB
type User struct {
	ID                  string     `bson:"id"`
	Username            string     `bson:"username"`
	Email               string     `bson:"email"`
	PasswordHash        string     `bson:"password_hash"`
	Role                string     `bson:"role"`
	TOTPEnabled         bool       `bson:"totp_enabled"`
	TOTPSecret          string     `bson:"totp_secret,omitempty"`
	TOTPSecretPending   string     `bson:"totp_secret_pending,omitempty"`
	CreatedAt           time.Time  `bson:"created_at"`
	LastLogin           *time.Time `bson:"last_login,omitempty"`
	FailedLoginAttempts int        `bson:"failed_login_attempts"`
	LockedUntil         *time.Time `bson:"locked_until,omitempty"`
}

// UserRepository handles user data operations
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// FindByUsername retrieves a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// ExistsByUsername checks if a user with the given username exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	return count > 0, err
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	return count > 0, err
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// Update updates a user document
func (r *UserRepository) Update(ctx context.Context, id string, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	return err
}

// UpdateWithUnset updates and unsets fields in a user document
func (r *UserRepository) UpdateWithUnset(ctx context.Context, id string, setFields, unsetFields bson.M) error {
	update := bson.M{}
	if len(setFields) > 0 {
		update["$set"] = setFields
	}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, update)
	return err
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// FindAll retrieves all users (without sensitive fields)
func (r *UserRepository) FindAll(ctx context.Context, limit int64) ([]*User, error) {
	opts := options.Find().SetLimit(limit)
	// Exclude sensitive fields
	opts.SetProjection(bson.M{
		"password_hash":        0,
		"totp_secret":          0,
		"totp_secret_pending":  0,
	})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// IncrementFailedAttempts increments the failed login attempts counter
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, id string, lockUntil *time.Time) error {
	update := bson.M{
		"$inc": bson.M{"failed_login_attempts": 1},
	}
	if lockUntil != nil {
		update["$set"] = bson.M{"locked_until": lockUntil}
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, update)
	return err
}

// ResetFailedAttempts resets the failed login attempts counter
func (r *UserRepository) ResetFailedAttempts(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{
		"$set":   bson.M{"failed_login_attempts": 0, "last_login": time.Now().UTC()},
		"$unset": bson.M{"locked_until": ""},
	})
	return err
}
