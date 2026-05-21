# API-MANAGEMENTZ Kubernetes Deployment

This folder contains a fully separated Kubernetes deployment for the API-MANAGEMENTZ application. Each resource is in its own YAML file so you can apply, inspect, or change them independently.

## File list

- `00-namespace.yml` — creates the `api-managementz` namespace.
- `01-secrets.yml` — stores sensitive values such as DB credentials and the auth secret.
- `02-configmap.yml` — stores non-secret application configuration values.
- `03-postgres-initdb-configmap.yml` — provides the SQL initialization script to Postgres.
- `04-postgres-pvc.yml` — requests persistent storage for the Postgres database.
- `05-postgres-service.yml` — internal service for Postgres access.
- `06-postgres-deployment.yml` — Postgres deployment configuration.
- `07-redis-service.yml` — internal service for Redis access.
- `08-redis-deployment.yml` — Redis deployment configuration.
- `09-api-service.yml` — external LoadBalancer service for the API.
- `10-api-deployment.yml` — API deployment which runs the `api-managementz` container.
- `resource.yml` — namespaced resource quota for the `api-managementz` namespace.
- `manifest.yml` — a note file pointing to the separated manifests.

## What each file does

### `00-namespace.yml`
Creates the `api-managementz` namespace to isolate the stack. Namespaced resources are all scoped here so they do not conflict with other clusters or apps.

### `01-secrets.yml`
Stores sensitive environment values in a Kubernetes Secret:
- `AUTH_SECRET` — application secret used for hashing and token logic.
- `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` — database credentials for Postgres.

The API deployment references this secret so credentials are not stored in plain text in the ConfigMap.

### `02-configmap.yml`
Stores non-sensitive configuration values as environment variables:
- `PORT` — service port used by the app container.
- `RATE_LIMIT_RPS` and `RATE_LIMIT_BURST` — rate limiter tuning.
- `DB_HOST` — service hostname for Postgres.
- `REDIS_URL` — Redis connection string.

This keeps configuration separate from secrets and makes overrides easy.

### `03-postgres-initdb-configmap.yml`
Contains the SQL schema for the API-MANAGEMENTZ database.

The Postgres container mounts this ConfigMap to `/docker-entrypoint-initdb.d`, so when the database is first created it initializes users, api_keys, usage tracking, and revocation metadata.

### `04-postgres-pvc.yml`
Requests a PersistentVolumeClaim of `5Gi` using `ReadWriteOnce`.

This ensures Postgres data survives pod restarts and rescheduling, instead of being lost when the pod cycles.

### `05-postgres-service.yml`
Defines a ClusterIP service for Postgres on port `5432`.

Other pods in the namespace can connect to Postgres using `postgres:5432`.

### `06-postgres-deployment.yml`
Deploys Postgres with:
- `postgres:16-alpine`
- environment variables from the Secret
- startup readiness using `pg_isready`
- mounted init SQL from `postgres-initdb`
- persistent storage via `postgres-pvc`

This deployment runs a single replica for the database.

### `07-redis-service.yml`
Defines a ClusterIP service for Redis on port `6379`.

The API pod connects to Redis with `redis://redis:6379`.

### `08-redis-deployment.yml`
Deploys Redis with `redis:7-alpine` and a readiness probe using `redis-cli ping`.

A single Redis replica is enough for the rate limiter in this baseline setup.

### `09-api-service.yml`
Exposes the API via a `LoadBalancer` service on port `80`.

It forwards traffic to the API deployment on container port `8080`.

### `10-api-deployment.yml`
Deploys the API application with:
- image `api-managementz:latest`
- `imagePullPolicy: IfNotPresent` for local clusters
- environment from the ConfigMap and Secret
- readiness and liveness probes against `/health`
- resource requests and limits

This deployment runs 2 replicas for availability.

### `resource.yml`
Defines a `ResourceQuota` for the namespace limiting the total CPU to `1` and memory to `1Gi`.

This helps prevent the namespace from consuming too many cluster resources.

## How to apply the stack

From `API-MANAGEMENTZ/kubernetes`:

```bash
kubectl apply -f .
```

If your cluster does not support `LoadBalancer`, use:

```bash
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz
```

## Notes

- Build and load the API image into your cluster before applying, or replace `api-managementz:latest` with a registry image reference.
- The Postgres init script only runs when an empty database is first created.
- Use `kubectl get all -n api-managementz` to inspect the deployment state.
- The stack assumes the Go API app reads `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, and `REDIS_URL` from the environment.
