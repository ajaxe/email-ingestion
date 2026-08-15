terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {} # Values supplied dynamically via backend config file
}

provider "aws" {
  region  = var.aws_region
  profile = "${var.aws_profile}"
}

# --- VPC & Subnets ---
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# --- AMI: Amazon Linux 2023 (ARM64) ---
data "aws_ami" "al2023_arm64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-arm64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# --- IAM Role for AWS Systems Manager (SSM Session Manager) ---
resource "aws_iam_role" "smtp_edge_role" {
  name = "email-ingestion-edge-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.smtp_edge_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "smtp_edge_profile" {
  name = "email-ingestion-edge-profile-${var.environment}"
  role = aws_iam_role.smtp_edge_role.name
}

# --- Security Group ---
resource "aws_security_group" "smtp_edge_sg" {
  name        = "email-ingestion-edge-sg-${var.environment}"
  description = "Security group for Edge SMTP receiver"
  vpc_id      = data.aws_vpc.default.id

  # Inbound SMTP (Port 25)
  ingress {
    description = "Inbound SMTP from any MTA"
    from_port   = 25
    to_port     = 25
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Inbound Submission / TLS (Port 587 - optional)
  ingress {
    description = "Inbound submission/TLS"
    from_port   = 587
    to_port     = 587
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH Access (Optional, prefer AWS SSM)
  ingress {
    description = "SSH Admin Access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.admin_cidr]
  }

  # Outbound Egress (DNS, HTTPS, External Redis / DB)
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "email-ingestion-edge-sg"
    Environment = var.environment
  }
}

# --- EC2 Instance (t4g.nano Graviton ARM64) ---
resource "aws_instance" "smtp_edge" {
  ami                  = data.aws_ami.al2023_arm64.id
  instance_type        = "t4g.nano"
  subnet_id            = data.aws_subnets.default.ids[0]
  iam_instance_profile = aws_iam_instance_profile.smtp_edge_profile.name

  vpc_security_group_ids = [aws_security_group.smtp_edge_sg.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 8
    delete_on_termination = true
    encrypted             = true

    tags = {
      Name = "email-ingestion-edge-root"
    }
  }

  user_data = <<-EOF
              #!/bin/bash
              set -e
              # Setup 1GB swap safety on 512MB RAM instance
              fallocate -l 1G /swapfile
              chmod 600 /swapfile
              mkswap /swapfile
              swapon /swapfile
              echo '/swapfile none swap sw 0 0' >> /etc/fstab

              dnf update -y
              EOF

  tags = {
    Name        = "email-ingestion-edge-node"
    Environment = var.environment
  }
}

# --- Elastic IP Allocation & Attachment ---
resource "aws_eip" "smtp_eip" {
  domain   = "vpc"
  instance = aws_instance.smtp_edge.id

  tags = {
    Name        = "email-ingestion-edge-eip"
    Environment = var.environment
  }

  depends_on = [aws_instance.smtp_edge]
}

