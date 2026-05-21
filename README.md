Kubernetes Security Scanning (Kubesec)
This document details the automated security analysis applied to the API-MANAGEMENTZ Kubernetes manifests. As part of our DevSecOps commitment, no manifest is deployed to production without passing a automated security compliance check.

How it Works in our Pipeline
During the k8s-validate stage of our GitHub Actions workflow, kubesec scans all configuration files matching kubernetes/*.yml.

The utility analyzes the resource definitions against known cloud-native vulnerabilities and calculates an aggregate security score.

Gatekeeping & Failure Thresholds
Score < 5 (Critical Risk): The pipeline forcefully drops with an exit code, pulling the emergency brake and blocking the deployment job.

Score 5–7 (Warning): The build passes, but issues are flagged in the logs as technical debt to be remediated.

Score > 7 (Secure): The manifest meets strict enterprise-grade isolation standards and is cleared for cluster scheduling.

Targeted Security Policies
Kubesec evaluates our microservices (Go API, PostgreSQL, and Redis) against several critical security vectors. Below are the key rules enforced in our manifests:

1. Privilege Escalation Prevention
Rule: AllowPrivilegeEscalation

Risk Context: Containers running with default permissions can leverage parent kernel exploits to gain unauthorized host-level privileges.

Our Standard: Every workload must explicitly declare allowPrivilegeEscalation: false inside its container securityContext.

2. Least Privilege Runtime
Rule: runAsNonRoot

Risk Context: If a container running as root (UID 0) is compromised, the attacker gains root-level capabilities over the underlying cluster node.

Our Standard: Workloads like postgres:16-alpine and redis:7-alpine are restricted from executing as root using runAsNonRoot: true.

3. File System Immutability
Rule: readOnlyRootFilesystem

Risk Context: Malicious scripts or reverse-shells often attempt to download and execute binaries directly inside the container's ephemeral storage layer.

Our Standard: The container root filesystem is locked down with readOnlyRootFilesystem: true. Persistent operations are strictly delegated to dedicated volume mounts (e.g., postgres-pvc).
