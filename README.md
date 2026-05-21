# API-MANAGEMENTZ

A security-focused API key management service built in Go. Handles API key issuance, rotation, revocation, usage tracking, and per-key rate limiting.

## Overview

- Local development uses Docker Compose with Postgres and Redis.
- Kubernetes deployment manifests are stored under `kubernetes/`.
- CI/CD is implemented with GitHub Actions and includes secret scanning, Trivy, and image build validation.
- Container images are published to GitHub Container Registry (GHCR).

## Features

- **Key issuance** — register or login to receive an API key
- **Key rotation** — every login issues a fresh key and invalidates the old one
- **Key revocation** — revoke API keys immediately via `DELETE /key`
- **Usage tracking** — request count and last used timestamp per key
- **Rate limiting** — per-key token bucket backed by Redis
- **Audit logging** — structured JSON audit entries for auth and key events
- **Input validation** — email validation, password length enforcement, bcrypt handling
- **Health check** — `GET /health` verifies app and database connectivity

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /register | None | Create account and receive an API key |
| POST | /login | None | Authenticate and rotate API key |
| GET | /whoami | API Key | Verify API key and return user info |
| GET | /key/info | API Key | Get usage and revocation status |
| DELETE | /key | API Key | Revoke current API key |
| GET | /health | None | App and database health status |

## Authentication

Pass the API key in one of these headers:

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

# Copy and configure environment variables
cp .env.example .env
# Edit .env and fill in any secret values
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| AUTH_SECRET | — | Required for API key hashing and signing |
| DB_HOST | localhost | Postgres host |
| DB_USER | — | Postgres username |
| DB_PASSWORD | — | Postgres password |
| DB_NAME | — | Postgres database name |
| DB_SSLMODE | disable | Postgres SSL mode |
| REDIS_URL | redis://localhost:6379 | Redis connection URL |
| RATE_LIMIT_RPS | 5 | Requests per second per key |
| RATE_LIMIT_BURST | 10 | Burst allowed per key |

## Kubernetes deployment

The Kubernetes manifests live in `kubernetes/`.

Use the `.env` file for sensitive values and keep it out of source control. `API-MANAGEMENTZ/.env` is ignored by `.gitignore`.

Apply the stack from the repository root:

```bash
kubectl apply -f kubernetes/
```

For secrets, create the Kubernetes secret from `.env` before applying the stack:

```bash
kubectl create namespace api-managementz --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic api-managementz-secrets \
  --from-env-file=.env \
  -n api-managementz
```

If you are using a remote registry, update `kubernetes/10-api-deployment.yml` with the correct image reference.

If your cluster does not support `LoadBalancer`, port-forward the API service instead:

```bash
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz
```

## Architecture

```
cmd/api/          — application entry point
internal/
  auth/           — key generation, hashing, API key auth logic
  audit/          — structured JSON audit logging
  db/             — PostgreSQL connection and pooling
  handlers/       — HTTP request handlers and validation
  ratelimit/      — per-key rate limiting backed by Redis
  store/          — store interface
    memstore/     — local in-memory fallback
    postgresstore/— PostgreSQL implementation
migrations/       — SQL schema files
```

## CI/CD Pipeline

The GitHub Actions pipeline validates and scans the code on every push to `main`.

**CI**
- `gofmt` — format checking
- `go vet` — static analysis
- `go build` — build verification
- `Gitleaks` — secret scanning
- `Trivy` — vulnerability scanning

**Build**
- Build and push Docker images to GitHub Container Registry
- Scan the pushed image with Trivy
- Generate an SBOM with Syft

**Kubernetes validation**
- Validate manifests with kubeval
- Run security checks with kubesec

## Stack

- **Go** — API and business logic
- **PostgreSQL** — persistent storage
- **Redis** — rate limit state
- **Docker** — container packaging
- **GitHub Actions** — CI/CD automation
