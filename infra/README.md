# Infrastructure as Code - Core Banking Lab

## Architecture Overview

### Deployment Flow
```
GitHub Actions → Terraform → Ansible → k3s Cluster → Application
```

### Components

1. **Terraform** (`infra/terraform/`)
   - EC2 t4g.small instance (ARM-based, cost-effective)
   - VPC with public subnet
   - Security Groups (SSH, HTTP, HTTPS, k3s)
   - Key pair for SSH access
   - IAM roles and policies
   - Route53 (optional domain setup)

2. **Ansible** (`infra/ansible/`)
   - EC2 instance provisioning
   - k3s cluster installation
   - Docker registry setup
   - Application deployment
   - Monitoring stack deployment

3. **GitHub Actions** (`ci-cd/`)
   - Build and push Docker images
   - Terraform plan/apply
   - Ansible deployment
   - Environment management (dev/staging/prod)

### AWS Resources

- **EC2**: t4g.small (2 vCPU, 2GB RAM, ARM64)
- **Storage**: 20GB gp3 EBS volume
- **Network**: VPC, Subnet, Internet Gateway, Security Groups
- **Domain**: Optional Route53 hosted zone

### Security

- SSH key-based access
- Security groups with minimal required ports
- IAM roles with least privilege
- Secrets management via GitHub Secrets
- TLS/SSL certificates via Let's Encrypt

### Cost Estimation

- EC2 t4g.small: ~$12/month
- EBS 20GB gp3: ~$2/month
- Data transfer: ~$1-5/month
- **Total**: ~$15-20/month

## Quick Start

1. **Prerequisites**
   ```bash
   # Install required tools
   brew install terraform ansible awscli
   
   # Configure AWS credentials
   aws configure
   ```

2. **Deploy Infrastructure**
   ```bash
   cd infra/terraform
   terraform init
   terraform plan
   terraform apply
   ```

3. **Deploy Application**
   ```bash
   cd ../ansible
   ansible-playbook -i inventory/aws_ec2.yml playbooks/deploy.yml
   ```

4. **CI/CD Setup**
   - Configure GitHub Secrets
   - Push to trigger deployment

## Environment Management

- **Development**: Single EC2 instance
- **Staging**: Separate EC2 instance  
- **Production**: Enhanced security + monitoring

## Monitoring & Backup

- Prometheus/Grafana deployed via k3s
- CloudWatch integration
- Automated EBS snapshots
- Log aggregation