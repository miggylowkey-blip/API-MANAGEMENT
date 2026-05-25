terraform {
  required_version = ">=1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket  = "api-managementz-tf-state"
    key     = "terraform/terraform.tfstate"
    region  = "ap-southeast-2"
    encrypt = true
  }
}

provider "aws" {
  region = var.region
}

module "vpc" {
  source = "./modules/vpc"
  name   = "api-managementz-vpc"
  cidr_block           = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24"]
  enable_nat_gateway   = true
  eks_cluster_name     = "api-managementzzz"
}

module "ecr" {
  source = "./modules/ecr"
}

module "security_group" {
  source = "./modules/security_group"
  vpc_id = module.vpc.vpc_id
}

module "rds" {
  source                 = "./modules/rds"
  password               = var.db_password
  vpc_security_group_ids = [module.security_group.id]
}

module "ecs" {
  source          = "./modules/ecs"
  container_image = "${module.ecr.repository_url}:latest"
  awslogs_region  = var.region
  subnets         = module.vpc.private_subnet_ids
  security_groups = [module.security_group.id]
  environment = {
    PORT             = "8080"
    DB_HOST          = module.rds.address
    DB_USER          = "api_managementz"
    DB_PASSWORD      = var.db_password
    DB_NAME          = "postgres"
    DB_SSLMODE       = "require"
    AUTH_SECRET      = var.auth_secret
    RATE_LIMIT_RPS   = "5"
    RATE_LIMIT_BURST = "10"
    REDIS_URL        = "redis://localhost:6379"
  }
}

module "eks" {
  source = "./modules/eks"
  cluster_name = "api-managementzzz"
  cluster_subnet_ids = concat(module.vpc.public_subnet_ids, module.vpc.private_subnet_ids)
  node_group_subnet_ids = module.vpc.private_subnet_ids
  kubernetes_version = "1.29"
  instance_types = ["t3.medium"]
  desired_capacity = 2
  min_size = 1
  max_size = 3

} 