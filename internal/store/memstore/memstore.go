package memstore

import (
	"api-managementz/internal/auth"
	"api-managementz/internal/store"
	"context"
	"strings"
	"sync"
	"time"
)

type keyRecord struct {
	userID     int64
	hash       string
	usageCount int64
	lastUsedAt *time.Time
	revoked    bool
	revokedAt  *time.Time
}

type Store struct {
	mu sync.Mutex

	nextID int64
	users  map[int64]auth.User
	byMail map[string]int64

	keyByHash   map[string]*keyRecord
	keyByUserID map[int64]*keyRecord
}

func New() *Store {
	return &Store{
		nextID:      1,
		users:       make(map[int64]auth.User),
		byMail:      make(map[string]int64),
		keyByHash:   make(map[string]*keyRecord),
		keyByUserID: make(map[int64]*keyRecord),
	}
}

func (s *Store) CreateUser(ctx context.Context, user auth.User) (auth.User, error) {
	_ = ctx
	email := strings.TrimSpace(strings.ToLower(user.Email))
	if email == "" {
		return auth.User{}, store.ErrConflict
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byMail[email]; exists {
		return auth.User{}, store.ErrConflict
	}
	user.ID = s.nextID
	s.nextID++
	user.Email = email
	s.users[user.ID] = user
	s.byMail[email] = user.ID
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	_ = ctx
	email = strings.TrimSpace(strings.ToLower(email))
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byMail[email]
	if !ok {
		return auth.User{}, store.ErrNotFound
	}
	return s.users[id], nil
}

func (s *Store) SaveAPIKey(ctx context.Context, userID int64, apiKeyHash string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.keyByUserID[userID]; ok {
		delete(s.keyByHash, old.hash)
	}
	rec := &keyRecord{userID: userID, hash: apiKeyHash}
	s.keyByHash[apiKeyHash] = rec
	s.keyByUserID[userID] = rec
	return nil
}

func (s *Store) FindUserIDByAPIKeyHash(ctx context.Context, apiKeyHash string) (int64, bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.keyByHash[apiKeyHash]
	if !ok || rec.revoked {
		return 0, false, nil
	}
	return rec.userID, true, nil
}

func (s *Store) RevokeAPIKey(ctx context.Context, userID int64) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.keyByUserID[userID]
	if !ok || rec.revoked {
		return store.ErrNotFound
	}
	now := time.Now()
	rec.revoked = true
	rec.revokedAt = &now
	return nil
}

func (s *Store) RecordKeyUsage(ctx context.Context, userID int64) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.keyByUserID[userID]
	if !ok {
		return store.ErrNotFound
	}
	now := time.Now()
	rec.usageCount++
	rec.lastUsedAt = &now
	return nil
}

func (s *Store) GetKeyInfo(ctx context.Context, userID int64) (store.KeyInfo, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.keyByUserID[userID]
	if !ok {
		return store.KeyInfo{}, store.ErrNotFound
	}
	return store.KeyInfo{
		UserID:     rec.userID,
		UsageCount: rec.usageCount,
		LastUsedAt: rec.lastUsedAt,
		Revoked:    rec.revoked,
		RevokedAt:  rec.revokedAt,
	}, nil
}
func (s *Store) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}
