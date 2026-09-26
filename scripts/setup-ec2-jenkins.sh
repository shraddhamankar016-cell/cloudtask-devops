#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# setup-ec2-jenkins.sh
#
# Provisions a fresh Amazon Linux 2023 / Ubuntu 22.04 EC2 instance as a
# self-hosted Jenkins CI/CD server with Docker, AWS CLI v2, kubectl and
# Terraform installed. Run this once via SSH (or as EC2 user-data) on a
# new instance.
#
# Usage:
#   chmod +x setup-ec2-jenkins.sh
#   sudo ./setup-ec2-jenkins.sh
# ---------------------------------------------------------------------------
set -euo pipefail

echo "==> Updating system packages"
if command -v apt-get >/dev/null 2>&1; then
  PKG_MANAGER="apt-get"
  sudo apt-get update -y
  sudo apt-get upgrade -y
else
  PKG_MANAGER="dnf"
  sudo dnf update -y
fi

echo "==> Installing base utilities"
if [ "$PKG_MANAGER" = "apt-get" ]; then
  sudo apt-get install -y curl unzip git ca-certificates gnupg lsb-release
else
  sudo dnf install -y curl unzip git
fi

echo "==> Installing Docker"
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker "$(whoami)"
sudo systemctl enable docker --now
rm -f get-docker.sh

echo "==> Installing AWS CLI v2"
curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -q awscliv2.zip
sudo ./aws/install --update
rm -rf awscliv2.zip aws

echo "==> Installing kubectl"
KUBECTL_VERSION=$(curl -Ls https://dl.k8s.io/release/stable.txt)
curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
rm -f kubectl

echo "==> Installing Go"
GO_VERSION="1.22.2"
curl -LO "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
rm -f "go${GO_VERSION}.linux-amd64.tar.gz"
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin

echo "==> Installing Terraform"
TERRAFORM_VERSION="1.7.5"
curl -LO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
unzip -q "terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
sudo mv terraform /usr/local/bin/
rm -f "terraform_${TERRAFORM_VERSION}_linux_amd64.zip"

echo "==> Installing Java (required by Jenkins)"
if [ "$PKG_MANAGER" = "apt-get" ]; then
  sudo apt-get install -y fontconfig openjdk-17-jre
else
  sudo dnf install -y java-17-amazon-corretto
fi

echo "==> Installing Jenkins"
if [ "$PKG_MANAGER" = "apt-get" ]; then
  curl -fsSL https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key | sudo tee \
    /usr/share/keyrings/jenkins-keyring.asc > /dev/null
  echo deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc] \
    https://pkg.jenkins.io/debian-stable binary/ | sudo tee \
    /etc/apt/sources.list.d/jenkins.list > /dev/null
  sudo apt-get update -y
  sudo apt-get install -y jenkins
else
  sudo wget -O /etc/yum.repos.d/jenkins.repo https://pkg.jenkins.io/redhat-stable/jenkins.repo
  sudo rpm --import https://pkg.jenkins.io/redhat-stable/jenkins.io-2023.key
  sudo dnf install -y jenkins
fi

sudo usermod -aG docker jenkins
sudo systemctl enable jenkins --now

echo "==> Done! Jenkins is starting on port 8080."
echo "    Initial admin password:"
sudo cat /var/lib/jenkins/secrets/initialAdminPassword 2>/dev/null || \
  echo "    (not found yet — wait a few seconds and re-run: sudo cat /var/lib/jenkins/secrets/initialAdminPassword)"
