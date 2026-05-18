package handlers

import (
	"api-managementz/internal/audit"
	"api-managementz/internal/auth"
	"api-managementz/internal/ratelimit"
	"api-managementz/internal/store"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	store store.Store
	rl    ratelimit.Limiter
	audit *audit.Logger
}

func NewService(st store.Store, rl ratelimit.Limiter, aud *audit.Logger) *Service {
	return &Service{store: st, rl: rl, audit: aud}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validateEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1 && at+dot < len(email)-1
}

func ValidateRequest(email, password string) (string, int) {
	if len(email) > 72 {
		return "email_too_long", http.StatusBadRequest
	}
	if len(password) < 8 {
		return "password_too_short", http.StatusBadRequest
	}
	if !validateEmail(email) {
		return "invalid_email", http.StatusBadRequest
	}
	return "", 0
}
func (s *Service) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email_and_password_required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := auth.NewUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"invalid_input"}`, http.StatusBadRequest)
		return
	}

	created, err := s.store.CreateUser(ctx, user)
	if err != nil {
		if err == store.ErrConflict {
			http.Error(w, `{"error":"email_already_exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"db_error"}`, http.StatusInternalServerError)
		return
	}
	if req.Email == "" || req.Password == "" {

		http.Error(w, `{"error":"email_and_password_required"}`, http.StatusBadRequest)
		return
	}
	if errMsg, status := ValidateRequest(created.Email, req.Password); errMsg != "" {
		http.Error(w, `{"error":"`+errMsg+`"}`, status)
		return
	}

	// Create an API key for the new user.
	key, keyHash, err := auth.NewAPIKey()
	if err != nil {
		http.Error(w, `{"error":"could_not_generate_api_key"}`, http.StatusInternalServerError)
		return
	}

	if err := s.store.SaveAPIKey(ctx, created.ID, keyHash); err != nil {
		http.Error(w, `{"error":"db_error"}`, http.StatusInternalServerError)
		return
	}

	s.audit.AuthEvent(ctx, audit.AuthEvent{
		Name:    "register",
		Email:   created.Email,
		UserID:  created.ID,
		Success: true,
		When:    time.Now(),
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id": created.ID,
		"email":   created.Email,
		"api_key": key, // only shown once
	})
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email_and_password_required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.audit.AuthEvent(ctx, audit.AuthEvent{Name: "login", Email: req.Email, Success: false, When: time.Now()})
		http.Error(w, `{"error":"invalid_credentials"}`, http.StatusUnauthorized)
		return
	}
	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		s.audit.AuthEvent(ctx, audit.AuthEvent{Name: "login", Email: req.Email, UserID: user.ID, Success: false, When: time.Now()})
		http.Error(w, `{"error":"invalid_credentials"}`, http.StatusUnauthorized)
		return
	}

	// Rotate API key on login (simple, safer default).
	key, keyHash, err := auth.NewAPIKey()
	if err != nil {
		http.Error(w, `{"error":"could_not_generate_api_key"}`, http.StatusInternalServerError)
		return
	}
	if err := s.store.SaveAPIKey(ctx, user.ID, keyHash); err != nil {
		http.Error(w, `{"error":"db_error"}`, http.StatusInternalServerError)
		return
	}

	s.audit.AuthEvent(ctx, audit.AuthEvent{
		Name:    "login",
		Email:   user.Email,
		UserID:  user.ID,
		Success: true,
		When:    time.Now(),
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"api_key": key,
	})
}

func (s *Service) WhoAmI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	apiKey := auth.ExtractAPIKey(r)
	if apiKey == "" {
		http.Error(w, `{"error":"api_key_required"}`, http.StatusUnauthorized)
		return
	}

	if len(apiKey) < 16 || len(apiKey) > 64 {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
		return
	}

	ctx, userID, ok, err := s.authenticateAPIKey(ctx, apiKey)
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
		return
	}

	go s.store.RecordKeyUsage(r.Context(), userID)

	if !s.rl.Allow(apiKey) {
		http.Error(w, `{"error":"rate_limited"}`, http.StatusTooManyRequests)
		return
	}

	s.audit.KeyEvent(ctx, audit.KeyEvent{
		Name:    "whoami",
		UserID:  userID,
		Success: true,
		When:    time.Now(),
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id": userID,
	})
}

func (s *Service) authenticateAPIKey(ctx context.Context, apiKey string) (context.Context, int64, bool, error) {
	keyHash := auth.HashAPIKey(apiKey)
	userID, ok, err := s.store.FindUserIDByAPIKeyHash(ctx, keyHash)
	if err != nil {
		return ctx, 0, false, err
	}
	if ok {
		ctx = audit.SetUserID(ctx, userID)
	}
	return ctx, userID, ok, nil
}

func (s *Service) RevokeKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	apiKey := auth.ExtractAPIKey(r)
	if apiKey == "" {
		http.Error(w, `{"error":"api_key_required"}`, http.StatusUnauthorized)
		return
	}

	ctx, userID, ok, err := s.authenticateAPIKey(ctx, apiKey)
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
		return
	}
	go s.store.RecordKeyUsage(r.Context(), userID)

	if err := s.store.RevokeAPIKey(ctx, userID); err != nil {
		http.Error(w, `{"error":"revocation_failed"}`, http.StatusInternalServerError)
		return
	}

	s.audit.KeyEvent(ctx, audit.KeyEvent{
		Name:    "revoke",
		UserID:  userID,
		Success: true,
		When:    time.Now(),
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "api_key_revoked",
	})
}

func (s *Service) KeyInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	apiKey := auth.ExtractAPIKey(r)
	if apiKey == "" {
		http.Error(w, `{"error":"api_key_required"}`, http.StatusUnauthorized)
		return
	}

	ctx, userID, ok, err := s.authenticateAPIKey(ctx, apiKey)
	if err != nil {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
		return
	}

	info, err := s.store.GetKeyInfo(ctx, userID)
	if err != nil {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":      info.UserID,
		"usage_count":  info.UsageCount,
		"last_used_at": info.LastUsedAt,
		"revoked":      info.Revoked,
		"revoked_at":   info.RevokedAt,
	})
}
func (s *Service) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	err := s.store.Ping(ctx)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "unhealthy",
			"db":     "unreachable",
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
		"db":     "reachable",
	})
}
