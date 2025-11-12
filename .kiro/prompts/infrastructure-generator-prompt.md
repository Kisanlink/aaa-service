# Infrastructure Generator Prompt

You are an expert DevOps/SRE engineer tasked with analyzing a microservice codebase and generating production-ready AWS infrastructure using CloudFormation, CodeDeploy, and ECS Fargate.

## Objective

Analyze the codebase and generate a complete, working infrastructure setup that:
1. Follows AWS best practices
2. Is consistent with the Kisanlink infrastructure patterns
3. Requires minimal iteration and troubleshooting
4. Works on the first deployment

## Analysis Phase

### Step 1: Codebase Analysis
Perform a thorough analysis of the codebase to understand:

1. **Application Type & Runtime**
   - Language and version (Go, Node.js, Python, etc.)
   - Framework (Gin, Express, FastAPI, etc.)
   - Build system (go build, npm, pip, etc.)
   - Dependencies and package management

2. **Service Dependencies**
   - Database requirements (PostgreSQL, MySQL, etc.)
   - Cache requirements (Redis, Memcached, etc.)
   - Message queues (SQS, SNS, RabbitMQ, etc.)
   - External AWS services (S3, SES, etc.)
   - Third-party APIs

3. **Configuration Requirements**
   - Environment variables needed
   - Secrets (database passwords, API keys, etc.)
   - Feature flags or configuration files
   - Certificates or keys

4. **Health & Readiness**
   - Health check endpoint (path, expected response)
   - Startup time estimation
   - Graceful shutdown requirements
   - Database migration strategy

5. **Resource Requirements**
   - Memory needs (analyze for: ORM, caching, concurrent connections)
   - CPU needs (analyze for: background workers, heavy computation)
   - Storage needs (logs, temporary files)
   - Network bandwidth

6. **Security Requirements**
   - TLS/SSL needs
   - Authentication mechanisms
   - Authorization patterns
   - Secrets rotation
   - Network isolation needs

### Step 2: Document Findings

Create a summary document with:
```markdown
# Infrastructure Requirements for [SERVICE-NAME]

## Service Profile
- **Type**: [REST API / gRPC / Worker / etc.]
- **Runtime**: [Go 1.22 / Node.js 20 / etc.]
- **Port**: [8080 / 3000 / etc.]
- **Health Endpoint**: [/health]

## Dependencies
- **Database**: [PostgreSQL 15 with connection pooling]
- **Cache**: [Redis 7 with TLS enabled]
- **External Services**: [List with authentication methods]

## Resource Allocation
- **CPU**: [512 / 1024 / 2048]
- **Memory**: [1024 / 2048 / 4096]
- **Justification**: [Based on X concurrent connections, Y MB cache, Z background workers]

## Environment Variables Required
[List all env vars with descriptions and example values]

## Secrets Required
[List all secrets that need to be in Secrets Manager]

## Startup Characteristics
- **Estimated Startup Time**: [30s / 60s / 120s]
- **Migration Strategy**: [Auto-migration / Separate job]
- **Health Check Grace Period**: [120s / 180s / 240s]
```

## Generation Phase

### Step 3: Generate CloudFormation Template

Generate `cloudformation.yaml` with the following sections:

#### Parameters
- Environment (beta/prod)
- VPC configuration
- Database endpoints (if external)
- Service version/tag
- Resource sizing (CPU, Memory)
- Auto-scaling parameters

#### Key Components to Include

1. **Application Load Balancer**
   ```yaml
   - ALB with HTTPS listener (certificate from ACM)
   - Target group with proper health checks:
     - Path: [from analysis]
     - Matcher: 200,401 (if auth required)
     - Interval: 30s
     - Timeout: 5s
     - HealthyThreshold: 2
     - UnhealthyThreshold: 3
   - Security group allowing 80/443
   ```

2. **ECS Cluster & Service**
   ```yaml
   - Fargate cluster
   - Service with:
     - DesiredCount: 1 (beta), 2+ (prod)
     - DeploymentConfiguration:
       - MinimumHealthyPercent: 100
       - MaximumPercent: 200
     - NetworkConfiguration:
       - Subnets: Private subnets
       - SecurityGroups: Application SG
   ```

3. **Task Definition**
   ```yaml
   - Family: [service-name]-[env]
   - NetworkMode: awsvpc
   - RequiresCompatibilities: FARGATE
   - Cpu: [from analysis]
   - Memory: [from analysis]
   - ContainerDefinitions:
     - Image: !Sub ${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/[repo]:${ImageTag}
     - PortMappings: [from analysis]
     - Environment: [non-secret vars]
     - Secrets: [from Secrets Manager]
     - LogConfiguration:
       - CloudWatch Logs with proper retention
     - HealthCheck: NONE (rely on ALB only if migrations run on startup)
   ```

4. **Database (if needed)**
   ```yaml
   - RDS Instance (or Aurora)
   - Subnet group in private subnets
   - Security group allowing access from app SG
   - Automated backups enabled
   - Encryption at rest
   - Parameter group with optimizations
   ```

5. **ElastiCache Redis (if needed)**
   ```yaml
   - Redis cluster
   - Subnet group in private subnets
   - Security group allowing access from app SG
   - TransitEncryptionEnabled: true
   - AtRestEncryptionEnabled: true
   - Auth token in Secrets Manager
   - Automatic failover (if Multi-AZ)
   ```

6. **Security Groups**
   ```yaml
   - ALB SG: Allow 80/443 from 0.0.0.0/0
   - Application SG: Allow app port from ALB SG
   - Database SG: Allow DB port from Application SG
   - Redis SG: Allow 6379 from Application SG
   ```

7. **IAM Roles**
   ```yaml
   - Task Execution Role: ECR, CloudWatch Logs, Secrets Manager
   - Task Role: S3, SES, SQS, etc. (based on analysis)
   ```

8. **Outputs**
   ```yaml
   - ALB DNS Name
   - ECS Cluster ARN
   - ECS Service ARN
   - Database Endpoint (if created)
   - Redis Endpoint (if created)
   ```

#### Critical CloudFormation Patterns

1. **Resource Naming Convention**
   ```yaml
   !Sub "${ServiceName}-${Environment}-[resource-type]"
   ```

2. **Tags on All Resources**
   ```yaml
   Tags:
     - Key: Environment
       Value: !Ref Environment
     - Key: Service
       Value: !Ref ServiceName
     - Key: ManagedBy
       Value: CloudFormation
   ```

3. **DependsOn Chains**
   - Ensure proper dependency ordering
   - Database → Application
   - Cache → Application
   - Target Group → Load Balancer Listener

### Step 4: Generate buildspec.yml

```yaml
version: 0.2

phases:
  pre_build:
    commands:
      - echo Logging in to Amazon ECR...
      - aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com
      - REPOSITORY_URI=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/[service-name]
      - COMMIT_HASH=$(echo $CODEBUILD_RESOLVED_SOURCE_VERSION | cut -c 1-7)
      - IMAGE_TAG=${COMMIT_HASH:=latest}

  build:
    commands:
      - echo Build started on `date`
      - echo Building the Docker image...
      - docker build -t $REPOSITORY_URI:latest .
      - docker tag $REPOSITORY_URI:latest $REPOSITORY_URI:$IMAGE_TAG

  post_build:
    commands:
      - echo Build completed on `date`
      - echo Pushing the Docker images...
      - docker push $REPOSITORY_URI:latest
      - docker push $REPOSITORY_URI:$IMAGE_TAG
      - echo Writing image definitions file...
      - printf '[{"name":"[container-name]","imageUri":"%s"}]' $REPOSITORY_URI:$IMAGE_TAG > imagedefinitions.json

artifacts:
  files:
    - imagedefinitions.json
    - appspec.yaml
    - taskdef.json
```

**Customize based on:**
- Language-specific build commands
- Test running requirements
- Multi-stage builds

### Step 5: Generate appspec.yaml

```yaml
version: 0.0
Resources:
  - TargetService:
      Type: AWS::ECS::Service
      Properties:
        TaskDefinition: <TASK_DEFINITION>
        LoadBalancerInfo:
          ContainerName: "[container-name]"
          ContainerPort: [port]
Hooks:
  - BeforeInstall: "[lambda-arn-if-needed]"
  - AfterInstall: "[lambda-arn-if-needed]"
  - AfterAllowTestTraffic: "[lambda-arn-if-needed]"
  - BeforeAllowTraffic: "[lambda-arn-if-needed]"
  - AfterAllowTraffic: "[lambda-arn-if-needed]"
```

### Step 6: Generate taskdef.json Template

Create a template that matches the CloudFormation task definition exactly.

### Step 7: Generate Dockerfile (if missing)

Analyze the language and create an optimized multi-stage Dockerfile:

**For Go:**
```dockerfile
FROM golang:[version]-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE [port]
CMD ["./main"]
```

**For Node.js:**
```dockerfile
FROM node:[version]-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .

FROM node:[version]-alpine
WORKDIR /app
COPY --from=builder /app .
EXPOSE [port]
CMD ["node", "index.js"]
```

### Step 8: Generate deployment_guide.md

Create comprehensive deployment instructions including:
- Prerequisites
- AWS CLI commands to create stack
- Environment-specific parameters
- Rollback procedures
- Troubleshooting common issues

## Validation Checklist

Before finalizing, verify:

- [ ] All environment variables are documented and sourced
- [ ] All secrets are in Secrets Manager (not hardcoded)
- [ ] Health check timeout > startup time + migration time
- [ ] ALB target group allows authentication responses (200,401)
- [ ] Container has NO health check if migrations run on startup
- [ ] Database security group allows app access
- [ ] Redis has TLS enabled if TransitEncryptionEnabled
- [ ] Application code configured for TLS (if Redis encrypted)
- [ ] IAM roles have minimal required permissions
- [ ] CloudWatch logs have retention period set
- [ ] Auto-scaling policies defined (if needed)
- [ ] DependsOn chains prevent creation order issues
- [ ] All resources properly tagged
- [ ] Outputs include all necessary endpoints

## Common Pitfalls to Avoid

1. **Health Check Issues**
   - Container health check killing tasks during migrations
   - Health check timeout < startup time
   - Not allowing 401 responses for authenticated endpoints

2. **TLS Configuration**
   - ElastiCache TLS enabled but application not configured
   - Missing TLS client configuration in code

3. **Resource Sizing**
   - Undersized memory causing OOM kills
   - Undersized CPU causing slow startup

4. **Networking**
   - Tasks in public subnets without proper route tables
   - Security groups not allowing necessary traffic

5. **Secrets Management**
   - Hardcoded secrets in task definition
   - Not using Secrets Manager references

6. **Dependencies**
   - Missing DependsOn causing race conditions
   - Services starting before dependencies ready

## Output Format

Provide the following files:
1. `cloudformation.yaml` - Complete infrastructure template
2. `buildspec.yml` - CodeBuild specification
3. `appspec.yaml` - CodeDeploy specification
4. `taskdef.json` - Task definition template
5. `Dockerfile` - If not present or needs optimization
6. `deployment_guide.md` - Step-by-step deployment instructions
7. `infrastructure_analysis.md` - The analysis from Step 2
8. `.env.example` - Example environment variables

## Success Criteria

The infrastructure should:
1. Deploy successfully on first attempt
2. Service becomes healthy within expected timeframe
3. All dependencies accessible and working
4. No manual intervention required post-deployment
5. Logs available in CloudWatch
6. Health checks passing
7. Application functioning correctly
