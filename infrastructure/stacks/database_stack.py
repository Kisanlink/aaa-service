from aws_cdk import (
    Stack,
    aws_ec2 as ec2,
    aws_rds as rds,
    aws_secretsmanager as secretsmanager,
    aws_iam as iam,
    CfnOutput,
    Duration,
    RemovalPolicy,
)
from constructs import Construct

class DatabaseStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        # Create VPC
        vpc = ec2.Vpc(
            self, "DatabaseVPC",
            max_azs=2,
            nat_gateways=1,
            subnet_configuration=[
                ec2.SubnetConfiguration(
                    name="Public",
                    subnet_type=ec2.SubnetType.PUBLIC,
                    cidr_mask=24
                ),
                ec2.SubnetConfiguration(
                    name="Private",
                    subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS,
                    cidr_mask=24
                )
            ]
        )

        # Create security group for Aurora
        db_security_group = ec2.SecurityGroup(
            self, "DatabaseSecurityGroup",
            vpc=vpc,
            description="Security group for Aurora PostgreSQL",
            allow_all_outbound=True
        )

        # Allow inbound PostgreSQL traffic from VPC
        db_security_group.add_ingress_rule(
            ec2.Peer.ipv4(vpc.vpc_cidr_block),
            ec2.Port.tcp(5432),
            "Allow PostgreSQL traffic from VPC"
        )

        # Create database credentials in Secrets Manager
        db_credentials = secretsmanager.Secret(
            self, "DatabaseCredentials",
            generate_secret_string=secretsmanager.SecretStringGenerator(
                secret_string_template='{"username": "aaa_user"}',
                generate_string_key="password",
                exclude_characters='"@/\\'
            )
        )

        # Get the IAM roles from the IAM stack
        db_access_role = iam.Role.from_role_arn(
            self, "ImportedDatabaseAccessRole",
            role_arn=self.node.try_get_context("database_access_role_arn")
        )

        monitoring_role = iam.Role.from_role_arn(
            self, "ImportedMonitoringRole",
            role_arn=self.node.try_get_context("monitoring_role_arn")
        )

        backup_role = iam.Role.from_role_arn(
            self, "ImportedBackupRole",
            role_arn=self.node.try_get_context("backup_role_arn")
        )

        # Create Aurora PostgreSQL cluster
        cluster = rds.DatabaseCluster(
            self, "DatabaseCluster",
            engine=rds.DatabaseClusterEngine.aurora_postgres(
                version=rds.AuroraPostgresEngineVersion.VER_16_1
            ),
            credentials=rds.Credentials.from_secret(db_credentials),
            instance_props=rds.InstanceProps(
                vpc=vpc,
                vpc_subnets=ec2.SubnetSelection(
                    subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
                ),
                security_groups=[db_security_group],
                instance_type=ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.MEDIUM
                ),
                monitoring_role=monitoring_role
            ),
            instances=1,
            backup=rds.BackupProps(
                retention=Duration.days(7),
                preferred_window="03:00-04:00"
            ),
            monitoring_interval=Duration.seconds(60),
            parameter_group=rds.ParameterGroup.from_parameter_group_name(
                self, "ParameterGroup",
                parameter_group_name="default.aurora-postgresql16"
            ),
            removal_policy=RemovalPolicy.SNAPSHOT,
            associated_roles=[db_access_role, backup_role]
        )

        # Output the cluster endpoint
        CfnOutput(
            self, "ClusterEndpoint",
            value=cluster.cluster_endpoint.hostname,
            description="Aurora PostgreSQL cluster endpoint"
        )

        # Output the secret ARN
        CfnOutput(
            self, "SecretARN",
            value=db_credentials.secret_arn,
            description="Database credentials secret ARN"
        ) 