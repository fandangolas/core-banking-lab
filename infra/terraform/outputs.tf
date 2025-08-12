output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.k3s.id
}

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_eip.k3s.public_ip
}

output "instance_private_ip" {
  description = "Private IP address of the EC2 instance"
  value       = aws_instance.k3s.private_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the EC2 instance"
  value       = aws_instance.k3s.public_dns
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "subnet_id" {
  description = "ID of the public subnet"
  value       = aws_subnet.public.id
}

output "security_group_id" {
  description = "ID of the security group"
  value       = aws_security_group.k3s.id
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh -i ~/.ssh/${var.project_name}-key ubuntu@${aws_eip.k3s.public_ip}"
}

output "application_urls" {
  description = "URLs for accessing the applications"
  value = {
    api        = "http://${aws_eip.k3s.public_ip}:8080"
    dashboard  = "http://${aws_eip.k3s.public_ip}:3000"
    prometheus = "http://${aws_eip.k3s.public_ip}:9090"
    grafana    = "http://${aws_eip.k3s.public_ip}:3001"
  }
}