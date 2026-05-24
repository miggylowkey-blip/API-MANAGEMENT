variable "cluster_name" {
  type    = string
  default = "api-managementz-cluster"
}

variable "service_name" {
  type    = string
  default = "api-managementz-service"
}

variable "task_family" {
  type    = string
  default = "api-managementz-task"
}

variable "container_name" {
  type    = string
  default = "api-managementz"
}

variable "container_image" {
  type = string
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "cpu" {
  type    = number
  default = 256
}

variable "memory" {
  type    = number
  default = 512
}

variable "subnets" {
  type = list(string)
}

variable "security_groups" {
  type = list(string)
}

variable "awslogs_region" {
  type = string
}

variable "environment" {
  type = map(string)
  default = {}
}

variable "task_execution_role_name" {
  type    = string
  default = "api-managementz-ecs-execution-role"
}

variable "log_group_name" {
  type    = string
  default = "/ecs/api-managementz-task"
}
