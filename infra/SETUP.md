# Infrastructure Setup Guide

## Prerequisites

1. **AWS Account** with programmatic access
2. **GitHub Repository** with Actions enabled
3. **SSH Key Pair** for EC2 access
4. **Domain** (optional) for custom URLs

## Step 1: Generate SSH Key Pair

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f ~/.ssh/core-banking-lab-key -C "core-banking-lab"

# Your public key (add to GitHub Secrets)
cat ~/.ssh/core-banking-lab-key.pub

# Your private key (add to GitHub Secrets)
cat ~/.ssh/core-banking-lab-key
```

## Step 2: Configure GitHub Secrets

Go to your repository → Settings → Secrets and variables → Actions

Add these secrets:

### AWS Credentials
```
AWS_ACCESS_KEY_ID: <your-aws-access-key>
AWS_SECRET_ACCESS_KEY: <your-aws-secret-key>
```

### SSH Keys
```
SSH_PUBLIC_KEY: <content-of-core-banking-lab-key.pub>
SSH_PRIVATE_KEY: <content-of-core-banking-lab-key>
```

## Step 3: Configure AWS IAM User

Create an IAM user with these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "vpc:*",
        "iam:CreateRole",
        "iam:AttachRolePolicy",
        "iam:CreateInstanceProfile",
        "iam:AddRoleToInstanceProfile",
        "iam:PassRole",
        "iam:GetRole",
        "iam:GetInstanceProfile",
        "iam:ListAttachedRolePolicies",
        "iam:TagRole",
        "route53:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## Step 4: Manual Deployment (Optional)

### Local Terraform Deployment

```bash
# Clone repository
git clone <your-repo>
cd core-banking-lab/infra/terraform

# Copy and configure variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

# Initialize and deploy
terraform init
terraform plan
terraform apply
```

### Local Ansible Deployment

```bash
cd ../ansible

# Install requirements
pip install ansible boto3 botocore kubernetes
ansible-galaxy collection install -r requirements.yml

# Configure AWS credentials
export AWS_ACCESS_KEY_ID=<your-key>
export AWS_SECRET_ACCESS_KEY=<your-secret>

# Deploy application
ansible-playbook -i inventory/aws_ec2.yml playbooks/deploy.yml
```

## Step 5: Automated Deployment via GitHub Actions

### Deploy to Development
```bash
# Push to main branch triggers automatic deployment
git push origin main
```

### Manual Deployment
1. Go to Actions tab in GitHub
2. Select "Deploy to AWS" workflow
3. Click "Run workflow"
4. Choose environment and action
5. Click "Run workflow"

## Step 6: Verify Deployment

After successful deployment, check these URLs:

- **Banking API**: `http://<instance-ip>:8080/prometheus`
- **Dashboard**: `http://<instance-ip>:3000`
- **Prometheus**: `http://<instance-ip>:9090`
- **Grafana**: `http://<instance-ip>:3001` (admin/admin123)

## Cost Management

### Estimated Monthly Costs
- EC2 t4g.small: ~$12
- EBS 20GB gp3: ~$2
- Data transfer: ~$1-5
- **Total**: ~$15-20/month

### Cost Optimization
```bash
# Stop instance when not needed
aws ec2 stop-instances --instance-ids <instance-id>

# Start instance when needed
aws ec2 start-instances --instance-ids <instance-id>

# Terminate everything
terraform destroy
```

## Troubleshooting

### SSH Connection Issues
```bash
# Check security group rules
aws ec2 describe-security-groups --group-ids <sg-id>

# Test SSH connection
ssh -i ~/.ssh/core-banking-lab-key ubuntu@<instance-ip>
```

### Kubernetes Issues
```bash
# Check k3s status
sudo systemctl status k3s

# Check pods
kubectl get pods -n banking

# Check logs
kubectl logs -n banking deployment/banking-api
```

### GitHub Actions Issues
- Check AWS credentials are correct
- Verify SSH keys are properly formatted
- Check instance is running and accessible
- Review workflow logs for specific errors

## Security Best Practices

1. **Rotate SSH keys** regularly
2. **Use least privilege** IAM policies
3. **Enable CloudTrail** for audit logging
4. **Set up billing alerts**
5. **Use HTTPS** for production deployments
6. **Regular security updates** via user-data script

## Next Steps

1. **Domain Setup**: Configure Route53 for custom domain
2. **HTTPS/TLS**: Add Let's Encrypt certificates
3. **Monitoring**: Set up CloudWatch alarms
4. **Backup**: Configure EBS snapshots
5. **High Availability**: Multi-AZ deployment
6. **CI/CD**: Add testing and staging environments