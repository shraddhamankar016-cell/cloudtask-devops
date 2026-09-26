#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# bootstrap-tf-backend.sh
#
# Creates the S3 bucket + DynamoDB lock table that Terraform's "s3" backend
# needs for remote state storage and state locking. Run this ONCE before
# uncommenting the backend "s3" block in terraform/provider.tf.
#
# Usage:
#   ./scripts/bootstrap-tf-backend.sh <bucket-name> <aws-region>
# ---------------------------------------------------------------------------
set -euo pipefail

BUCKET_NAME="${1:?Usage: bootstrap-tf-backend.sh <bucket-name> <aws-region>}"
AWS_REGION="${2:?Usage: bootstrap-tf-backend.sh <bucket-name> <aws-region>}"
LOCK_TABLE="cloudtask-terraform-locks"

echo "==> Creating S3 bucket for Terraform state: $BUCKET_NAME"
aws s3api create-bucket \
  --bucket "$BUCKET_NAME" \
  --region "$AWS_REGION" \
  --create-bucket-configuration LocationConstraint="$AWS_REGION" 2>/dev/null || \
  echo "Bucket may already exist, continuing..."

aws s3api put-bucket-versioning \
  --bucket "$BUCKET_NAME" \
  --versioning-configuration Status=Enabled

aws s3api put-bucket-encryption \
  --bucket "$BUCKET_NAME" \
  --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

echo "==> Creating DynamoDB table for state locking: $LOCK_TABLE"
aws dynamodb create-table \
  --table-name "$LOCK_TABLE" \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$AWS_REGION" 2>/dev/null || \
  echo "Table may already exist, continuing..."

echo "==> Done. Now uncomment the backend \"s3\" block in terraform/provider.tf with:"
echo "    bucket         = \"$BUCKET_NAME\""
echo "    region         = \"$AWS_REGION\""
echo "    dynamodb_table = \"$LOCK_TABLE\""
