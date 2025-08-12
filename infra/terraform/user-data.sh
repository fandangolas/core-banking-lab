#!/bin/bash

# Update system
apt-get update
apt-get upgrade -y

# Set hostname
hostnamectl set-hostname ${hostname}

# Install required packages
apt-get install -y \
    curl \
    wget \
    git \
    htop \
    vim \
    unzip \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release

# Create ubuntu user sudo access
usermod -aG sudo ubuntu

# Configure automatic security updates
echo 'Unattended-Upgrade::Automatic-Reboot "false";' >> /etc/apt/apt.conf.d/50unattended-upgrades
systemctl enable unattended-upgrades

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=arm64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io
usermod -aG docker ubuntu

# Install k3s (will be done via Ansible, but prepare)
# curl -sfL https://get.k3s.io | sh -

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Configure firewall (UFW)
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 6443/tcp  # k3s API
ufw allow 8080/tcp  # Banking API
ufw allow 3000/tcp  # Dashboard
ufw allow 3001/tcp  # Grafana
ufw allow 9090/tcp  # Prometheus

# Create directories for applications
mkdir -p /opt/banking-app
chown ubuntu:ubuntu /opt/banking-app

# Enable services
systemctl enable docker
systemctl start docker

# Log completion
echo "$(date): User data script completed" >> /var/log/user-data.log