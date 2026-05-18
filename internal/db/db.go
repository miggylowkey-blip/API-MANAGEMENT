package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func ConfigFromEnv() Config {
	return Config{
		Host:     getenv("DB_HOST", "localhost"),
		User:     getenv("DB_USER", "postgres"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}
}

func Connect(cfg Config) (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)

	var lastErr error
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			return db, nil
		} else {
			lastErr = err
			time.Sleep(1 * time.Second)
		}
	}

	_ = db.Close()
	return nil, fmt.Errorf("failed to connect to database after 30 attempts: %w", lastErr)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
