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

Using AWS CLI, start SSM sessions

```pwsh
aws ssm start-session --target <instance_id_from_output>
```