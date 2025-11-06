# CI/CD Pipeline Overview

This document provides a high-level overview of the AAA service CI/CD pipeline setup.

## Architecture

```
┌─────────────┐
│   GitHub    │
│  Repository │
└──────┬──────┘
       │ Push to main
       ▼
┌─────────────────────────────────────────┐
│           CodePipeline                   │
│  ┌───────────────────────────────────┐  │
│  │ Source Stage                       │  │
│  │  - Pulls code from repository      │  │
│  └────────┬──────────────────────────┘  │
│           │                              │
│  ┌────────▼──────────────────────────┐  │
│  │ Build Stage (CodeBuild)            │  │
│  │  - Builds Docker image             │  │
│  │  - Pushes to ECR                   │  │
│  │  - Prepares task definition        │  │
│  │  - Creates artifacts               │  │
│  └────────┬──────────────────────────┘  │
│           │                              │
│  ┌────────▼──────────────────────────┐  │
│  │ Deploy Stage (CodeDeploy)          │  │
│  │  - Blue-green deployment           │  │
│  │  - Traffic shifting                │  │
│  │  - Automatic rollback              │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
       │
       ▼
┌─────────────────────┐
│   ECS on Fargate    │
│  Zero-downtime      │
│    Deployment       │
└─────────────────────┘
```

## Files in This Repository

### Deployment Files

| File | Purpose |
|------|---------|
| `buildspec.yml` | CodeBuild instructions for building Docker image and creating artifacts |
| `appspec.yml` | CodeDeploy deployment specification for ECS blue-green deployment |
| `task-definition.json` | ECS task definition template with placeholders |
| `cloudformation.yaml` | Main infrastructure (VPC, ECS, RDS, Redis, ALB) |
| `cloudformation-codedeploy.yaml` | CodeDeploy setup (application, deployment group, IAM roles) |

### Documentation

| File | Purpose |
|------|---------|
| `CODEDEPLOY_SETUP.md` | Complete guide for setting up blue-green deployments |
| `DEPLOYMENT.md` | Infrastructure deployment guide |
| `TASK_DEFINITION_README.md` | Manual ECS deployment guide |
| `CICD_README.md` | This file - pipeline overview |

### Docker

| File | Purpose |
|------|---------|
| `Dockerfile` | Multi-stage Docker build configuration |

## Quick Start

### 1. Deploy Infrastructure

```bash
# Deploy main infrastructure stack
aws cloudformation create-stack \
  --stack-name aaa-service-dev \
  --template-body file://cloudformation.yaml \
  --parameters ParameterKey=Environment,ParameterValue=dev \
  --capabilities CAPABILITY_NAMED_IAM
```

### 2. Create Test Target Group

```bash
# Get VPC ID
VPC_ID=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`VPCId`].OutputValue' \
  --output text)

# Create test target group
aws elbv2 create-target-group \
  --name aaa-service-dev-test-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id $VPC_ID \
  --target-type ip \
  --health-check-path /health
```

### 3. Deploy CodeDeploy Stack

```bash
# Get required ARNs
ALB_LISTENER_ARN=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`HTTPListenerArn`].OutputValue' \
  --output text)

PROD_TG_ARN=$(aws elbv2 describe-target-groups \
  --names aaa-service-dev-http-tg \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text)

TEST_TG_ARN=$(aws elbv2 describe-target-groups \
  --names aaa-service-dev-test-tg \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text)

# Deploy
aws cloudformation create-stack \
  --stack-name aaa-service-dev-codedeploy \
  --template-body file://cloudformation-codedeploy.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=dev \
    ParameterKey=ALBListenerArn,ParameterValue=$ALB_LISTENER_ARN \
    ParameterKey=ProductionTargetGroupArn,ParameterValue=$PROD_TG_ARN \
    ParameterKey=TestTargetGroupArn,ParameterValue=$TEST_TG_ARN \
  --capabilities CAPABILITY_NAMED_IAM
```

### 4. Configure CodePipeline

Add a Deploy stage to your existing CodePipeline that uses CodeDeploy.

See `CODEDEPLOY_SETUP.md` for detailed instructions.

### 5. Test the Pipeline

```bash
# Make a code change and push
git add .
git commit -m "test: trigger pipeline"
git push origin main

# Watch deployment
aws deploy list-deployments \
  --application-name aaa-service-dev

# View service URL
aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' \
  --output text
```

## What Happens on Git Push

1. **CodePipeline Detects Change** - Triggers on push to main
2. **CodeBuild Runs** - Executes `buildspec.yml`
   - Builds Docker image
   - Tags with commit SHA
   - Pushes to ECR
   - Prepares task definition
   - Creates artifacts
3. **CodeDeploy Executes** - Uses `appspec.yml`
   - Creates new task set (GREEN)
   - Shifts traffic from BLUE to GREEN
   - Monitors health checks
   - Terminates old tasks
   - Automatic rollback on failure

## Deployment Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| **AllAtOnce** | Instant 100% traffic shift | Dev/staging |
| **Canary10Percent5Minutes** | 10% → wait 5 min → 90% | Production |
| **Linear10PercentEvery1Minutes** | Gradual 10% every minute | Critical services |

## Key Features

✅ **Zero-downtime deployments** - Traffic shifts seamlessly
✅ **Automatic rollback** - Reverts on failure
✅ **Health monitoring** - Validates before completion
✅ **Cost-optimized** - Minimal double-capacity time
✅ **Environment-specific** - Dev, staging, prod configs
✅ **Docker-based** - Consistent builds everywhere

## Environment Variables

Configure these in CodeBuild:

```
AWS_ACCOUNT_ID       - Your AWS account ID
IMAGE_REPO_NAME      - aaa-service
ENVIRONMENT          - dev/staging/prod
DB_HOST              - RDS endpoint
REDIS_HOST           - Redis endpoint
AWS_DEFAULT_REGION   - us-east-1 (or your region)
```

## Monitoring

### Deployment Status
```bash
aws deploy get-deployment --deployment-id d-XXXXXXXXX
```

### Service Health
```bash
aws ecs describe-services \
  --cluster aaa-service-dev-cluster \
  --services aaa-service-dev
```

### Application Logs
```bash
aws logs tail /ecs/aaa-service-dev --follow
```

### Metrics
- CodePipeline → View pipeline execution
- CodeDeploy → View deployment history
- CloudWatch → View logs and metrics
- ECS → View task status

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Build fails | Check CodeBuild logs, verify Dockerfile |
| Deployment fails | Check task definition, verify IAM roles |
| Health checks fail | Check /health endpoint, security groups |
| Rollback triggered | Check CloudWatch alarms, application logs |

See `CODEDEPLOY_SETUP.md` for detailed troubleshooting.

## Cost Estimate

| Component | Monthly Cost |
|-----------|--------------|
| CodePipeline | $1 (first pipeline free) |
| CodeBuild | $0.005/min (~$5-10/month) |
| CodeDeploy | Free for ECS |
| ECR Storage | $0.10/GB (~$1-5/month) |
| **Total** | **~$7-16/month** |

Note: Double-capacity during deployments adds ~$0.10 per deployment

## Security

- ✅ Secrets stored in AWS Secrets Manager
- ✅ IAM roles with least-privilege
- ✅ Private subnets for ECS tasks
- ✅ Encrypted ECR repositories
- ✅ VPC isolation

## Support

- **Infrastructure Issues**: See `DEPLOYMENT.md`
- **Pipeline Issues**: See `CODEDEPLOY_SETUP.md`
- **Manual Deployments**: See `TASK_DEFINITION_README.md`

## Next Steps

1. ✅ Deploy infrastructure stacks
2. ✅ Configure CodePipeline deploy stage
3. ✅ Test deployment in dev
4. ✅ Set up production pipeline
5. ✅ Configure monitoring and alerts
6. ✅ Document runbooks for your team
