# Quick Start: Deploy Complete CI/CD Pipeline

This guide will get you from zero to a fully automated CI/CD pipeline in ~1 hour.

## Prerequisites Checklist

- [ ] AWS CLI configured (`aws sts get-caller-identity`)
- [ ] GitHub Personal Access Token with `repo` and `admin:repo_hook` permissions
- [ ] ECR repository exists (✅ Already done: `aaa-service`)
- [ ] You're in the project root: `/Users/kaushik/aaa-service`

## Step 1: Store GitHub Token (2 minutes)

```bash
# Replace YOUR_GITHUB_TOKEN with your actual token
aws secretsmanager create-secret \
  --name aaa-service-github-token \
  --description "GitHub PAT for AAA service pipelines" \
  --secret-string '{"token":"YOUR_GITHUB_TOKEN_HERE"}' \
  --region ap-south-1

# Verify it was created
aws secretsmanager describe-secret \
  --secret-id aaa-service-github-token \
  --region ap-south-1 \
  --query 'ARN' \
  --output text
```

Save the ARN that's printed - you'll need it for Step 4.

## Step 2: Deploy Infrastructure Stacks (45-60 minutes)

All three stacks can be deployed in parallel to save time.

### Deploy Beta Environment

```bash
cd infrastructure

aws cloudformation deploy \
  --stack-name aaa-service-beta \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/beta.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1 \
  --no-fail-on-empty-changeset
```

**Time**: ~15-20 minutes

### Deploy Staging Environment (In Another Terminal)

```bash
cd infrastructure

aws cloudformation deploy \
  --stack-name aaa-service-staging \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/staging.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1 \
  --no-fail-on-empty-changeset
```

**Time**: ~15-20 minutes

### Deploy Production Environment (In Another Terminal)

```bash
cd infrastructure

aws cloudformation deploy \
  --stack-name aaa-service-prod \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/prod.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1 \
  --no-fail-on-empty-changeset
```

**Time**: ~15-20 minutes

### Monitor Progress

```bash
# Check beta
watch -n 10 'aws cloudformation describe-stacks --stack-name aaa-service-beta --query "Stacks[0].StackStatus" --output text'

# Check staging
watch -n 10 'aws cloudformation describe-stacks --stack-name aaa-service-staging --query "Stacks[0].StackStatus" --output text'

# Check prod
watch -n 10 'aws cloudformation describe-stacks --stack-name aaa-service-prod --query "Stacks[0].StackStatus" --output text'
```

Wait for all three to show `CREATE_COMPLETE` or `UPDATE_COMPLETE`.

## Step 3: Set Up Production Blue-Green Deployment (5 minutes)

Production uses CodeDeploy for zero-downtime blue-green deployments.

### Create Test Target Group

```bash
# Get VPC ID from production stack
VPC_ID=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`VPCId`].OutputValue' \
  --output text \
  --region ap-south-1)

echo "VPC ID: $VPC_ID"

# Create test target group
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

### Deploy CodeDeploy Stack

```bash
# Get required ARNs
ALB_LISTENER_ARN=$(aws elbv2 describe-listeners \
  --load-balancer-arn $(aws cloudformation describe-stacks \
    --stack-name aaa-service-prod \
    --query 'Stacks[0].Outputs[?contains(OutputKey, `LoadBalancer`)].OutputValue' \
    --output text \
    --region ap-south-1 | xargs aws elbv2 describe-load-balancers \
    --names --query 'LoadBalancers[0].LoadBalancerArn' --output text --region ap-south-1) \
  --query 'Listeners[?Port==`80`].ListenerArn' \
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

echo "ALB Listener ARN: $ALB_LISTENER_ARN"
echo "Production TG ARN: $PROD_TG_ARN"
echo "Test TG ARN: $TEST_TG_ARN"

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

# Wait for completion
aws cloudformation wait stack-create-complete \
  --stack-name aaa-service-prod-codedeploy \
  --region ap-south-1
```

## Step 4: Deploy CI/CD Pipelines (5 minutes)

This creates 3 CodePipeline pipelines with GitHub webhooks for automatic triggering.

```bash
cd infrastructure

# Get the GitHub token ARN from Step 1
GITHUB_TOKEN_ARN=$(aws secretsmanager describe-secret \
  --secret-id aaa-service-github-token \
  --region ap-south-1 \
  --query 'ARN' \
  --output text)

# Create parameter file
cat > pipeline-params.json <<EOF
[
  {
    "ParameterKey": "GitHubOwner",
    "ParameterValue": "YOUR_GITHUB_USERNAME"
  },
  {
    "ParameterKey": "GitHubTokenSecretArn",
    "ParameterValue": "$GITHUB_TOKEN_ARN"
  }
]
EOF

# IMPORTANT: Edit pipeline-params.json and replace YOUR_GITHUB_USERNAME
# with your actual GitHub username or organization name
nano pipeline-params.json  # or vim, code, etc.

# Deploy the pipeline stack
aws cloudformation create-stack \
  --stack-name aaa-service-pipeline \
  --template-body file://cloudformation-pipeline.yaml \
  --parameters file://pipeline-params.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1

# Monitor deployment
watch -n 5 'aws cloudformation describe-stacks --stack-name aaa-service-pipeline --query "Stacks[0].StackStatus" --output text --region ap-south-1'
```

**Time**: ~5 minutes

Wait for `CREATE_COMPLETE`.

## Step 5: Create Git Branches (1 minute)

```bash
cd /Users/kaushik/aaa-service

# Create beta branch
git checkout -b beta
git push -u origin beta

# Create staging branch
git checkout main
git checkout -b staging
git push -u origin staging

# Return to main
git checkout main
```

## Step 6: Verify Auto-Trigger Setup (2 minutes)

The CodePipeline webhooks are automatically configured to trigger on git push. Let's verify:

```bash
# List all webhooks
aws codepipeline list-webhooks --region ap-south-1 --output table

# You should see 3 webhooks:
# - aaa-service-beta-webhook (triggers on push to beta branch)
# - aaa-service-staging-webhook (triggers on push to staging branch)
# - aaa-service-prod-webhook (triggers on push to main branch)
```

Check GitHub webhooks:
1. Go to: https://github.com/YOUR_USERNAME/aaa-service/settings/hooks
2. You should see 3 webhooks from AWS CodePipeline
3. Each should show a green checkmark

## Step 7: Test the Pipeline! (5 minutes)

### Test Beta Deployment

```bash
git checkout beta
echo "# Test deployment $(date)" >> README.md
git add README.md
git commit -m "test: trigger beta pipeline"
git push origin beta
```

**Monitor the pipeline:**
```bash
# Watch pipeline status
watch -n 5 'aws codepipeline get-pipeline-state --name aaa-service-beta-pipeline --region ap-south-1 --query "stageStates[*].{Stage:stageName,Status:latestExecution.status}" --output table'

# View build logs
aws logs tail /aws/codebuild/aaa-service-build --follow --region ap-south-1
```

### Test Staging Deployment

```bash
git checkout staging
git merge beta
git push origin staging
```

### Test Production Deployment

```bash
git checkout main
git merge staging
git push origin main
```

**Note**: Production requires manual approval. Approve it via AWS Console:
1. Go to: https://console.aws.amazon.com/codesuite/codepipeline/pipelines
2. Click on `aaa-service-prod-pipeline`
3. Click the "Review" button in the Approval stage
4. Click "Approve"

Or via CLI:
```bash
# Get the approval token
TOKEN=$(aws codepipeline get-pipeline-state \
  --name aaa-service-prod-pipeline \
  --region ap-south-1 \
  --query 'stageStates[?stageName==`Approval`].actionStates[0].latestExecution.token' \
  --output text)

# Approve deployment
aws codepipeline put-approval-result \
  --pipeline-name aaa-service-prod-pipeline \
  --stage-name Approval \
  --action-name ManualApproval \
  --result status=Approved,summary="Approved via CLI" \
  --token $TOKEN \
  --region ap-south-1
```

## Verification: Check Your Deployments

```bash
# Get service URLs
echo "Beta: $(aws cloudformation describe-stacks --stack-name aaa-service-beta --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' --output text --region ap-south-1)"

echo "Staging: $(aws cloudformation describe-stacks --stack-name aaa-service-staging --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' --output text --region ap-south-1)"

echo "Production: $(aws cloudformation describe-stacks --stack-name aaa-service-prod --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' --output text --region ap-south-1)"

# Test health endpoints
curl http://$(aws cloudformation describe-stacks --stack-name aaa-service-beta --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' --output text --region ap-south-1)/health
```

## 🎉 Success! Your Pipeline is Live

You now have:
- ✅ 3 environments running (beta, staging, production)
- ✅ 3 automated pipelines with GitHub webhooks
- ✅ Auto-deployment on push to beta/staging branches
- ✅ Blue-green deployment with manual approval for production
- ✅ Zero-downtime deployments

## Daily Workflow

```bash
# Develop feature
git checkout beta
# ... make changes ...
git commit -am "feat: new feature"
git push origin beta
# ✅ Automatically deploys to beta

# Test in staging
git checkout staging
git merge beta
git push origin staging
# ✅ Automatically deploys to staging

# Release to production
git checkout main
git merge staging
git push origin main
# ✅ Triggers pipeline → Manual approval → Blue-green deploy
```

## Monitoring Commands

```bash
# Pipeline status
aws codepipeline get-pipeline-state --name aaa-service-beta-pipeline --region ap-south-1

# Service health
aws ecs describe-services --cluster aaa-service-beta-cluster --services aaa-service-beta --region ap-south-1

# Application logs
aws logs tail /ecs/aaa-service-beta --follow --region ap-south-1

# Build logs
aws logs tail /aws/codebuild/aaa-service-build --follow --region ap-south-1
```

## Troubleshooting

### Pipeline doesn't trigger automatically
- Check GitHub webhooks are active (green checkmark)
- Verify GitHub token has `repo` and `admin:repo_hook` permissions
- Check CloudWatch Events: `aws events list-rules --region ap-south-1`

### Build fails
- Check CodeBuild logs: `aws logs tail /aws/codebuild/aaa-service-build --follow`
- Verify Dockerfile and buildspec.yml are correct
- Check ECR permissions

### Deployment fails
- Check ECS service events
- Verify task definition is valid
- Check security groups allow ALB → ECS traffic
- Verify secrets exist in Secrets Manager

### Need help?
See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for detailed troubleshooting steps.
