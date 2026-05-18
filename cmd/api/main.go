package main

import (
	"api-managementz/internal/audit"
	"api-managementz/internal/db"
	"api-managementz/internal/handlers"
	"api-managementz/internal/ratelimit"
	"api-managementz/internal/store"
	"api-managementz/internal/store/memstore"
	"api-managementz/internal/store/postgresstore"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	st := buildStoreOrFallback()
	rl := buildRateLimiter()
	aud := audit.NewLoggerFromEnv()

	svc := handlers.NewService(st, rl, aud)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", svc.Register)
	mux.HandleFunc("POST /login", svc.Login)
	mux.HandleFunc("DELETE /key", svc.RevokeKey)
	mux.HandleFunc("GET /key/info", svc.KeyInfo)
	mux.HandleFunc("GET /health", svc.HealthCheck)
	mux.Handle("GET /whoami", http.HandlerFunc(svc.WhoAmI))

	handler := audit.WithRequestID(audit.WithAuditMiddleware(aud, mux))

	addr := ":" + port
	fmt.Println("API-MANAGEMENTZ listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func buildRateLimiter() ratelimit.Limiter {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf(">>> Redis URL invalid, using in-memory limiter: %v", err)
		return ratelimit.NewPerKeyLimiterFromEnv()
	}

	client := redis.NewClient(opt)
	log.Println(">>> Using Redis rate limiter")
	return ratelimit.NewRedisLimiter(client, 5)
}

func buildStoreOrFallback() store.Store {
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("DB_NAME") == "" {
		log.Println(">>> Using memstore")
		return memstore.New()
	}

	log.Println(">>> Connecting to Postgres...")
	sqlDB, err := db.Connect(db.ConfigFromEnv())
	if err != nil {
		log.Printf(">>> DB connect failed, using memstore: %v", err)
		return memstore.New()
	}
	log.Println(">>> Connected to Postgres")
	return postgresstore.New(sqlDB)
}
