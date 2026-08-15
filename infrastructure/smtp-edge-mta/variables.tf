variable "aws_profile" {
  type = string
}
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "environment" {
  type    = string
  default = "production"
}

variable "admin_cidr" {
  type        = string
  description = "Allowed IP range for administrative access"
}