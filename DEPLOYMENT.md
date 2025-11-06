# AAA Service AWS Deployment Guide

This guide walks you through deploying the AAA service to AWS using the CloudFormation template.

## Cost Optimization

The CloudFormation template is optimized for cost efficiency:

- **Single NAT Gateway**: Saves ~$32/month compared to dual NAT Gateway setup
- **Configurable Multi-AZ**: Disable for dev/staging, enable for production
- **Estimated Costs**:
  - Dev/Staging (EnableMultiAZ=false): ~$50-80/month
  - Production (EnableMultiAZ=true): ~$150-200/month

### Cost Breakdown

| Component | Dev (Single-AZ) | Prod (Multi-AZ) |
|-----------|----------------|-----------------|
| ECS Fargate (2 tasks) | $20-30 | $40-60 |
| RDS PostgreSQL (t3.micro) | $15-20 | $30-40 |
| ElastiCache Redis (t3.micro) | $10-15 | $20-30 |
| NAT Gateway | $32 | $32 |
| Application Load Balancer | $20-25 | $20-25 |
| Data Transfer | $5-10 | $10-20 |
| **Total** | **$50-80** | **$150-200** |

## Prerequisites

1. **AWS CLI** installed and configured
2. **Docker** for building the container image
3. **ECR Repository** for storing the Docker image
4. **ACM Certificate** (optional, for HTTPS)

## Step 1: Build and Push Docker Image

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Create ECR repository (if not exists)
aws ecr create-repository --repository-name aaa-service --region us-east-1

# Build the Docker image
docker build -t aaa-service:latest .

# Tag the image
docker tag aaa-service:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:latest

# Push to ECR
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:latest
```

## Step 2: Deploy CloudFormation Stack

### Development Environment (Cost-Optimized)

```bash
aws cloudformation create-stack \
  --stack-name aaa-service-dev \
  --template-body file://cloudformation.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=dev \
    ParameterKey=ContainerImage,ParameterValue=<account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:latest \
    ParameterKey=EnableMultiAZ,ParameterValue=false \
    ParameterKey=EnableAutoMigrate,ParameterValue=true \
    ParameterKey=EnableSeed,ParameterValue=true \
    ParameterKey=EnableDocs,ParameterValue=true \
    ParameterKey=DBInstanceClass,ParameterValue=db.t3.micro \
    ParameterKey=CacheNodeType,ParameterValue=cache.t3.micro \
    ParameterKey=TaskCPU,ParameterValue=512 \
    ParameterKey=TaskMemory,ParameterValue=1024 \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### Staging Environment

```bash
aws cloudformation create-stack \
  --stack-name aaa-service-staging \
  --template-body file://cloudformation.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=staging \
    ParameterKey=ContainerImage,ParameterValue=<account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:latest \
    ParameterKey=EnableMultiAZ,ParameterValue=false \
    ParameterKey=EnableAutoMigrate,ParameterValue=true \
    ParameterKey=EnableSeed,ParameterValue=false \
    ParameterKey=EnableDocs,ParameterValue=true \
    ParameterKey=DBInstanceClass,ParameterValue=db.t3.small \
    ParameterKey=CacheNodeType,ParameterValue=cache.t3.small \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### Production Environment

```bash
aws cloudformation create-stack \
  --stack-name aaa-service-prod \
  --template-body file://cloudformation.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=prod \
    ParameterKey=ContainerImage,ParameterValue=<account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:latest \
    ParameterKey=EnableMultiAZ,ParameterValue=true \
    ParameterKey=EnableAutoMigrate,ParameterValue=false \
    ParameterKey=EnableSeed,ParameterValue=false \
    ParameterKey=EnableDocs,ParameterValue=false \
    ParameterKey=DBInstanceClass,ParameterValue=db.r6g.large \
    ParameterKey=CacheNodeType,ParameterValue=cache.r6g.large \
    ParameterKey=TaskCPU,ParameterValue=1024 \
    ParameterKey=TaskMemory,ParameterValue=2048 \
    ParameterKey=CertificateArn,ParameterValue=arn:aws:acm:us-east-1:ACCOUNT_ID:certificate/CERT_ID \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

## Step 3: Monitor Stack Creation

```bash
# Check stack status
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].StackStatus' \
  --output text

# Watch stack events
aws cloudformation describe-stack-events \
  --stack-name aaa-service-dev \
  --max-items 10
```

## Step 4: Get Deployment Outputs

```bash
# Get ALB DNS name
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' \
  --output text

# Get all outputs
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs'
```

## Step 5: Verify Deployment

```bash
# Get the ALB DNS
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' \
  --output text)

# Test health endpoint
curl http://$ALB_DNS/health

# Test API (if docs enabled)
curl http://$ALB_DNS/swagger/index.html
```

## Step 6: Access Secrets

```bash
# Get database credentials
aws secretsmanager get-secret-value \
  --secret-id aaa-service-dev-db-credentials \
  --query SecretString \
  --output text | jq -r '.password'

# Get JWT secret
aws secretsmanager get-secret-value \
  --secret-id aaa-service-dev-jwt-secret \
  --query SecretString \
  --output text
```

## Updating the Stack

```bash
# Update with new container image
aws cloudformation update-stack \
  --stack-name aaa-service-dev \
  --template-body file://cloudformation.yaml \
  --parameters \
    ParameterKey=ContainerImage,ParameterValue=<account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:v1.1.0 \
  --capabilities CAPABILITY_NAMED_IAM

# Or use previous parameter values
aws cloudformation update-stack \
  --stack-name aaa-service-dev \
  --use-previous-template \
  --parameters \
    ParameterKey=ContainerImage,ParameterValue=<account-id>.dkr.ecr.us-east-1.amazonaws.com/aaa-service:v1.1.0 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Deleting the Stack

```bash
# Delete the stack (will snapshot RDS before deletion)
aws cloudformation delete-stack --stack-name aaa-service-dev

# Wait for deletion to complete
aws cloudformation wait stack-delete-complete --stack-name aaa-service-dev
```

## Troubleshooting

### Check ECS Task Logs

```bash
# Get log group name
LOG_GROUP=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`LogGroupName`].OutputValue' \
  --output text)

# View recent logs
aws logs tail $LOG_GROUP --follow
```

### Check ECS Service Status

```bash
# Get cluster and service name
CLUSTER=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
  --output text)

SERVICE=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ECSServiceName`].OutputValue' \
  --output text)

# Describe service
aws ecs describe-services \
  --cluster $CLUSTER \
  --services $SERVICE
```

### Check Target Group Health

```bash
# Get target group ARN from AWS Console or CLI
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>
```

## CloudWatch Alarms

The template automatically creates the following alarms:

- **High CPU**: Triggers when ECS CPU utilization exceeds 80%
- **High Memory**: Triggers when ECS memory utilization exceeds 80%
- **Unhealthy Targets**: Triggers when ALB has unhealthy targets
- **Database Connections**: Triggers when RDS connections exceed 150

Configure SNS notifications to receive alerts:

```bash
# Create SNS topic
aws sns create-topic --name aaa-service-alerts

# Subscribe email to topic
aws sns subscribe \
  --topic-arn arn:aws:sns:us-east-1:ACCOUNT_ID:aaa-service-alerts \
  --protocol email \
  --notification-endpoint your-email@example.com
```

## Best Practices

1. **Use separate AWS accounts** for dev, staging, and production
2. **Enable Multi-AZ** for production environments
3. **Set up automated backups** for RDS (already configured)
4. **Use AWS Secrets Manager rotation** for database credentials
5. **Configure CloudWatch dashboards** for monitoring
6. **Set up AWS WAF** for the ALB in production
7. **Use VPC Flow Logs** for network troubleshooting
8. **Enable AWS Config** for compliance tracking

## Security Considerations

1. **Secrets**: All sensitive data stored in AWS Secrets Manager
2. **Encryption**: RDS and Redis encryption at rest enabled
3. **Network**: Private subnets for ECS, RDS, and Redis
4. **IAM**: Least privilege roles for ECS tasks
5. **Security Groups**: Restricted access between components
6. **HTTPS**: Use ACM certificates for production

## Cost Optimization Tips

1. **Disable Multi-AZ** for non-production environments
2. **Use Fargate Spot** for dev/staging (modify template)
3. **Right-size instances** based on actual usage
4. **Enable auto-scaling** to scale down during off-hours
5. **Use CloudWatch Insights** to identify unused resources
6. **Set up billing alerts** to track spending
