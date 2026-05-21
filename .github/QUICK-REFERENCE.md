# Kubernetes Deployment Quick Reference

Fast lookup for common commands and troubleshooting for API-MANAGEMENTZ Kubernetes deployment.

## Quick Deploy Commands

```bash
# Apply all Kubernetes manifests
kubectl apply -f API-MANAGEMENTZ/kubernetes/

# Watch deployment status in real-time
kubectl rollout status deployment/api-managementz -n api-managementz --watch

# View all resources in the namespace
kubectl get all -n api-managementz

# Port-forward to test locally
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz

# Test health endpoint
curl http://localhost:8080/health
```

## Namespace Management

```bash
# List all namespaces
kubectl get namespaces

# Create namespace
kubectl create namespace api-managementz

# Delete entire namespace and all resources
kubectl delete namespace api-managementz

# Switch default namespace
kubectl config set-context --current --namespace=api-managementz
```

## Pod Commands

```bash
# List pods
kubectl get pods -n api-managementz

# Get detailed pod info
kubectl describe pod <pod-name> -n api-managementz

# View pod logs
kubectl logs <pod-name> -n api-managementz

# Follow logs (tail)
kubectl logs -f <pod-name> -n api-managementz

# View logs from all pods with label
kubectl logs -l app=api-managementz -n api-managementz --tail=100

# Execute command in pod
kubectl exec -it <pod-name> -n api-managementz -- /bin/sh

# Copy file from pod
kubectl cp <pod-name>:/path/to/file ./local-file -n api-managementz

# Copy file to pod
kubectl cp ./local-file <pod-name>:/path/to/file -n api-managementz
```

## Deployment Commands

```bash
# Get deployment info
kubectl get deployment api-managementz -n api-managementz

# Get detailed deployment info
kubectl describe deployment api-managementz -n api-managementz

# Scale replicas
kubectl scale deployment api-managementz --replicas=3 -n api-managementz

# Update image
kubectl set image deployment/api-managementz \
  api-managementz=ghcr.io/owner/api-managementz:new-tag \
  -n api-managementz

# Check rollout history
kubectl rollout history deployment/api-managementz -n api-managementz

# Rollback to previous version
kubectl rollout undo deployment/api-managementz -n api-managementz

# Rollback to specific revision
kubectl rollout undo deployment/api-managementz --to-revision=2 -n api-managementz

# Edit deployment directly
kubectl edit deployment api-managementz -n api-managementz

# Patch deployment
kubectl patch deployment api-managementz -n api-managementz -p '{"spec":{"replicas":2}}'
```

## Service Commands

```bash
# List services
kubectl get svc -n api-managementz

# Get service details
kubectl describe svc api-managementz -n api-managementz

# Get LoadBalancer IP/hostname
kubectl get svc api-managementz -n api-managementz -o jsonpath='{.status.loadBalancer.ingress[0]}'

# Port-forward to service
kubectl port-forward svc/api-managementz 8080:80 -n api-managementz

# Port-forward to pod
kubectl port-forward pod/<pod-name> 8080:8080 -n api-managementz
```

## Secrets & ConfigMap Commands

```bash
# List secrets
kubectl get secrets -n api-managementz

# Get secret value (base64-encoded)
kubectl get secret api-managementz-secrets -n api-managementz -o jsonpath='{.data.AUTH_SECRET}'

# Decode secret value
kubectl get secret api-managementz-secrets -n api-managementz \
  -o jsonpath='{.data.DB_PASSWORD}' | base64 -d

# Update secret
kubectl create secret generic api-managementz-secrets \
  --from-literal=AUTH_SECRET=new-secret \
  --dry-run=client -o yaml | kubectl apply -f -

# View ConfigMap
kubectl get configmap api-managementz-config -n api-managementz -o yaml

# Edit ConfigMap
kubectl edit configmap api-managementz-config -n api-managementz
```

## Storage Commands

```bash
# List persistent volumes
kubectl get pv

# List persistent volume claims
kubectl get pvc -n api-managementz

# Get PVC details
kubectl describe pvc postgres-pvc -n api-managementz

# Check PVC usage
kubectl exec -it deployment/postgres -n api-managementz -- df -h /var/lib/postgresql/data
```

## Database Commands

```bash
# Get Postgres pod name
POSTGRES_POD=$(kubectl get pods -n api-managementz -l app=postgres -o jsonpath='{.items[0].metadata.name}')

# Connect to Postgres
kubectl exec -it $POSTGRES_POD -n api-managementz -- psql -U api_managementz -d api_managementz

# Run SQL query
kubectl exec $POSTGRES_POD -n api-managementz -- \
  psql -U api_managementz -d api_managementz -c "SELECT * FROM users;"

# Backup database
kubectl exec $POSTGRES_POD -n api-managementz -- \
  pg_dump -U api_managementz -d api_managementz > backup.sql

# Restore database
kubectl exec -i $POSTGRES_POD -n api-managementz -- \
  psql -U api_managementz -d api_managementz < backup.sql
```

## Redis Commands

```bash
# Get Redis pod name
REDIS_POD=$(kubectl get pods -n api-managementz -l app=redis -o jsonpath='{.items[0].metadata.name}')

# Connect to Redis
kubectl exec -it $REDIS_POD -n api-managementz -- redis-cli

# Run Redis command
kubectl exec $REDIS_POD -n api-managementz -- redis-cli INFO

# Check Redis memory
kubectl exec $REDIS_POD -n api-managementz -- redis-cli INFO memory

# Flush Redis cache
kubectl exec $REDIS_POD -n api-managementz -- redis-cli FLUSHALL
```

## Troubleshooting Commands

```bash
# Get pod events (why pod failed to start)
kubectl describe pod <pod-name> -n api-managementz

# Get namespace events
kubectl get events -n api-managementz --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n api-managementz
kubectl top nodes

# Check node status
kubectl get nodes

# Describe node
kubectl describe node <node-name>

# Check API server logs
kubectl logs -n kube-system deployment/kube-apiserver

# Check kubelet logs
journalctl -u kubelet -f

# Test DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nslookup postgres.api-managementz.svc.cluster.local

# Test service connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://api-managementz.api-managementz.svc.cluster.local/health
```

## Monitoring & Logs

```bash
# Get previous logs (if pod crashed and restarted)
kubectl logs <pod-name> --previous -n api-managementz

# Get logs from specific container
kubectl logs <pod-name> -c postgres -n api-managementz

# Stream logs from multiple pods
kubectl logs -l app=api-managementz -f -n api-managementz

# Get logs with timestamps
kubectl logs <pod-name> --timestamps=true -n api-managementz

# Get logs since 30 minutes ago
kubectl logs <pod-name> --since=30m -n api-managementz

# Get logs for pod that's in CrashLoopBackOff
kubectl logs <pod-name> --tail=100 -n api-managementz
```

## Common Issues & Fixes

### Pod stuck in Pending

```bash
# Check events
kubectl describe pod <pod-name> -n api-managementz

# Check node resources
kubectl top nodes

# Check PVC status
kubectl describe pvc postgres-pvc -n api-managementz
```

### Pod in CrashLoopBackOff

```bash
# Get logs from previous run
kubectl logs <pod-name> --previous -n api-managementz

# Check probe configuration
kubectl get pod <pod-name> -n api-managementz -o yaml | grep -A 10 probe

# Increase initial delay
kubectl patch deployment <deployment> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container>","readinessProbe":{"initialDelaySeconds":30}}]}}}}'
```

### ImagePullBackOff

```bash
# For local clusters, load image
kind load docker-image ghcr.io/owner/api-managementz:tag

# Or change pull policy
kubectl patch deployment api-managementz -n api-managementz \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"api-managementz","imagePullPolicy":"Never"}]}}}}'
```

### Service has no endpoints

```bash
# Check if pods are running
kubectl get pods -l app=api-managementz -n api-managementz

# Check service selector
kubectl get svc api-managementz -n api-managementz -o yaml | grep selector

# Check labels on pods
kubectl get pods -n api-managementz --show-labels
```

### High CPU/Memory usage

```bash
# Check resource usage
kubectl top pods -n api-managementz --sort-by=memory

# View pod resource requests/limits
kubectl describe pod <pod-name> -n api-managementz | grep -A 5 "Limits\|Requests"

# Check deployment limits
kubectl get deployment api-managementz -n api-managementz -o yaml | grep -A 10 resources
```

## Useful Aliases

Add to `~/.bashrc` or `~/.zshrc`:

```bash
alias k='kubectl'
alias kn='kubectl get nodes'
alias kp='kubectl get pods -n api-managementz'
alias kd='kubectl get deployment -n api-managementz'
alias ks='kubectl get svc -n api-managementz'
alias kdesc='kubectl describe'
alias klogs='kubectl logs'
alias kex='kubectl exec -it'
alias kpf='kubectl port-forward'
alias kctx='kubectl config current-context'
```

Usage:
```bash
k get all -n api-managementz
kp                                      # Lists pods in api-managementz
kdesc pod/<pod-name> -n api-managementz
klogs -f deployment/api-managementz -n api-managementz
```

## One-Liners

```bash
# Restart all pods in deployment
kubectl rollout restart deployment/api-managementz -n api-managementz

# Delete all pods (will be recreated)
kubectl delete pods --all -n api-managementz

# Get all resources in YAML format
kubectl get all -n api-managementz -o yaml

# Find pods with errors
kubectl get pods -n api-managementz --field-selector=status.phase!=Running

# Count pods by status
kubectl get pods -n api-managementz -o jsonpath='{range .items[*]}{.status.phase}{"\n"}{end}' | sort | uniq -c

# Get container images used
kubectl get pods -n api-managementz -o jsonpath='{range .items[*]}{.spec.containers[*].image}{"\n"}{end}'

# Get all environment variables from pod
kubectl exec <pod-name> -n api-managementz -- env
```
