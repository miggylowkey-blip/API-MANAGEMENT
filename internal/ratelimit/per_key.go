package ratelimit

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type PerKeyLimiter struct {
	mu sync.Mutex

	rps   float64
	burst float64

	buckets map[string]*bucket
}

type Limiter interface {
	Allow(key string) bool
}

func NewPerKeyLimiter(rps float64, burst int) *PerKeyLimiter {
	if rps <= 0 {
		rps = 5
	}
	if burst <= 0 {
		burst = 10
	}
	return &PerKeyLimiter{
		rps:     rps,
		burst:   float64(burst),
		buckets: make(map[string]*bucket),
	}
}

func NewPerKeyLimiterFromEnv() *PerKeyLimiter {
	return NewPerKeyLimiter(envFloat("RATE_LIMIT_RPS", 5), envInt("RATE_LIMIT_BURST", 10))
}

func (l *PerKeyLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.burst - 1, last: now}
		return true
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = minFloat(l.burst, b.tokens+(elapsed*l.rps))
		b.last = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func envFloat(key string, def float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
