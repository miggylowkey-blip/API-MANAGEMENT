# Kubernetes Security Policies & Best Practices for API-MANAGEMENTZ

This guide describes security measures implemented in the Kubernetes deployment and recommendations for hardening further.

## Current Security Implementations

### 1. Pod Security Policy

#### Non-root containers
```yaml
# All containers run as non-root
- postgres:16-alpine      # Runs as postgres user
- redis:7-alpine          # Runs as redis user
- api-managementz         # Configured to run as non-root
```

#### Resource limits
```yaml
# All containers have CPU and memory limits
resources:
  requests:
    cpu: "250m"
    memory: "256Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

Prevents resource exhaustion and noisy neighbor issues.

#### Readiness & Liveness probes
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 20
  periodSeconds: 20
```

Ensures traffic is only sent to healthy pods and restarts unhealthy ones.

### 2. Secret Management

Sensitive values are stored in Kubernetes Secrets, not ConfigMaps:

```yaml
stringData:
  AUTH_SECRET: change_me              # Application secret
  DB_USER: api_managementz            # Database user
  DB_PASSWORD: api_managementz        # Database password
  DB_NAME: api_managementz            # Database name
  DB_SSLMODE: disable                 # SSL mode
```

**Best practice:** Change `AUTH_SECRET` and `DB_PASSWORD` before production deployment:

```bash
kubectl create secret generic api-managementz-secrets \
  -n api-managementz \
  --from-literal=AUTH_SECRET=$(openssl rand -base64 32) \
  --from-literal=DB_USER=api_managementz \
  --from-literal=DB_PASSWORD=$(openssl rand -base64 32) \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 3. Namespace Isolation

All resources are deployed to the `api-managementz` namespace:

```bash
# List resources only in this namespace
kubectl get all -n api-managementz

# Prevents cross-namespace interference
# Resources are isolated from other apps
```

### 4. Resource Quotas

ResourceQuota limits namespace resource consumption:

```yaml
ResourceQuota:
  cpu: "1"          # Total CPU limited to 1 core
  memory: 1Gi       # Total memory limited to 1 GB
```

Prevents runaway deployments from consuming entire cluster.

### 5. Persistent Storage

PersistentVolumeClaim ensures data persistence without privileged access:

```yaml
PersistentVolumeClaim:
  accessModes:
    - ReadWriteOnce   # Only one pod can mount at a time
  storage: 5Gi        # Allocated storage size
```

### 6. Service Types

- **Postgres & Redis:** `ClusterIP` (internal only)
- **API:** `LoadBalancer` (external access)

ClusterIP services are not accessible from outside the cluster, protecting databases.

---

## Recommended Enhancements

### 1. Pod Security Standards

Enforce pod security standards using admission controllers:

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: api-managementz-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'MustRunAs'
    seLinuxOptions:
      level: "s0:c123,c456"
  readOnlyRootFilesystem: false
```

**Installation:**

```bash
# Apply the policy
kubectl apply -f - <<EOF
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: api-managementz-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
EOF
```

### 2. Network Policies

Restrict network traffic to only required connections:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-managementz-network-policy
  namespace: api-managementz
spec:
  podSelector:
    matchLabels:
      app: api-managementz
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - protocol: TCP
          port: 6379
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 53  # DNS
        - protocol: UDP
          port: 53
```

**Apply:**
```bash
kubectl apply -f network-policy.yml -n api-managementz
```

**Verification:**
```bash
kubectl get networkpolicies -n api-managementz
```

### 3. RBAC (Role-Based Access Control)

Create service accounts and roles for least privilege access:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-managementz
  namespace: api-managementz
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: api-managementz-role
  namespace: api-managementz
rules:
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: api-managementz-rolebinding
  namespace: api-managementz
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: api-managementz-role
subjects:
  - kind: ServiceAccount
    name: api-managementz
    namespace: api-managementz
```

**Update deployments to use:**
```yaml
spec:
  serviceAccountName: api-managementz
```

### 4. Secrets Encryption at Rest

Enable etcd encryption for secrets:

```yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: <base64-encoded-32-byte-key>
      - identity: {}
```

**For managed clusters (EKS, AKS, GKE):**
- Most cloud providers encrypt secrets by default
- Verify with your provider's documentation

### 5. Image Scanning & Registry Scanning

#### In CI/CD pipeline (already implemented):
- Trivy scans images for vulnerabilities
- SBOM (Software Bill of Materials) is generated
- Results published to GitHub Security tab

#### Additional recommendations:
```bash
# Scan images on schedule
trivy image --severity HIGH,CRITICAL ghcr.io/owner/api-managementz:latest

# Enable vulnerability scanning in registry
# GitHub: Settings > Code security > Dependabot
# ECR: Registries > Enable image scanning
```

### 6. Audit Logging

Enable Kubernetes audit logging to track API access:

```yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: RequestResponse
    omitStages:
      - RequestReceived
    resources:
      - group: ""
        resources: ["secrets"]
    namespaces: ["api-managementz"]
  - level: Metadata
    omitStages:
      - RequestReceived
    resources:
      - group: "apps"
        resources: ["deployments", "statefulsets"]
```

### 7. TLS for Service Communication

Enable TLS between services:

```bash
# Generate certificates
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Create secret
kubectl create secret tls api-managementz-tls \
  --cert=cert.pem \
  --key=key.pem \
  -n api-managementz
```

### 8. Monitoring & Alerting

Monitor security events:

```bash
# Watch for pod security violations
kubectl get events -n api-managementz --sort-by='.lastTimestamp'

# Monitor pod restarts
kubectl get pods -n api-managementz -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'
```

---

## Security Checklist

### Pre-Deployment

- [ ] Change `AUTH_SECRET` to a strong random value
- [ ] Change `DB_PASSWORD` to a strong random value
- [ ] Review and update resource requests/limits
- [ ] Verify all images are scanned and pass security checks
- [ ] Test readiness and liveness probes

### Deployment

- [ ] Create namespace before deployment
- [ ] Apply all Kubernetes manifests
- [ ] Verify no pods in `CrashLoopBackOff` or `Pending` state
- [ ] Test health checks manually
- [ ] Monitor logs for security-related events

### Post-Deployment

- [ ] Review CI/CD security scan results
- [ ] Confirm Network Policies are in place (if using)
- [ ] Enable Pod Security Standards (if using Kubernetes 1.25+)
- [ ] Set up audit logging
- [ ] Configure monitoring and alerting

### Ongoing

- [ ] Review and update base images regularly
- [ ] Run Trivy scans on production images
- [ ] Monitor security advisories
- [ ] Update Kubernetes cluster regularly
- [ ] Review RBAC permissions periodically

---

## Compliance Frameworks

### SLSA (Supply Chain Levels for Software Artifacts)

The CI/CD pipeline provides SLSA compliance:

- ✅ **Build provenance** — GitHub Actions records all build steps
- ✅ **Source integrity** — Git commits are signed
- ✅ **Dependency tracking** — SBOM generated for all images
- ✅ **Access controls** — KUBECONFIG secret access logged

### OWASP Top 10 for Kubernetes

| Issue | Mitigation |
|-------|-----------|
| **Insecure Workload Configuration** | Pod Security Standards, resource limits |
| **Supply Chain Vulnerabilities** | Image scanning (Trivy), dependency tracking |
| **Overly Permissive RBAC** | Least privilege service accounts |
| **Lack of Network Segmentation** | Network Policies |
| **Secrets Management** | Kubernetes Secrets, encryption at rest |
| **Insufficient Logging** | Audit logging enabled |
| **Insecure Container Registry** | Private registry (ghcr.io with auth) |
| **Vulnerable Components** | CI/CD scanning, regular updates |
| **Misconfiguration** | kubesec validation, linting |
| **Lack of Resource Management** | Resource quotas, limits |

---

## Resources

- [Kubernetes Security Documentation](https://kubernetes.io/docs/concepts/security/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [OWASP Kubernetes Top 10](https://kubernetes.io/docs/concepts/security/)
