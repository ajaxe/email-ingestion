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

variable "public_key" {
  type = string
  default = "email_ingestion_mta_ec2.pub"
}