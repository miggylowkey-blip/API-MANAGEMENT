# CI/CD Setup Guide for Kubernetes Deployment

This guide walks through setting up the CI/CD pipeline to deploy API-MANAGEMENTZ to Kubernetes with full security scanning.

## Prerequisites

- GitHub repository with API-MANAGEMENTZ code
- Kubernetes cluster with kubectl access
- `kubectl` installed locally
- kubeconfig file configured for your cluster

## Step 1: Prepare Your Kubernetes Cluster

### Create the namespace and resources

```bash
# Apply all Kubernetes manifests
kubectl apply -f API-MANAGEMENTZ/kubernetes/

# Verify namespace created
kubectl get namespace api-managementz

# Check all resources
kubectl get all -n api-managementz
```

### Verify connectivity

```bash
# Test Postgres
kubectl exec -it deployment/postgres -n api-managementz -- pg_isready -U api_managementz

# Test Redis
kubectl exec -it deployment/redis -n api-managementz -- redis-cli ping
```

## Step 2: Prepare Your kubeconfig

The CI/CD pipeline needs credentials to deploy to your cluster. We store this securely as a GitHub Secret.

### Export your kubeconfig

```bash
# Encode your kubeconfig as base64
cat ~/.kube/config | base64

# Copy the entire output (starts with "YXBpVmVyc2lvbjog...")
```

### Store in GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings > Secrets and variables > Actions**
3. Click **New repository secret**
4. Enter:
   - **Name:** `KUBECONFIG`
   - **Secret:** Paste the base64-encoded kubeconfig from above
5. Click **Add secret**

### Verify the secret

```bash
# Using GitHub CLI
gh secret list --repo owner/repo

# Expected output:
# KUBECONFIG                    ***
```

## Step 3: Build and Push Initial Image

The pipeline expects an image to exist before deploying. You can either:

### Option A: Let the pipeline build it (recommended)

1. Commit and push the `.github/workflows/ci-cd.yml` file
2. GitHub Actions will automatically:
   - Run CI checks
   - Build the Docker image
   - Push to GitHub Container Registry (ghcr.io)
   - Deploy to Kubernetes

### Option B: Build and push manually

```bash
# Log in to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin

# Build image
docker build -t ghcr.io/your-org/api-managementz:latest .

# Push image
docker push ghcr.io/your-org/api-managementz:latest

# Load into local cluster (if using kind/minikube)
kind load docker-image ghcr.io/your-org/api-managementz:latest
```

## Step 4: Configure GitHub Secrets (Optional)

If using AWS ECR instead of GitHub Container Registry:

```bash
# Store AWS credentials in GitHub Secrets
gh secret set AWS_ACCESS_KEY_ID --body "your-access-key"
gh secret set AWS_SECRET_ACCESS_KEY --body "your-secret-key"
gh secret set AWS_REGION --body "us-east-1"
gh secret set ECR_REPOSITORY --body "api-managementz"
```

## Step 5: Trigger the Pipeline

### Method 1: Push to main branch

```bash
git add .github/workflows/ci-cd.yml API-MANAGEMENTZ/kubernetes/
git commit -m "Add Kubernetes deployment and enhanced CI/CD"
git push origin main
```

### Method 2: Manual trigger

1. Go to **Actions** tab in GitHub
2. Select **CI / CD Pipeline with Kubernetes Deployment**
3. Click **Run workflow**
4. Select **main** branch and click **Run workflow**

## Step 6: Monitor the Pipeline

### Watch in GitHub Actions UI

1. Go to **Actions** tab
2. Click the latest workflow run
3. Monitor stages: CI → Build → K8S Validate → Deploy → Verify

### Watch in terminal

```bash
# Watch deployment status
kubectl rollout status deployment/api-managementz -n api-managementz --watch

# Watch pod creation
kubectl get pods -n api-managementz --watch

# View logs
kubectl logs -f deployment/api-managementz -n api-managementz
```

## Step 7: Verify Deployment

### Check deployment status

```bash
kubectl get deployment api-managementz -n api-managementz
kubectl get pods -n api-managementz
```

### Test the API health endpoint

```bash
# Port-forward to the service
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz

# In another terminal, test
curl http://localhost:8080/health
```

### Check all resources

```bash
kubectl get all -n api-managementz
```

## Step 8: Set Up Monitoring (Optional)

### View security scan results

1. Go to **Security > Code scanning alerts**
2. Review Trivy findings (filesystem and image scans)
3. Review Gitleaks findings (secret scanning)

### Download SBOM

1. Go to **Actions** tab
2. Click the latest successful workflow
3. Scroll to **Artifacts**
4. Download `sbom` artifact
5. Use for compliance (SLSA, SPDX, etc.)

## Troubleshooting

### Pipeline fails at "Set kubeconfig" stage

**Error:** `No such file or directory`

**Fix:** Verify the `KUBECONFIG` secret is base64-encoded correctly:
```bash
# Re-encode and update the secret
cat ~/.kube/config | base64 | gh secret set KUBECONFIG
```

### Deployment times out waiting for pods

**Error:** `timed out waiting for the condition`

**Fix:** Check pod logs for issues:
```bash
kubectl logs -n api-managementz -l app=api-managementz
kubectl describe pod <pod-name> -n api-managementz
```

### Image pull fails

**Error:** `ImagePullBackOff`

**Fix:** For local clusters, load the image manually:
```bash
kind load docker-image ghcr.io/your-org/api-managementz:latest
```

Or update the deployment to use `imagePullPolicy: Never`:
```bash
kubectl patch deployment api-managementz -n api-managementz -p '{"spec":{"template":{"spec":{"containers":[{"name":"api-managementz","imagePullPolicy":"Never"}]}}}}'
```

### Kubesec security scan blocks deployment

**Error:** `kubesec scan: score < 5`

**Fix:** Review the scan results and update manifests:
```bash
# See full results
kubesec scan kubernetes/*.yml

# Address missing resource limits, run as root, etc.
```

Common fixes:
- Add resource limits to deployments
- Use non-root containers
- Enable securityContext
- Add network policies

## Next Steps

1. ✅ Verify deployment is running
2. ✅ Test API endpoints
3. ✅ Set up alerts for failed deployments
4. ✅ Document any custom configurations
5. ✅ Train team on deployment procedures

## Support

For issues or questions:

1. Check GitHub Actions logs
2. Review [CI-CD-GUIDE.md](./CI-CD-GUIDE.md) for detailed pipeline info
3. Check Kubernetes events: `kubectl describe pod <pod-name> -n api-managementz`
4. Check pod logs: `kubectl logs <pod-name> -n api-managementz`
