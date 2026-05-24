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

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

module "ecr" {
  source = "./modules/ecr"
}

module "security_group" {
  source = "./modules/security_group"
  vpc_id = data.aws_vpc.default.id
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
  subnets         = data.aws_subnets.default.ids
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