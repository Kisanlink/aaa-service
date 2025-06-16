from aws_cdk import (
    Stack,
    aws_ecr as ecr,
    aws_iam as iam,
    CfnOutput,
    RemovalPolicy,
    Duration
)
from constructs import Construct

class ContainerStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        # Create ECR repositories
        aaa_service_repo = ecr.Repository(
            self, "AaaServiceRepository",
            repository_name="aaa-service",
            removal_policy=RemovalPolicy.RETAIN,
            image_scan_on_push=True,
            lifecycle_rules=[
                ecr.LifecycleRule(
                    max_image_count=30,
                    description="Keep only the last 30 images"
                )
            ]
        )

        spicedb_repo = ecr.Repository(
            self, "SpiceDBRepository",
            repository_name="spicedb",
            removal_policy=RemovalPolicy.RETAIN,
            image_scan_on_push=True,
            lifecycle_rules=[
                ecr.LifecycleRule(
                    max_image_count=10,
                    description="Keep only the last 10 images"
                )
            ]
        )

        aadhaar_validation_repo = ecr.Repository(
            self, "AadhaarValidationRepository",
            repository_name="aadhaar-validation-service",
            removal_policy=RemovalPolicy.RETAIN,
            image_scan_on_push=True,
            lifecycle_rules=[
                ecr.LifecycleRule(
                    max_image_count=30,
                    description="Keep only the last 30 images"
                )
            ]
        )

        # Create IAM role for ECR access
        ecr_access_role = iam.Role(
            self, "EcrAccessRole",
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
            description="Role for ECR access from ECS tasks"
        )

        # Add ECR pull policy
        ecr_access_role.add_to_policy(
            iam.PolicyStatement(
                effect=iam.Effect.ALLOW,
                actions=[
                    "ecr:GetDownloadUrlForLayer",
                    "ecr:BatchGetImage",
                    "ecr:BatchCheckLayerAvailability",
                    "ecr:GetAuthorizationToken"
                ],
                resources=[
                    aaa_service_repo.repository_arn,
                    spicedb_repo.repository_arn,
                    aadhaar_validation_repo.repository_arn
                ]
            )
        )

        # Output repository URIs
        CfnOutput(
            self, "AaaServiceRepositoryURI",
            value=aaa_service_repo.repository_uri,
            description="AAA Service ECR repository URI"
        )

        CfnOutput(
            self, "SpiceDBRepositoryURI",
            value=spicedb_repo.repository_uri,
            description="SpiceDB ECR repository URI"
        )

        CfnOutput(
            self, "AadhaarValidationRepositoryURI",
            value=aadhaar_validation_repo.repository_uri,
            description="Aadhaar Validation Service ECR repository URI"
        ) 