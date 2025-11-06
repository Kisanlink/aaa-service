# CodeDeploy Blue-Green Deployment Setup

This guide explains how to set up automated blue-green deployments for the AAA service using AWS CodeDeploy and your existing CodePipeline.

## Overview

Your current setup:
- ✅ CodePipeline builds Docker images
- ✅ CodeBuild pushes images to ECR
- ❌ No automated deployment to ECS

After setup:
- ✅ Automatic blue-green deployments
- ✅ Zero-downtime updates
- ✅ Automatic rollback on failure
- ✅ Gradual traffic shifting options

## What Happens When You Push to Main

### Complete Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 1. Developer pushes code to main branch                                 │
└────────────────────────┬────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 2. CodePipeline triggers automatically                                   │
│    - Source stage: Pulls latest code from repository                    │
└────────────────────────┬────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 3. CodeBuild executes buildspec.yml                                     │
│    ├─ Builds Docker image                                               │
│    ├─ Tags with commit SHA (e.g., abc1234)                             │
│    ├─ Pushes to ECR: aaa-service:abc1234                              │
│    ├─ Creates imagedefinitions.json                                     │
│    ├─ Generates taskdef.json with placeholders replaced                │
│    └─ Outputs artifacts: imagedefinitions.json, taskdef.json, appspec.yml│
└────────────────────────┬────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 4. CodeDeploy Blue-Green Deployment begins                              │
│                                                                          │
│    Step 1: Register new task definition                                 │
│    ├─ Creates revision with new image URI                              │
│    └─ Task definition: aaa-service:10 -> aaa-service:11               │
│                                                                          │
│    Step 2: Provision GREEN task set                                     │
│    ├─ Launches new tasks with updated image                            │
│    ├─ Tasks: 2 running (BLUE: abc0000, GREEN: abc1234)                │
│    └─ Waits for GREEN tasks to become healthy                          │
│                                                                          │
│    Step 3: Run lifecycle hooks (if configured)                          │
│    ├─ BeforeInstall: Pre-deployment validation                         │
│    ├─ AfterAllowTestTraffic: Health checks on test traffic            │
│    └─ BeforeAllowTraffic: Final validation before production           │
│                                                                          │
│    Step 4: Traffic shifting (based on strategy)                         │
│    ├─ AllAtOnce: Instant 100% shift to GREEN                          │
│    ├─ Canary: 10% → wait → 90%                                        │
│    └─ Linear: Gradual 10% every N minutes                             │
│         ALB routes: BLUE (80%) ──┐                                     │
│                     GREEN (20%) ─┘                                     │
│                                                                          │
│    Step 5: Monitor deployment health                                     │
│    ├─ Checks target health in GREEN target group                       │
│    ├─ Monitors CloudWatch alarms                                       │
│    └─ If unhealthy: Triggers automatic rollback                        │
│                                                                          │
│    Step 6: Complete deployment                                           │
│    ├─ All traffic now on GREEN (100%)                                  │
│    ├─ BLUE tasks remain for rollback window (5 min default)            │
│    └─ After wait time: Terminates BLUE tasks                           │
│                                                                          │
│    Result: ✅ Zero-downtime deployment complete                         │
│    - New version (abc1234) serving all traffic                          │
│    - Old version (abc0000) terminated                                    │
│    - Rollback available via CodeDeploy console                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Setup Instructions

### Prerequisites

1. Existing CodePipeline with Source and Build stages
2. ECR repository for aaa-service
3. Running ECS service (deployed via CloudFormation)
4. Two target groups (blue and green) attached to ALB

### Step 1: Create Test Target Group

Your ALB needs a second target group for blue-green deployments.

```bash
# Get VPC ID from CloudFormation
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
  --health-check-enabled \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3 \
  --matcher HttpCode=200
```

### Step 2: Update ECS Service for CodeDeploy

Your ECS service must use the CODE_DEPLOY deployment controller:

```bash
# Note: This requires recreating the ECS service
# You'll need to update your CloudFormation template

# Add to ECS service in cloudformation.yaml:
DeploymentController:
  Type: CODE_DEPLOY
```

### Step 3: Deploy CodeDeploy Stack

```bash
# Get required values
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

# Deploy CodeDeploy stack
aws cloudformation create-stack \
  --stack-name aaa-service-dev-codedeploy \
  --template-body file://cloudformation-codedeploy.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=dev \
    ParameterKey=ALBListenerArn,ParameterValue=$ALB_LISTENER_ARN \
    ParameterKey=ProductionTargetGroupArn,ParameterValue=$PROD_TG_ARN \
    ParameterKey=TestTargetGroupArn,ParameterValue=$TEST_TG_ARN \
    ParameterKey=TrafficShiftingType,ParameterValue=AllAtOnce \
  --capabilities CAPABILITY_NAMED_IAM
```

### Step 4: Update CodePipeline

Add a Deploy stage to your existing pipeline:

```json
{
  "name": "Deploy",
  "actions": [
    {
      "name": "DeployToECS",
      "actionTypeId": {
        "category": "Deploy",
        "owner": "AWS",
        "provider": "CodeDeployToECS",
        "version": "1"
      },
      "inputArtifacts": [
        {
          "name": "BuildArtifact"
        }
      ],
      "configuration": {
        "ApplicationName": "aaa-service-dev",
        "DeploymentGroupName": "aaa-service-dev-dg",
        "TaskDefinitionTemplateArtifact": "BuildArtifact",
        "TaskDefinitionTemplatePath": "taskdef.json",
        "AppSpecTemplateArtifact": "BuildArtifact",
        "AppSpecTemplatePath": "appspec.yml",
        "Image1ArtifactName": "BuildArtifact",
        "Image1ContainerName": "IMAGE1_NAME"
      }
    }
  ]
}
```

### Step 5: Configure Build Environment Variables

In your CodeBuild project, add environment variables:

```yaml
Environment Variables:
  - AWS_ACCOUNT_ID: <your-account-id>
  - IMAGE_REPO_NAME: aaa-service
  - ENVIRONMENT: dev
  - DB_HOST: <rds-endpoint>
  - REDIS_HOST: <redis-endpoint>
```

Or use Parameter Store/Secrets Manager:

```yaml
- DB_HOST:
    type: PARAMETER_STORE
    value: /aaa-service/dev/db-host
- REDIS_HOST:
    type: PARAMETER_STORE
    value: /aaa-service/dev/redis-host
```

## Traffic Shifting Strategies

### AllAtOnce (Default)
- Instant 100% traffic shift
- Fastest deployment
- Higher risk
- Best for: Dev/staging environments

### Canary10Percent5Minutes
- 10% traffic to new version
- Wait 5 minutes
- Shift remaining 90%
- Best for: Production with moderate traffic

### Linear10PercentEvery1Minutes
- Gradual 10% increments every minute
- 10 steps total (10 minutes)
- Safest option
- Best for: Critical production services

To change strategy:

```bash
aws cloudformation update-stack \
  --stack-name aaa-service-dev-codedeploy \
  --use-previous-template \
  --parameters \
    ParameterKey=TrafficShiftingType,ParameterValue=Canary10Percent5Minutes \
  --capabilities CAPABILITY_NAMED_IAM
```

## Monitoring Deployments

### View Deployment Status

```bash
# List deployments
aws deploy list-deployments \
  --application-name aaa-service-dev \
  --deployment-group-name aaa-service-dev-dg

# Get deployment details
aws deploy get-deployment \
  --deployment-id d-XXXXXXXXX

# Watch deployment
aws deploy get-deployment \
  --deployment-id d-XXXXXXXXX \
  --query 'deploymentInfo.status' \
  --output text
```

### CloudWatch Logs

```bash
# View ECS task logs
aws logs tail /ecs/aaa-service-dev --follow --filter-pattern "ERROR"
```

### CodeDeploy Console

Navigate to: AWS Console → CodeDeploy → Applications → aaa-service-dev

View:
- Deployment history
- Traffic shifting progress
- Health check status
- Rollback events

## Rollback

### Automatic Rollback

CodeDeploy automatically rolls back if:
- Deployment fails
- CloudWatch alarms trigger
- Health checks fail

### Manual Rollback

```bash
# Stop active deployment and rollback
aws deploy stop-deployment \
  --deployment-id d-XXXXXXXXX \
  --auto-rollback-enabled

# Or redeploy previous version
aws deploy create-deployment \
  --application-name aaa-service-dev \
  --deployment-group-name aaa-service-dev-dg \
  --revision revisionType=S3,s3Location={bucket=my-bucket,key=previous.zip,bundleType=zip}
```

## Lifecycle Hooks (Optional)

Add Lambda functions for automated testing during deployment:

### Pre-Deployment Validation (BeforeInstall)
- Verify database connectivity
- Check service dependencies
- Validate configuration

### Health Check Validation (AfterAllowTestTraffic)
- Run smoke tests against test target group
- Validate API endpoints
- Check critical workflows

### Post-Deployment Validation (AfterAllowTraffic)
- Run integration tests
- Verify metrics
- Send notifications

Example Lambda hook:

```python
import boto3
import requests

def lambda_handler(event, context):
    # Get deployment info
    deployment_id = event['DeploymentId']
    lifecycle_event = event['LifecycleEventHookExecutionId']

    # Run health check
    response = requests.get('http://test-endpoint/health')

    codedeploy = boto3.client('codedeploy')

    if response.status_code == 200:
        # Success - continue deployment
        codedeploy.put_lifecycle_event_hook_execution_status(
            deploymentId=deployment_id,
            lifecycleEventHookExecutionId=lifecycle_event,
            status='Succeeded'
        )
    else:
        # Failure - trigger rollback
        codedeploy.put_lifecycle_event_hook_execution_status(
            deploymentId=deployment_id,
            lifecycleEventHookExecutionId=lifecycle_event,
            status='Failed'
        )
```

## Troubleshooting

### Deployment Fails Immediately

**Issue**: Deployment fails before traffic shift

**Solutions**:
1. Check ECS service uses CODE_DEPLOY controller
2. Verify task definition is valid
3. Check IAM roles have correct permissions
4. Ensure target groups are healthy

### Deployment Stuck in Progress

**Issue**: Traffic shift never completes

**Solutions**:
1. Check health checks on test target group
2. Verify new tasks are running and healthy
3. Review CloudWatch logs for errors
4. Check security group allows ALB → ECS traffic

### Automatic Rollback Triggered

**Issue**: Deployment rolls back automatically

**Solutions**:
1. Check CloudWatch alarms
2. Review target health metrics
3. Check application logs for errors
4. Verify database connectivity from new tasks

### Old Tasks Not Terminating

**Issue**: Blue tasks remain after deployment

**Solutions**:
1. Check TerminationWaitTime parameter
2. Verify deployment completed successfully
3. Manually terminate old task set if needed

## Cost Considerations

Blue-green deployments temporarily run **2x tasks**:

- During deployment: BLUE tasks + GREEN tasks
- Cost impact: ~5-10 minutes of double capacity per deployment
- Example: 2 tasks normally = 4 tasks during deployment

To optimize:
- Use shorter TerminationWaitTime for non-prod
- Deploy during off-peak hours
- Use Fargate Spot for staging/dev

## Best Practices

1. **Test in Lower Environments First**
   - Always deploy to dev → staging → prod
   - Never deploy directly to production

2. **Use Gradual Traffic Shifting in Production**
   - Canary or Linear strategies
   - Gives time to catch issues before full rollout

3. **Monitor During Deployments**
   - Watch CloudWatch metrics
   - Check application logs
   - Verify health checks

4. **Set Up Lifecycle Hooks**
   - Automate validation
   - Catch issues early
   - Reduce manual testing

5. **Configure Alarms**
   - Set up meaningful alerts
   - Trigger automatic rollbacks
   - Get notified of issues

6. **Document Rollback Procedures**
   - Have a runbook ready
   - Practice rollbacks in lower environments
   - Know how to stop deployments quickly

## Next Steps

1. ✅ Deploy CodeDeploy CloudFormation stack
2. ✅ Update CodePipeline with Deploy stage
3. ✅ Configure build environment variables
4. ✅ Test deployment in dev environment
5. ✅ Set up lifecycle hooks (optional)
6. ✅ Configure CloudWatch alarms
7. ✅ Roll out to staging and production

After setup, every push to main will automatically deploy with zero downtime! 🚀
