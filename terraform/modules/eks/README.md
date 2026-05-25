# EKS Module

This module creates:

- `aws_eks_cluster`
- `aws_eks_node_group`
- IAM role for the EKS control plane
- IAM role for the managed node group
- Managed policy attachments for cluster and node group permissions

## Example

```hcl
module "eks" {
  source = "./modules/eks"

  cluster_name        = "api-managementz-eks"
  subnet_ids          = module.vpc.private_subnet_ids
  node_group_name     = "api-managementz-eks-nodes"
  kubernetes_version  = "1.29"
  instance_types      = ["t3.medium"]
  desired_capacity    = 2
  min_size            = 1
  max_size            = 3
  capacity_type       = "ON_DEMAND"
  disk_size           = 20
  cluster_role_name   = "api-managementz-eks-cluster-role"
  node_role_name      = "api-managementz-eks-node-role"
  tags = {
    Environment = "prod"
    Project     = "api-managementz"
  }
}
```
