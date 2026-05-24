variable "identifier" {
  type    = string
  default = "api-managementz-db"
}

variable "engine" {
  type    = string
  default = "postgres"
}

variable "engine_version" {
  type    = string
  default = "16"
}

variable "instance_class" {
  type    = string
  default = "db.t3.micro"
}

variable "allocated_storage" {
  type    = number
  default = 20
}

variable "db_name" {
  type    = string
  default = "postgres"
}

variable "username" {
  type    = string
  default = "api_managementz"
}

variable "password" {
  type      = string
  sensitive = true
}

variable "vpc_security_group_ids" {
  type    = list(string)
  default = []
}

variable "publicly_accessible" {
  type    = bool
  default = true
}

variable "skip_final_snapshot" {
  type    = bool
  default = true
}
