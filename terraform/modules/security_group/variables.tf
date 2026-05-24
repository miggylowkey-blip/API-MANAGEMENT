variable "name" {
  type    = string
  default = "api-managementz-sg"
}

variable "description" {
  type    = string
  default = "Security group for API-MANAGEMENTZ"
}

variable "vpc_id" {
  type = string
}

variable "ingress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_blocks = list(string)
  }))
  default = [
    { from_port = 8080, to_port = 8080, protocol = "tcp", cidr_blocks = ["0.0.0.0/0"] },
    { from_port = 5432, to_port = 5432, protocol = "tcp", cidr_blocks = ["0.0.0.0/0"] },
  ]
}

variable "egress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_blocks = list(string)
  }))
  default = [
    { from_port = 0, to_port = 0, protocol = "-1", cidr_blocks = ["0.0.0.0/0"] },
  ]
}
