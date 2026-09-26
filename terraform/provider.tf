terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Remote state — uncomment and configure once the S3 bucket + DynamoDB
  # lock table exist (see scripts/bootstrap-tf-backend.sh).
  #
  # backend "s3" {
  #   bucket         = "cloudtask-terraform-state"
  #   key            = "cloudtask/terraform.tfstate"
  #   region         = "ap-south-1"
  #   dynamodb_table = "cloudtask-terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "CloudTask"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}
