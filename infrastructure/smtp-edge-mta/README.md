# IaC for SMTP MAT at Edge

Terraform configuration for provisioning SMTP server at the edge using AWS EC2 and Route53. This setup is designed for handling email traffic efficiently and securely.

## Configuration Setup

1. **Terraform Variables**: Set up the necessary variables in `terraform.tfvars` such as AWS region, environment, and admin CIDR block. Setup vairables in `backend.tfvars` to setup S3 backend for state management.

## Terraform instructions

```pwsh
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

## Remote Connection

EC2 instance has been setup with SSH key, we can login using it as follows:

```pwsh
ssh -i ~/.ssh/email_ingestion_mta_ec2 ec2-user@$(terraform output -raw public_ip)
```

## Initial Server Configuration

One time server configuration when a new EC2 instance is created.

```bash
# 1. Create a dedicated system user
sudo useradd --system --no-create-home --shell /sbin/nologin email-ingest

# 2. Create the unified directory tree
sudo mkdir -p /opt/email-ingest/logs

# 3. Restrict permissions to the service user
sudo chown -R email-ingest:email-ingest /opt/email-ingest
sudo chmod 700 /opt/email-ingest
```
