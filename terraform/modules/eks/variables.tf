variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
}

variable "cluster_subnet_ids" {
  type        = list(string)
  description = "Subnet IDs for the EKS cluster — public and private"
}

variable "node_group_subnet_ids" {
  type        = list(string)
  description = "Subnet IDs for worker nodes — private only"
}

variable "node_group_name" {
  type        = string
  default     = "default-node-group"
  description = "Name of the EKS managed node group"
}

variable "cluster_role_name" {
  type        = string
  default     = "eks-cluster-role"
  description = "IAM role name for the EKS cluster control plane"
}

variable "node_role_name" {
  type        = string
  default     = "eks-node-group-role"
  description = "IAM role name for the EKS managed node group"
}

variable "kubernetes_version" {
  type        = string
  default     = "1.29"
  description = "Kubernetes version for the EKS cluster"
}

variable "instance_types" {
  type        = list(string)
  default     = ["t3.medium"]
  description = "EC2 instance types for the managed node group"
}

variable "desired_capacity" {
  type        = number
  default     = 2
  description = "Desired number of nodes in the managed node group"
}

variable "min_size" {
  type        = number
  default     = 1
  description = "Minimum number of nodes in the managed node group"
}

variable "max_size" {
  type        = number
  default     = 3
  description = "Maximum number of nodes in the managed node group"
}

variable "disk_size" {
  type        = number
  default     = 20
  description = "Disk size for worker nodes in GiB"
}

variable "capacity_type" {
  type        = string
  default     = "ON_DEMAND"
  description = "Capacity type for the managed node group (ON_DEMAND or SPOT)"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags to apply to EKS resources"
}
