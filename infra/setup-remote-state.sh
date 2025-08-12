#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Setting up Terraform remote state backend...${NC}"

# Configuration
BUCKET_NAME="core-banking-lab-terraform-state-718277288108"
DYNAMODB_TABLE="core-banking-lab-terraform-locks"
REGION="us-east-1"

# Create S3 bucket for state storage
echo -e "${YELLOW}Creating S3 bucket: $BUCKET_NAME${NC}"
if aws s3 mb "s3://$BUCKET_NAME" --region "$REGION" 2>/dev/null; then
    echo -e "${GREEN}✅ S3 bucket created successfully${NC}"
else
    echo -e "${YELLOW}⚠️  S3 bucket may already exist or insufficient permissions${NC}"
fi

# Enable versioning on the bucket
echo -e "${YELLOW}Enabling versioning on S3 bucket...${NC}"
aws s3api put-bucket-versioning \
    --bucket "$BUCKET_NAME" \
    --versioning-configuration Status=Enabled \
    2>/dev/null && echo -e "${GREEN}✅ Versioning enabled${NC}" || echo -e "${YELLOW}⚠️  Could not enable versioning${NC}"

# Enable encryption
echo -e "${YELLOW}Enabling encryption on S3 bucket...${NC}"
aws s3api put-bucket-encryption \
    --bucket "$BUCKET_NAME" \
    --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}' \
    2>/dev/null && echo -e "${GREEN}✅ Encryption enabled${NC}" || echo -e "${YELLOW}⚠️  Could not enable encryption${NC}"

# Block public access
echo -e "${YELLOW}Blocking public access on S3 bucket...${NC}"
aws s3api put-public-access-block \
    --bucket "$BUCKET_NAME" \
    --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
    2>/dev/null && echo -e "${GREEN}✅ Public access blocked${NC}" || echo -e "${YELLOW}⚠️  Could not block public access${NC}"

# Create DynamoDB table for state locking
echo -e "${YELLOW}Creating DynamoDB table: $DYNAMODB_TABLE${NC}"
if aws dynamodb create-table \
    --table-name "$DYNAMODB_TABLE" \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region "$REGION" 2>/dev/null; then
    echo -e "${GREEN}✅ DynamoDB table created successfully${NC}"
    
    # Wait for table to be active
    echo -e "${YELLOW}Waiting for DynamoDB table to become active...${NC}"
    aws dynamodb wait table-exists --table-name "$DYNAMODB_TABLE" --region "$REGION"
    echo -e "${GREEN}✅ DynamoDB table is active${NC}"
else
    echo -e "${YELLOW}⚠️  DynamoDB table may already exist or insufficient permissions${NC}"
fi

echo -e "${GREEN}🎉 Remote state backend setup complete!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Run 'terraform init' to initialize the remote backend"
echo "2. Migrate your existing state with 'terraform init -migrate-state'"
echo ""
echo -e "${YELLOW}Backend configuration:${NC}"
echo "  Bucket: $BUCKET_NAME"
echo "  DynamoDB Table: $DYNAMODB_TABLE"
echo "  Region: $REGION"