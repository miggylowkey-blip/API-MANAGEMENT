# API-MANAGEMENTZ

A security-focused API key management service built in Go. Handles key issuance, rotation, revocation, usage tracking, and rate limiting — the infrastructure layer for controlling API access.

## Live

Deployed on AWS ECS with RDS Postgres and Redis rate limiting.

```
http://3.25.207.102:8080/health
```

## Features

- **Key issuance** — register or login to receive an API key
- **Key rotation** — every login issues a fresh key, old one invalidated instantly
- **Key revocation** — kill a key immediately via `DELETE /key`
- **Usage tracking** — request count and last used timestamp per key
- **Rate limiting** — per-key token bucket backed by Redis, persists across restarts
- **Audit logging** — every auth and key event logged as structured JSON
- **Input validation** — email format, password length, bcrypt 72-byte truncation cap
- **Health check** — `GET /health` verifies app and database connectivity

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /register | None | Create account, returns API key |
| POST | /login | None | Authenticate, rotates and returns new API key |
| GET | /whoami | API Key | Verify key, returns user ID |
| GET | /key/info | API Key | Usage count, last used, revocation status |
| DELETE | /key | API Key | Revoke current API key instantly |
| GET | /health | None | App and database health status |

## Authentication

Pass your API key on every request:

```bash
Authorization: Bearer YOUR_API_KEY
# or
X-API-Key: YOUR_API_KEY
```

## Run locally

**Prerequisites:** Docker, Go 1.24+

```bash
# Start Postgres and Redis
docker compose up -d

# Copy and fill environment variables
cp .env.example .env

# Run the API
go run ./cmd/api
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| AUTH_SECRET | — | Required in production — salts API key hashes |
| DB_HOST | localhost | Postgres host |
| DB_USER | — | Postgres user |
| DB_PASSWORD | — | Postgres password |
| DB_NAME | — | Postgres database name |
| DB_SSLMODE | disable | Postgres SSL mode (`require` in production) |
| REDIS_URL | redis://localhost:6379 | Redis connection URL |
| RATE_LIMIT_RPS | 5 | Requests per second per key |
| RATE_LIMIT_BURST | 10 | Burst allowance per key |

## Architecture

```
cmd/api/          — entry point, wires everything together
internal/
  auth/           — key generation, hashing, password bcrypt
  audit/          — structured JSON audit logging
  db/             — postgres connection and pooling
  handlers/       — HTTP handlers and input validation
  ratelimit/      — per-key rate limiting (in-memory + Redis)
  store/          — store interface
    memstore/     — in-memory fallback for local dev
    postgresstore/— production postgres implementation
migrations/       — SQL schema files
```

## CI/CD Pipeline

Every push to `main` runs:

**CI**
- `gofmt` — formatting gate
- `go vet` — static analysis
- `go build` — build verification
- **Gitleaks** — secret and credential scanning
- **Trivy** — CVE scanning on filesystem and dependencies (CRITICAL/HIGH)

**CD** (on CI pass)
- Builds Docker image
- Tags with commit SHA and `latest`
- Pushes to AWS ECR
- ECS pulls and deploys automatically

## Stack

- **Go** — API, middleware, business logic
- **PostgreSQL** — persistent storage (RDS in production)
- **Redis** — rate limit counters (ElastiCache-ready)
- **Docker** — containerized local dev and production
- **AWS ECS + RDS** — production deployment
- **GitHub Actions** — CI/CD pipeline
