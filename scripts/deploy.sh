#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# deploy.sh
#
# Manual helper for building, pushing to ECR and deploying CloudTask to
# an existing EKS cluster. Useful for local testing of the same steps
# the CI/CD pipelines automate.
#
# Usage:
#   ./scripts/deploy.sh <aws-account-id> <aws-region> <image-tag>
#
# Example:
#   ./scripts/deploy.sh 123456789012 ap-south-1 v1.0.3
# ---------------------------------------------------------------------------
set -euo pipefail

AWS_ACCOUNT_ID="${1:?Usage: deploy.sh <aws-account-id> <aws-region> <image-tag>}"
AWS_REGION="${2:?Usage: deploy.sh <aws-account-id> <aws-region> <image-tag>}"
IMAGE_TAG="${3:-latest}"

BACKEND_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/cloudtask-backend"
FRONTEND_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/cloudtask-frontend"
CLUSTER_NAME="cloudtask-eks"

echo "==> Logging in to Amazon ECR"
aws ecr get-login-password --region "$AWS_REGION" | \
  docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

echo "==> Building images"
docker build -t "${BACKEND_REPO}:${IMAGE_TAG}" ./backend
docker build -t "${FRONTEND_REPO}:${IMAGE_TAG}" ./frontend

echo "==> Pushing images"
docker push "${BACKEND_REPO}:${IMAGE_TAG}"
docker push "${FRONTEND_REPO}:${IMAGE_TAG}"

echo "==> Updating kubeconfig"
aws eks update-kubeconfig --region "$AWS_REGION" --name "$CLUSTER_NAME"

echo "==> Applying Kubernetes manifests"
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/backend-service.yaml
kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/frontend-service.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/ingress.yaml

echo "==> Rolling out new image versions"
kubectl set image deployment/cloudtask-backend backend="${BACKEND_REPO}:${IMAGE_TAG}" -n cloudtask
kubectl set image deployment/cloudtask-frontend frontend="${FRONTEND_REPO}:${IMAGE_TAG}" -n cloudtask

kubectl rollout status deployment/cloudtask-backend -n cloudtask --timeout=120s
kubectl rollout status deployment/cloudtask-frontend -n cloudtask --timeout=120s

echo "==> Deployment complete!"
kubectl get pods -n cloudtask
