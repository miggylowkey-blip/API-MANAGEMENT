variable "region" {
  default = "ap-southeast-2"
}

variable "db_password" {
  sensitive = true
}

variable "auth_secret" {
  sensitive = true
}