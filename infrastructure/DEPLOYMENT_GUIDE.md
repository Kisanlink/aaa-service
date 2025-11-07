# AAA Service Multi-Environment Deployment Guide

This guide walks you through setting up a complete CI/CD pipeline with three environments:
- **Beta**: Automatic deployment on push to `beta` branch
- **Staging**: Automatic deployment on push to `staging` branch (for PR testing)
- **Production**: Blue-green deployment with manual approval on merge to `main` branch

## Architecture Overview

```
Git Workflow                Pipeline                 Deployment
────────────                ────────                 ──────────

Push to beta     ──────────> Beta Pipeline    ──────> ECS Rolling Update (Beta)
                             (Auto)

Push to staging  ──────────> Staging Pipeline ──────> ECS Rolling Update (Staging)
                             (Auto)

Merge to main    ──────────> Prod Pipeline    ──────> CodeDeploy Blue-Green (Prod)
                             (Manual Approval)
```

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **GitHub Personal Access Token** with `repo` and `admin:repo_hook` permissions
3. **ECR Repository** already created (✅ Already exists: `aaa-service`)
4. **AWS Region**: `ap-south-1` (Mumbai)

## Step 1: Store GitHub Token in Secrets Manager

```bash
# Create a secret for GitHub token
aws secretsmanager create-secret \
  --name aaa-service-github-token \
  --description "GitHub personal access token for AAA service pipeline" \
  --secret-string '{"token":"YOUR_GITHUB_TOKEN_HERE"}' \
  --region ap-south-1

# Save the ARN - you'll need it
GITHUB_TOKEN_ARN=$(aws secretsmanager describe-secret \
  --secret-id aaa-service-github-token \
  --region ap-south-1 \
  --query 'ARN' \
  --output text)

echo "GitHub Token ARN: $GITHUB_TOKEN_ARN"
```

## Step 2: Deploy Infrastructure for All Environments

### Option A: Using Makefile (Recommended)

```bash
cd infrastructure

# Deploy all environments
make deploy-beta
make deploy-staging
make deploy-prod
```

### Option B: Using AWS CLI

```bash
# Deploy Beta
aws cloudformation deploy \
  --stack-name aaa-service-beta \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/beta.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1

# Deploy Staging
aws cloudformation deploy \
  --stack-name aaa-service-staging \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/staging.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1

# Deploy Production
aws cloudformation deploy \
  --stack-name aaa-service-prod \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/prod.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1
```

**Note**: Each deployment takes ~15-20 minutes due to RDS and ElastiCache provisioning.

## Step 3: Set Up Blue-Green Deployment for Production

Production uses CodeDeploy for zero-downtime blue-green deployments.

### 3.1: Create Test Target Group

```bash
# Get VPC ID from production stack
VPC_ID=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`VPCId`].OutputValue' \
  --output text \
  --region ap-south-1)

# Create test target group for blue-green deployment
aws elbv2 create-target-group \
  --name aaa-service-prod-test-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id $VPC_ID \
  --target-type ip \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 10 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3 \
  --region ap-south-1 \
  --tags Key=Name,Value=aaa-service-prod-test-tg Key=Environment,Value=prod
```

### 3.2: Deploy CodeDeploy Stack

```bash
# Get required ARNs
ALB_LISTENER_ARN=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-prod \
  --query 'Stacks[0].Outputs[?contains(OutputKey, `Listener`)].OutputValue' \
  --output text \
  --region ap-south-1)

PROD_TG_ARN=$(aws elbv2 describe-target-groups \
  --names aaa-service-prod-http-tg \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text \
  --region ap-south-1)

TEST_TG_ARN=$(aws elbv2 describe-target-groups \
  --names aaa-service-prod-test-tg \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text \
  --region ap-south-1)

# Deploy CodeDeploy stack
aws cloudformation create-stack \
  --stack-name aaa-service-prod-codedeploy \
  --template-body file://../cloudformation-codedeploy.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=prod \
    ParameterKey=ALBListenerArn,ParameterValue=$ALB_LISTENER_ARN \
    ParameterKey=ProductionTargetGroupArn,ParameterValue=$PROD_TG_ARN \
    ParameterKey=TestTargetGroupArn,ParameterValue=$TEST_TG_ARN \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1
```

### 3.3: Update Production ECS Service for Blue-Green

```bash
# Update the ECS service deployment controller to CODE_DEPLOY
# Note: This requires recreating the service
aws ecs update-service \
  --cluster aaa-service-prod-cluster \
  --service aaa-service-prod \
  --deployment-controller type=CODE_DEPLOY \
  --region ap-south-1
```

## Step 4: Deploy CI/CD Pipelines

```bash
# Create parameter file for pipeline
cat > pipeline-params.json <<EOF
[
  {
    "ParameterKey": "GitHubOwner",
    "ParameterValue": "YOUR_GITHUB_USERNAME_OR_ORG"
  },
  {
    "ParameterKey": "GitHubTokenSecretArn",
    "ParameterValue": "$GITHUB_TOKEN_ARN"
  }
]
EOF

# Deploy pipeline stack
aws cloudformation create-stack \
  --stack-name aaa-service-pipeline \
  --template-body file://cloudformation-pipeline.yaml \
  --parameters file://pipeline-params.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1

# Wait for completion
aws cloudformation wait stack-create-complete \
  --stack-name aaa-service-pipeline \
  --region ap-south-1
```

## Step 5: Set Up Git Branches

```bash
# From your repository root
cd /Users/kaushik/aaa-service

# Create and push beta branch
git checkout -b beta
git push -u origin beta

# Create and push staging branch
git checkout main
git checkout -b staging
git push -u origin staging

# Return to main
git checkout main
```

## Deployment Workflows

### Beta Environment (Development)
1. Make changes and commit to `beta` branch
2. Push to GitHub: `git push origin beta`
3. Pipeline automatically triggers
4. Build → Deploy to Beta ECS (rolling update)
5. **No approval required**

### Staging Environment (PR Testing)
1. Create PR or push to `staging` branch
2. Pipeline automatically triggers
3. Build → Deploy to Staging ECS (rolling update)
4. Test your changes
5. **No approval required**

### Production Environment (Production Release)
1. Merge PR to `main` branch
2. Pipeline automatically triggers
3. Build → **Manual Approval Step** → Blue-Green Deploy
4. Approve in AWS Console or CLI:
   ```bash
   aws codepipeline get-pipeline-state \
     --name aaa-service-prod-pipeline \
     --region ap-south-1

   # Get the approval token and approve
   aws codepipeline put-approval-result \
     --pipeline-name aaa-service-prod-pipeline \
     --stage-name Approval \
     --action-name ManualApproval \
     --result status=Approved,summary="Approved for production deployment" \
     --token YOUR_APPROVAL_TOKEN \
     --region ap-south-1
   ```
5. CodeDeploy performs blue-green deployment:
   - Creates new task set (GREEN)
   - Shifts test traffic → validates health
   - Shifts 10% production traffic → validates
   - Gradually shifts remaining traffic
   - Terminates old tasks (BLUE)

## Monitoring Deployments

### Check Pipeline Status
```bash
# Beta
aws codepipeline get-pipeline-state \
  --name aaa-service-beta-pipeline \
  --region ap-south-1

# Staging
aws codepipeline get-pipeline-state \
  --name aaa-service-staging-pipeline \
  --region ap-south-1

# Production
aws codepipeline get-pipeline-state \
  --name aaa-service-prod-pipeline \
  --region ap-south-1
```

### Check Deployment Status
```bash
# Beta/Staging (ECS rolling update)
aws ecs describe-services \
  --cluster aaa-service-beta-cluster \
  --services aaa-service-beta \
  --region ap-south-1

# Production (CodeDeploy blue-green)
aws deploy list-deployments \
  --application-name aaa-service-prod \
  --region ap-south-1

aws deploy get-deployment \
  --deployment-id d-XXXXXXXXX \
  --region ap-south-1
```

### View Service URLs
```bash
# Beta
aws cloudformation describe-stacks \
  --stack-name aaa-service-beta \
  --query 'Stacks[0].Outputs[?OutputKey==`HTTPEndpoint`].OutputValue' \
  --output text \
  --region ap-south-1

# Staging
aws cloudformation describe-stacks \
  --stack-name aaa-service-staging \
  --query 'Stacks[0].Outputs[?OutputKey==`HTTPEndpoint`].OutputValue' \
  --output text \
  --region ap-south-1

# Production
aws cloudformation describe-stacks \
  --stack-name aaa-service-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`HTTPEndpoint`].OutputValue' \
  --output text \
  --region ap-south-1
```

### View Logs
```bash
# Beta
aws logs tail /ecs/aaa-service-beta --follow --region ap-south-1

# Staging
aws logs tail /ecs/aaa-service-staging --follow --region ap-south-1

# Production
aws logs tail /ecs/aaa-service-prod --follow --region ap-south-1
```

## Troubleshooting

### Pipeline Fails to Trigger
- Check GitHub webhook configuration
- Verify GitHub token has correct permissions
- Check CloudWatch Events for errors

### Build Fails
- Check CodeBuild logs in CloudWatch
- Verify Dockerfile is correct
- Ensure buildspec.yml is in repository root

### Deployment Fails
- Check ECS service events
- Verify task definition is correct
- Check security groups and networking
- Verify secrets exist in Secrets Manager

### Blue-Green Deployment Fails
- Check CodeDeploy deployment logs
- Verify health check endpoint `/health` is working
- Check target group health checks
- Verify task can reach load balancer

### Rollback Production Deployment
```bash
# CodeDeploy automatically rolls back on failure
# Manual rollback:
aws deploy stop-deployment \
  --deployment-id d-XXXXXXXXX \
  --auto-rollback-enabled \
  --region ap-south-1
```

## Cost Optimization

### Development (Beta + Staging)
- Single NAT Gateway: ~$32/month
- RDS t3.micro: ~$15/month each
- Redis t3.micro: ~$12/month each
- ECS Fargate: ~$15/month each
- **Total: ~$100-120/month**

### Production
- Multi-AZ enabled
- Larger instances
- **Total: ~$150-200/month**

### Pipeline Costs
- CodePipeline: Free (first pipeline) + $1/additional pipeline
- CodeBuild: $0.005/minute (~$5-10/month)
- S3 artifacts: ~$1-5/month
- **Total: ~$7-16/month**

## Security Checklist

- ✅ Secrets stored in AWS Secrets Manager
- ✅ Private subnets for ECS tasks
- ✅ Security groups with least-privilege
- ✅ IAM roles with minimal permissions
- ✅ Encrypted RDS and Redis
- ✅ VPC isolation between environments
- ✅ Manual approval for production
- ✅ Encrypted S3 artifacts

## Next Steps

1. ✅ Set up DNS with Route53 (optional)
2. ✅ Configure ACM certificates for HTTPS
3. ✅ Set up CloudWatch alarms and SNS notifications
4. ✅ Configure automated testing in pipeline
5. ✅ Set up backup and disaster recovery
6. ✅ Document runbooks for operations team

## Support

For issues or questions:
1. Check CloudWatch Logs
2. Review CloudFormation events
3. Check ECS service events
4. Review CodePipeline execution history
