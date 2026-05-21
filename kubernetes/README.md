# Kubernetes deployment for API-MANAGEMENTZ

This folder contains the Kubernetes manifests for the API-MANAGEMENTZ stack. Each manifest is separated so you can deploy and update resources independently.

## Files

- `00-namespace.yml` — defines the `api-managementz` namespace
- `01-secrets.yml` — stores database and application secrets
- `02-configmap.yml` — stores non-sensitive runtime configuration
- `03-postgres-initdb-configmap.yml` — PostgreSQL initialization SQL
- `04-postgres-pvc.yml` — persistent volume claim for Postgres data
- `05-postgres-service.yml` — internal PostgreSQL service
- `06-postgres-deployment.yml` — PostgreSQL deployment
- `07-redis-service.yml` — internal Redis service
- `08-redis-deployment.yml` — Redis deployment
- `09-api-service.yml` — API service exposure
- `10-api-deployment.yml` — API application deployment
- `resource.yml` — namespace resource quota
- `manifest.yml` — helper note pointing to this folder

## How to use this folder

Before applying the manifests, create the API secret from `.env`:

```bash
cp .env.example .env
# edit .env with your real secret values
kubectl create namespace api-managementz --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic api-managementz-secrets \
  --from-env-file=.env \
  -n api-managementz
```

Then apply the full stack from `API-MANAGEMENTZ/kubernetes`:

```bash
kubectl apply -f kubernetes/
```

If your cluster cannot expose a `LoadBalancer`, forward the API port locally instead:

```bash
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz
```

## API image configuration

The API deployment uses a container image reference in `10-api-deployment.yml`. Replace `api-managementz:latest` with your registry and tag, for example:

```yaml
image: ghcr.io/<owner>/api-managementz:<sha>
```

For local development, `imagePullPolicy: IfNotPresent` is fine. For remote registry deployments, consider `imagePullPolicy: Always`.

## File summary

- `00-namespace.yml`: creates `api-managementz`.
- `01-secrets.yml`: contains `AUTH_SECRET`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSLMODE`.
- `02-configmap.yml`: contains `PORT`, `DB_HOST`, `REDIS_URL`, and rate limiting configuration.
- `03-postgres-initdb-configmap.yml`: holds SQL to initialize the Postgres schema.
- `04-postgres-pvc.yml`: requests `5Gi` storage with `ReadWriteOnce`.
- `05-postgres-service.yml`: exposes PostgreSQL internally on port `5432`.
- `06-postgres-deployment.yml`: deploys Postgres with init SQL and persistent storage.
- `07-redis-service.yml`: exposes Redis internally on port `6379`.
- `08-redis-deployment.yml`: deploys Redis with a readiness probe.
- `09-api-service.yml`: exposes the API on port `80`.
- `10-api-deployment.yml`: deploys the application with environment variables and health probes.
- `resource.yml`: limits namespace CPU and memory usage.

## Notes

- The Postgres init SQL only runs when the database is created from an empty volume.
- Ensure the API app environment variables match the values in `01-secrets.yml` and `02-configmap.yml`.
- Use `kubectl get all -n api-managementz` to inspect the deployed resources.
- When you change the API image reference, redeploy the stack or update the deployment manually.
