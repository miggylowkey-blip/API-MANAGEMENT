# This directory contains reusable Terraform modules for AWS resources.
#
# Modules are organized into subdirectories:
#   - ./modules/ecr
#   - ./modules/rds
#   - ./modules/ecs
#   - ./modules/security_group
#
# Use these modules from the root Terraform configuration with:
#   module "name" {
#     source = "./modules/<module_name>"
#     ...
#   }
