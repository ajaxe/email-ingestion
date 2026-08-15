# --- Outputs ---
output "instance_id" {
  description = "EC2 Instance ID"
  value       = aws_instance.smtp_edge.id
}

output "public_ip" {
  description = "Static Elastic IPv4 address (Point your DNS MX / A records here)"
  value       = aws_eip.smtp_eip.public_ip
}
