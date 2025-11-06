# ECS Task Definition Usage Guide

This guide explains how to use the `task-definition.json` file to deploy the AAA service directly to ECS without CloudFormation.

## Overview

The `task-definition.json` file is a template for ECS Fargate task definition. You need to replace placeholder values with your actual AWS resources before registering it.

## Placeholders to Replace

| Placeholder | Description | Example |
|------------|-------------|---------|
| `ENVIRONMENT` | Environment name | `dev`, `staging`, `prod` |
| `REGION` | AWS region | `us-east-1` |
| `ACCOUNT_ID` | Your AWS account ID | `123456789012` |
| `DB_HOST` | RDS PostgreSQL endpoint | `aaa-service-dev-db.xxxxx.us-east-1.rds.amazonaws.com` |
| `REDIS_HOST` | ElastiCache Redis endpoint | `aaa-service-dev-redis.xxxxx.cache.amazonaws.com` |
| `:latest` | Docker image tag | `:v1.0.0` |

## Prerequisites

1. **IAM Roles Created**:
   - Task Execution Role: `aaa-service-ENVIRONMENT-task-execution-role`
   - Task Role: `aaa-service-ENVIRONMENT-task-role`

2. **Secrets Manager Secrets Created**:
   - Database credentials: `aaa-service-ENVIRONMENT-db-credentials`
   - JWT secret: `aaa-service-ENVIRONMENT-jwt-secret`

3. **CloudWatch Log Group Created**:
   - Log group: `/ecs/aaa-service-ENVIRONMENT`

4. **ECR Image Pushed**:
   - Image available in ECR

## Step 1: Get Resource Values

### Get RDS Endpoint (if using CloudFormation)
```bash
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`RDSEndpoint`].OutputValue' \
  --output text
```

### Get Redis Endpoint (if using CloudFormation)
```bash
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`RedisEndpoint`].OutputValue' \
  --output text
```

### Get AWS Account ID
```bash
aws sts get-caller-identity --query Account --output text
```

## Step 2: Create Environment-Specific Task Definition

Create a copy of the task definition for your environment:

```bash
# Create task definition for dev
cp task-definition.json task-definition-dev.json
```

## Step 3: Replace Placeholders Manually

Edit `task-definition-dev.json` and replace:

1. Replace all `ENVIRONMENT` with `dev`
2. Replace all `REGION` with your region (e.g., `us-east-1`)
3. Replace all `ACCOUNT_ID` with your AWS account ID
4. Replace `DB_HOST` with your RDS endpoint
5. Replace `REDIS_HOST` with your Redis endpoint
6. Optionally replace `:latest` with specific image tag (e.g., `:v1.0.0`)

### Example using `sed` (one-liner):
```bash
sed -e 's/ENVIRONMENT/dev/g' \
    -e 's/REGION/us-east-1/g' \
    -e 's/ACCOUNT_ID/123456789012/g' \
    -e 's/DB_HOST/aaa-dev-db.xxxxx.us-east-1.rds.amazonaws.com/g' \
    -e 's/REDIS_HOST/aaa-dev-redis.xxxxx.cache.amazonaws.com/g' \
    task-definition.json > task-definition-dev.json
```

## Step 4: Register Task Definition

```bash
# Register the task definition
aws ecs register-task-definition \
  --cli-input-json file://task-definition-dev.json \
  --region us-east-1

# Or get the task definition ARN
TASK_DEF_ARN=$(aws ecs register-task-definition \
  --cli-input-json file://task-definition-dev.json \
  --region us-east-1 \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

echo "Registered: $TASK_DEF_ARN"
```

## Step 5: Update ECS Service

```bash
# Update service with new task definition
aws ecs update-service \
  --cluster aaa-service-dev-cluster \
  --service aaa-service-dev \
  --task-definition aaa-service-dev \
  --region us-east-1

# Or force new deployment
aws ecs update-service \
  --cluster aaa-service-dev-cluster \
  --service aaa-service-dev \
  --task-definition aaa-service-dev \
  --force-new-deployment \
  --region us-east-1
```

## Step 6: Verify Deployment

```bash
# Check service status
aws ecs describe-services \
  --cluster aaa-service-dev-cluster \
  --services aaa-service-dev \
  --region us-east-1 \
  --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount}'

# Check task status
aws ecs list-tasks \
  --cluster aaa-service-dev-cluster \
  --service-name aaa-service-dev \
  --region us-east-1

# View logs
aws logs tail /ecs/aaa-service-dev --follow
```

## Task Definition Configuration

### Resource Allocation

Default values in the template:
- **CPU**: 512 (0.5 vCPU)
- **Memory**: 1024 MB (1 GB)

To modify, change the `cpu` and `memory` fields:

```json
{
  "cpu": "1024",
  "memory": "2048"
}
```

Valid CPU/Memory combinations for Fargate:

| CPU | Memory Options |
|-----|----------------|
| 256 | 512, 1024, 2048 |
| 512 | 1024, 2048, 3072, 4096 |
| 1024 | 2048, 3072, 4096, 5120, 6144, 7168, 8192 |
| 2048 | 4096 to 16384 (1 GB increments) |
| 4096 | 8192 to 30720 (1 GB increments) |

### Environment Variables

Key environment variables you may want to customize:

- `AAA_AUTO_MIGRATE`: Set to `"true"` to run migrations on startup
- `AAA_RUN_SEED`: Set to `"true"` to seed data on startup
- `AAA_ENABLE_DOCS`: Set to `"true"` to enable Swagger docs
- `LOG_LEVEL`: `"debug"`, `"info"`, `"warn"`, `"error"`
- `GIN_MODE`: `"release"` or `"debug"`

### Health Check

The health check is configured to:
- Check `/health` endpoint every 30 seconds
- Timeout after 10 seconds
- Allow 3 retries before marking unhealthy
- Wait 60 seconds before starting checks (startup grace period)

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Update task definition
  run: |
    sed -e 's/ENVIRONMENT/${{ env.ENVIRONMENT }}/g' \
        -e 's/REGION/${{ env.AWS_REGION }}/g' \
        -e 's/ACCOUNT_ID/${{ env.AWS_ACCOUNT_ID }}/g' \
        -e 's/DB_HOST/${{ env.DB_HOST }}/g' \
        -e 's/REDIS_HOST/${{ env.REDIS_HOST }}/g' \
        -e "s/:latest/:${{ github.sha }}/g" \
        task-definition.json > task-definition-updated.json

- name: Deploy to ECS
  uses: aws-actions/amazon-ecs-deploy-task-definition@v1
  with:
    task-definition: task-definition-updated.json
    service: aaa-service-${{ env.ENVIRONMENT }}
    cluster: aaa-service-${{ env.ENVIRONMENT }}-cluster
    wait-for-service-stability: true
```

## Troubleshooting

### Task fails to start

1. **Check IAM roles**: Ensure execution role has permissions for ECR, Secrets Manager, CloudWatch
2. **Check secrets**: Verify secrets exist in Secrets Manager
3. **Check image**: Ensure ECR image exists and is accessible
4. **Check logs**: View CloudWatch logs for errors

```bash
# Get task ID
TASK_ID=$(aws ecs list-tasks \
  --cluster aaa-service-dev-cluster \
  --service-name aaa-service-dev \
  --query 'taskArns[0]' \
  --output text)

# Describe task
aws ecs describe-tasks \
  --cluster aaa-service-dev-cluster \
  --tasks $TASK_ID
```

### Health check failing

Check if the application is listening on port 8080:

```bash
# SSH into task (if exec enabled)
aws ecs execute-command \
  --cluster aaa-service-dev-cluster \
  --task $TASK_ID \
  --container aaa-service \
  --command "/bin/sh" \
  --interactive
```

### Container logs

```bash
# Tail logs
aws logs tail /ecs/aaa-service-dev --follow

# Get recent errors
aws logs filter-log-events \
  --log-group-name /ecs/aaa-service-dev \
  --filter-pattern "ERROR" \
  --max-items 50
```

## Best Practices

1. **Version your task definitions**: Use git tags or semantic versioning for image tags
2. **Use separate files per environment**: `task-definition-dev.json`, `task-definition-prod.json`
3. **Don't commit environment-specific files**: Add `task-definition-*.json` to `.gitignore` (except the template)
4. **Test in lower environments first**: Always deploy to dev/staging before production
5. **Monitor deployments**: Watch CloudWatch metrics during and after deployment
6. **Keep secrets in Secrets Manager**: Never hardcode credentials in task definition
