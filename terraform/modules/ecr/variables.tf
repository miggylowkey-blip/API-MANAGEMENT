variable "repository_name" {
  type    = string
  default = "api-management/api"
}

variable "image_tag_mutability" {
  type    = string
  default = "MUTABLE"
}

variable "scan_on_push" {
  type    = bool
  default = true
}
