package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type ctxKey string

const (
	ctxKeyRequestID ctxKey = "request_id"
	ctxKeyUserID    ctxKey = "user_id"
)

type Logger struct {
	log *slog.Logger
}

func NewLoggerFromEnv() *Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	if os.Getenv("AUDIT_LOG_LEVEL") == "debug" {
		level.Set(slog.LevelDebug)
	}
	return &Logger{
		log: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})),
	}
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			rid = newRequestID()
		}
		w.Header().Set("X-Request-Id", rid)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

type AuthEvent struct {
	Name    string
	Email   string
	UserID  int64
	Success bool
	When    time.Time
}

func (l *Logger) AuthEvent(ctx context.Context, e AuthEvent) {
	rid, _ := ctx.Value(ctxKeyRequestID).(string)
	attrs := []slog.Attr{
		slog.String("type", "auth"),
		slog.String("name", e.Name),
		slog.String("request_id", rid),
		slog.String("email", e.Email),
		slog.Int64("user_id", e.UserID),
		slog.Bool("success", e.Success),
		slog.Time("ts", e.When),
	}
	l.log.Info("audit", slog.Attr{Key: "audit", Value: slog.GroupValue(attrs...)})
}

type KeyEvent struct {
	Name    string
	UserID  int64
	Success bool
	When    time.Time
}

func (l *Logger) KeyEvent(ctx context.Context, e KeyEvent) {
	rid, _ := ctx.Value(ctxKeyRequestID).(string)
	attrs := []slog.Attr{
		slog.String("type", "api_key"),
		slog.String("name", e.Name),
		slog.String("request_id", rid),
		slog.Int64("user_id", e.UserID),
		slog.Bool("success", e.Success),
		slog.Time("ts", e.When),
	}
	l.log.Info("audit", slog.Attr{Key: "audit", Value: slog.GroupValue(attrs...)})
}

func WithAuditMiddleware(l *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &recorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rr, r)

		rid, _ := r.Context().Value(ctxKeyRequestID).(string)
		uid, _ := r.Context().Value(ctxKeyUserID).(int64)

		attrs := []slog.Attr{
			slog.String("type", "http"),
			slog.String("request_id", rid),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rr.status),
			slog.Int("bytes", rr.bytes),
			slog.Int64("user_id", uid),
			slog.Duration("duration_ms", time.Since(start).Round(time.Millisecond)),
		}
		l.log.Info("audit", slog.Attr{Key: "audit", Value: slog.GroupValue(attrs...)})
	})
}

type recorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = 200
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}
