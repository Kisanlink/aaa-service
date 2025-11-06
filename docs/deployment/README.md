# Deployment Documentation

Complete guide to deploying and operating the AAA service on AWS.

## Getting Started

### Quick Start

1. **[Architecture Overview](ARCHITECTURE.md)** - Understand the system design
2. **[Deployment Guide](DEPLOYMENT.md)** - Deploy infrastructure to AWS
3. **[CI/CD Setup](CODEDEPLOY_SETUP.md)** - Configure automated deployments

## Documentation Index

### Infrastructure

- **[Architecture](ARCHITECTURE.md)**
  - System architecture and design principles
  - Component relationships
  - Technology stack details

- **[Deployment Guide](DEPLOYMENT.md)**
  - CloudFormation infrastructure setup
  - Cost optimization strategies
  - Environment configuration
  - Security best practices

- **[Task Definition Guide](TASK_DEFINITION_README.md)**
  - Manual ECS deployment reference
  - Task definition customization
  - Direct deployment without CloudFormation

### CI/CD Pipeline

- **[CI/CD Overview](CICD_README.md)**
  - Pipeline architecture diagram
  - End-to-end workflow
  - Quick start guide
  - Cost estimates

- **[CodeDeploy Setup](CODEDEPLOY_SETUP.md)**
  - Blue-green deployment configuration
  - Traffic shifting strategies
  - Rollback procedures
  - Troubleshooting guide

### Reference

- **[RBAC Hierarchy Matrix](AAA_RBAC_HIERARCHY_MATRIX.md)**
  - Role inheritance structure
  - Permission levels
  - Organization and group relationships

## Deployment Options

### Option 1: Full CloudFormation Stack (Recommended)

Complete infrastructure deployment with all components:

```bash
aws cloudformation create-stack \
  --stack-name aaa-service-dev \
  --template-body file://../../cloudformation.yaml \
  --parameters ParameterKey=Environment,ParameterValue=dev \
  --capabilities CAPABILITY_NAMED_IAM
```

**Includes**: VPC, ECS, RDS, Redis, ALB, IAM, CloudWatch, Secrets Manager

**Best for**: Production deployments, new environments

### Option 2: Manual ECS Deployment

Direct task definition registration for existing infrastructure:

```bash
# See TASK_DEFINITION_README.md for detailed instructions
aws ecs register-task-definition --cli-input-json file://task-definition-dev.json
```

**Best for**: Quick updates, testing, existing infrastructure

### Option 3: CI/CD Pipeline

Automated blue-green deployments via CodePipeline:

```bash
# See CODEDEPLOY_SETUP.md for setup
git push origin main  # Triggers automated deployment
```

**Best for**: Ongoing development, production operations

## Cost Estimates

| Environment | Infrastructure | CI/CD | Total/Month |
|-------------|---------------|-------|-------------|
| **Dev** | $50-80 | $7-16 | **$57-96** |
| **Staging** | $60-90 | $7-16 | **$67-106** |
| **Production** | $150-200 | $7-16 | **$157-216** |

### Cost Optimization

- Use `EnableMultiAZ=false` for dev/staging
- Single NAT Gateway saves ~$32/month
- Fargate Spot for non-production
- Right-size RDS and Redis instances

See [Deployment Guide](DEPLOYMENT.md) for detailed cost breakdown.

## Environment Configuration

### Development
- Single-AZ deployment
- Smaller instance sizes
- Auto-migrate and seed enabled
- Swagger docs enabled

### Staging
- Single-AZ deployment
- Medium instance sizes
- Auto-migrate enabled
- Docs enabled

### Production
- Multi-AZ deployment
- Large instance sizes
- Manual migrations
- Docs disabled
- Enhanced monitoring

## Support

### Common Issues

- **Deployment Failures**: See [Deployment Guide](DEPLOYMENT.md#troubleshooting)
- **Pipeline Issues**: See [CodeDeploy Setup](CODEDEPLOY_SETUP.md#troubleshooting)
- **gRPC Connection**: Requires HTTPS certificate (see Architecture)

### Getting Help

1. Check troubleshooting sections in respective guides
2. Review CloudWatch logs: `aws logs tail /ecs/aaa-service-{env} --follow`
3. Verify security groups and IAM roles
4. Check AWS service health dashboard

## Additional Resources

- [Main README](../../README.md) - Project overview
- [API Documentation](../API_EXAMPLES.md) - API usage examples
- [Implementation History](../implementation/) - Historical changes
