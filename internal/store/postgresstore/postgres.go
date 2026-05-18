package postgresstore

import (
	"api-managementz/internal/auth"
	"api-managementz/internal/store"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, user auth.User) (auth.User, error) {
	var created auth.User
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash`,
		user.Email,
		user.PasswordHash,
	).Scan(&created.ID, &created.Email, &created.PasswordHash)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return auth.User{}, store.ErrConflict
		}
		return auth.User{}, err
	}
	return created, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	var u auth.User
	err := s.db.QueryRowContext(ctx, `SELECT id, email, password_hash FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, store.ErrNotFound
		}
		return auth.User{}, err
	}
	return u, nil
}

func (s *Store) SaveAPIKey(ctx context.Context, userID int64, apiKeyHash string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO api_keys (user_id, api_key_hash, usage_count, revoked, revoked_at, last_used_at)
		 VALUES ($1, $2, 0, FALSE, NULL, NULL)
		 ON CONFLICT (user_id) DO UPDATE
		   SET api_key_hash = EXCLUDED.api_key_hash,
		       usage_count  = 0,
		       revoked      = FALSE,
		       revoked_at   = NULL,
		       last_used_at = NULL`,
		userID,
		apiKeyHash,
	)
	return err
}

func (s *Store) FindUserIDByAPIKeyHash(ctx context.Context, apiKeyHash string) (int64, bool, error) {
	var userID int64
	err := s.db.QueryRowContext(
		ctx,
		`SELECT user_id FROM api_keys WHERE api_key_hash=$1 AND revoked=FALSE`,
		apiKeyHash,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return userID, true, nil
}

func (s *Store) RevokeAPIKey(ctx context.Context, userID int64) error {
	res, err := s.db.ExecContext(
		ctx,
		`UPDATE api_keys SET revoked=TRUE, revoked_at=NOW() WHERE user_id=$1 AND revoked=FALSE`,
		userID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) RecordKeyUsage(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE api_keys SET usage_count=usage_count+1, last_used_at=NOW() WHERE user_id=$1`,
		userID,
	)
	return err
}

func (s *Store) GetKeyInfo(ctx context.Context, userID int64) (store.KeyInfo, error) {
	var info store.KeyInfo
	var lastUsedAt sql.NullTime
	var revokedAt sql.NullTime

	err := s.db.QueryRowContext(
		ctx,
		`SELECT user_id, usage_count, last_used_at, revoked, revoked_at
		 FROM api_keys WHERE user_id=$1`,
		userID,
	).Scan(&info.UserID, &info.UsageCount, &lastUsedAt, &info.Revoked, &revokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.KeyInfo{}, store.ErrNotFound
		}
		return store.KeyInfo{}, err
	}

	if lastUsedAt.Valid {
		t := lastUsedAt.Time
		info.LastUsedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		info.RevokedAt = &t
	}
	return info, nil
}
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

var _ = time.Now
