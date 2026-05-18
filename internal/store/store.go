package store

import (
	"api-managementz/internal/auth"
	"context"
	"errors"
	"time"
)

type KeyInfo struct {
	UserID     int64
	UsageCount int64
	LastUsedAt *time.Time
	Revoked    bool
	RevokedAt  *time.Time
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// Store abstracts the database. Implement this with Postgres/MySQL/etc.
type Store interface {
	CreateUser(ctx context.Context, user auth.User) (auth.User, error)
	GetUserByEmail(ctx context.Context, email string) (auth.User, error)

	// SaveAPIKey stores the hashed api key for a user
	SaveAPIKey(ctx context.Context, userID int64, apiKeyHash string) error

	// FindUserIDByAPIKeyHash returns the owner of the key if it exists
	FindUserIDByAPIKeyHash(ctx context.Context, apiKeyHash string) (int64, bool, error)

	// RevokeAPIKey revokes(deletes) the API initialized
	RevokeAPIKey(ctx context.Context, userID int64) error
	// RecordKeyUsage Records and tracks the usage of a key
	RecordKeyUsage(ctx context.Context, userID int64) error
	// GetKeyInfo get's the user's information or who the key belongs to
	GetKeyInfo(ctx context.Context, userID int64) (KeyInfo, error)
	// Pings the database to see if it's healthy
	Ping(ctx context.Context) error
}
