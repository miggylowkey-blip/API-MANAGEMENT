# CI/CD Pipeline Documentation for API-MANAGEMENTZ

This document describes the enhanced CI/CD pipeline for API-MANAGEMENTZ with Kubernetes deployment and security scanning.

## Pipeline Overview

The pipeline consists of 5 stages executed in sequence:

1. **CI** — Code Quality & Security Scanning
2. **Build** — Docker Image Creation & Image Scanning
3. **K8S Validate** — Kubernetes Manifest Validation & Security
4. **Deploy** — Kubernetes Deployment with Rollout Strategy
5. **Verify** — Post-deployment Verification & Smoke Tests

All stages run on the `main` branch push. Pull requests run CI and K8S Validate only (no deployment).

## Stage 1: Code Quality & Security Scanning (CI)

### What it does
- **Go formatting check** — enforces `gofmt` standards
- **Go vet** — static analysis for common Go mistakes
- **Build verification** — ensures code compiles
- **Gitleaks** — scans for hardcoded secrets and credentials
- **Trivy filesystem scan** — finds CVEs in dependencies and configuration
- **SARIF upload** — publishes results to GitHub Security tab

### Why it matters
Catches issues early before they reach production. Gitleaks prevents credentials in git; Trivy identifies vulnerable dependencies.

### Triggers
- Every push to `main`
- Every pull request to `main`
- Manual workflow dispatch

---

## Stage 2: Build Docker Image & Image Scanning (Build)

### What it does
- **Docker Buildx setup** — uses multi-platform builder
- **GitHub Container Registry login** — pushes to ghcr.io
- **Build and push image** — tags with branch, SHA, and semver
- **Trivy image scan** — scans the built image for vulnerabilities
- **SBOM generation** — creates Software Bill of Materials in CycloneDX format
- **Artifact upload** — stores SBOM for compliance

### Image tags generated
- `ghcr.io/owner/repo:main` — latest from main branch
- `ghcr.io/owner/repo:main-abc123def` — main branch with commit SHA
- `ghcr.io/owner/repo:v1.0.0` — semver tag if pushing a release tag

### Why it matters
Ensures the image is scanned for vulnerabilities before deployment. SBOM is required for supply chain security (SLSA, SPDX compliance).

### Secrets required
- `GITHUB_TOKEN` — automatically provided by GitHub Actions

---

## Stage 3: Kubernetes Manifest Validation (K8S Validate)

### What it does
- **Kubeval validation** — checks YAML syntax and schema compliance against Kubernetes specs
- **Kubesec security scan** — validates Kubernetes security best practices
  - Checks for missing resource limits
  - Detects dangerous pod specs (privileged mode, root user)
  - Validates RBAC and network policies
  - Flags insecure configurations

### Score interpretation
- Score < 5 = Critical security issues (blocks deployment)
- Score 5-7 = Warnings (should be addressed)
- Score > 7 = Good security posture

### Why it matters
Prevents misconfigurations and security issues before they reach the cluster. Catches missing resource limits, privilege escalation risks, and network policy gaps.

### Runs on
- CI job (every push and PR)
- K8S Validate job (before deployment)

---

## Stage 4: Deploy to Kubernetes (Deploy)

### Prerequisites
- Kubernetes cluster with kubectl access
- kubeconfig stored in `KUBECONFIG` secret (base64-encoded)
- Cluster must have the `api-managementz` namespace (created if missing)

### What it does

1. **Namespace creation** — ensures `api-managementz` namespace exists
2. **Image update** — updates the API deployment with the new image
3. **Apply manifests** — `kubectl apply -f kubernetes/`
4. **Rollout status** — waits for deployment to reach desired state (5m timeout)
5. **Pod readiness** — waits for pods to pass readiness probes (3m timeout)
6. **Health check** — port-forwards to the pod and tests `/health` endpoint
7. **Replica verification** — confirms all replicas are ready
8. **Status logging** — records deployment state for debugging

### Rollout strategy

The deployment uses Kubernetes' default `RollingUpdate` strategy:
- **maxSurge: 25%** — allows one extra pod during rollout
- **maxUnavailable: 25%** — allows one pod to be down during rollout
- **readinessProbe** — waits for `/health` before routing traffic
- **livenessProbe** — restarts if health check fails after 20 seconds

### Automatic rollback
If the deployment fails:
1. Readiness probe fails → pod marked as not ready
2. Liveness probe fails after 20s → pod restarted
3. Rollout status timeout → deployment stays in old state
4. Verify stage alerts on failure

Manual rollback:
```bash
kubectl rollout undo deployment/api-managementz -n api-managementz
```

### Secrets required
- `KUBECONFIG` — base64-encoded kubeconfig file

---

## Stage 5: Post-deployment Verification (Verify)

### What it does

1. **Smoke tests** — runs a curl pod to test the health endpoint
2. **Database connectivity check** — verifies Postgres is accessible
3. **Redis connectivity check** — verifies Redis responds to PING
4. **Deployment summary** — posts results to GitHub Actions summary

### Why it matters
Confirms the application and dependencies are healthy after deployment. Catches runtime issues that manifest only after deployment.

---

## Environment Variables & Secrets

### Required Secrets

Store these in **GitHub > Settings > Secrets and variables > Actions**:

| Secret | Description |
|--------|-------------|
| `KUBECONFIG` | Base64-encoded kubeconfig for kubectl access. Obtain with: `cat ~/.kube/config | base64` |

### Environment Variables (in workflow)

| Variable | Value | Purpose |
|----------|-------|---------|
| `REGISTRY` | `ghcr.io` | GitHub Container Registry |
| `IMAGE_NAME` | `${{ github.repository }}` | Owner/repo for image tagging |

---

## Setting up GitHub Secrets

### 1. Kubeconfig Secret
```bash
# On your local machine with kubectl access
cat ~/.kube/config | base64

# Copy output, then in GitHub:
# Settings > Secrets and variables > Actions > New repository secret
# Name: KUBECONFIG
# Value: <paste base64-encoded content>
```

### 2. Verify secrets are set
```bash
gh secret list --repo owner/repo
```

---

## Local Testing

### Test Kubernetes manifests locally
```bash
cd API-MANAGEMENTZ/kubernetes

# Validate syntax
kubeval *.yml --strict

# Security scan
kubesec scan *.yml

# Dry-run apply (no actual changes)
kubectl apply -f . --dry-run=client
```

### Test CI steps locally
```bash
# Go formatting
gofmt -l -e .

# Vet
go vet ./...

# Build
go build ./...

# Trivy filesystem scan
trivy fs . --severity CRITICAL,HIGH --exit-code 1
```

---

## Monitoring & Debugging

### Check deployment status
```bash
kubectl get all -n api-managementz
kubectl get deployment api-managementz -n api-managementz -o wide
kubectl get pods -n api-managementz
```

### View pod logs
```bash
kubectl logs -n api-managementz -l app=api-managementz --tail=100
```

### Check rollout history
```bash
kubectl rollout history deployment/api-managementz -n api-managementz
```

### Port-forward to test locally
```bash
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz
curl http://localhost:8080/health
```

---

## Security Considerations

### 1. Image Scanning
- All images are scanned with Trivy after build
- CRITICAL/HIGH vulnerabilities block deployment (exit code 1)
- SBOM is generated for supply chain visibility

### 2. Manifest Security
- Kubesec enforces resource limits, non-root containers, etc.
- Score < 5 blocks deployment
- Network policies and RBAC validated

### 3. Secret Management
- Sensitive values stored in Kubernetes Secrets, not ConfigMaps
- Database credentials retrieved from Secret at runtime
- kubeconfig accessed only during deploy stage

### 4. Credential Protection
- Gitleaks scans for hardcoded secrets
- Prevents secrets from being committed to git
- GitHub token used for container registry login

### 5. Pod Security
- Non-root containers (postgres:16-alpine, redis:7-alpine)
- Resource limits enforced (CPU, memory)
- Readiness/liveness probes for health monitoring
- Namespace isolation for multi-tenant clusters

---

## Troubleshooting

### Deployment times out
```bash
# Check pod logs
kubectl logs -n api-managementz -l app=api-managementz

# Check readiness probe
kubectl describe pod <pod-name> -n api-managementz

# Increase timeout in ci-cd.yml deploy step
```

### Image not found during deployment
```bash
# Verify image exists in registry
docker pull ghcr.io/owner/repo:main

# Check image tag in deployment
kubectl get deployment api-managementz -n api-managementz -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### Security scan failures
```bash
# Check Trivy results
docker run aquasec/trivy image ghcr.io/owner/repo:main

# Check Kubesec score
kubesec scan kubernetes/*.yml
```

### Database migration issues
```bash
# Check Postgres logs
kubectl logs -n api-managementz -l app=postgres

# Verify init script mounted
kubectl describe configmap postgres-initdb -n api-managementz
```

---

## Pipeline Workflow Diagram

```
Push to main branch
        ↓
    ┌───────────────────────────────────┐
    │ CI: Code Quality & Security Scan  │
    │ (gofmt, vet, build, gitleaks,    │
    │  trivy filesystem)                │
    └───────────────────────────────────┘
        ↓ (on success)
    ┌───────────────────────────────────┐
    │ Build: Docker & Image Scanning    │
    │ (docker build, push, trivy image, │
    │  sbom generation)                 │
    └───────────────────────────────────┘
        ↓ (parallel)
    ┌──────────────────┐  ┌──────────────────┐
    │ K8S Validate     │  │ (other jobs)     │
    │ (kubeval,        │  │                  │
    │  kubesec)        │  │                  │
    └──────────────────┘  └──────────────────┘
        ↓ (all success)
    ┌───────────────────────────────────┐
    │ Deploy: Kubernetes Deployment     │
    │ (apply manifests, rollout,        │
    │  health checks)                   │
    └───────────────────────────────────┘
        ↓ (on success)
    ┌───────────────────────────────────┐
    │ Verify: Post-deployment Tests     │
    │ (smoke tests, db check,           │
    │  redis check)                     │
    └───────────────────────────────────┘
        ↓ (final status posted to GitHub)
    ✅ Deployment Complete
```

---

## Next Steps

1. Store kubeconfig in `KUBECONFIG` secret
2. Commit `.github/workflows/ci-cd.yml` to main branch
3. Push and watch the pipeline run
4. Monitor deployment in GitHub Actions tab
5. Verify in cluster: `kubectl get all -n api-managementz`
