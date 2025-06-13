from aws_cdk import (
    Stack,
    aws_iam as iam,
    CfnOutput
)
from constructs import Construct

class IamStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        # Create IAM role for database access
        db_access_role = iam.Role(
            self, "DatabaseAccessRole",
            assumed_by=iam.ServicePrincipal("rds.amazonaws.com"),
            description="Role for database access and management",
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonRDSEnhancedMonitoringRole"),
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonRDSDirectoryServiceAccess")
            ]
        )

        # Add custom policy for database operations
        db_access_role.add_to_policy(
            iam.PolicyStatement(
                effect=iam.Effect.ALLOW,
                actions=[
                    "rds-db:connect",
                    "rds:DescribeDBInstances",
                    "rds:DescribeDBClusters",
                    "rds:ModifyDBInstance",
                    "rds:ModifyDBCluster",
                    "rds:StartDBInstance",
                    "rds:StopDBInstance",
                    "rds:RebootDBInstance",
                    "rds:CreateDBInstance",
                    "rds:DeleteDBInstance",
                    "rds:CreateDBCluster",
                    "rds:DeleteDBCluster",
                    "rds:RestoreDBInstanceFromDBSnapshot",
                    "rds:RestoreDBClusterFromSnapshot",
                    "rds:CreateDBSnapshot",
                    "rds:DeleteDBSnapshot",
                    "rds:CopyDBSnapshot",
                    "rds:CopyDBClusterSnapshot",
                    "rds:DescribeDBSnapshots",
                    "rds:DescribeDBClusterSnapshots",
                    "rds:ModifyDBSnapshotAttribute",
                    "rds:ModifyDBClusterSnapshotAttribute",
                    "rds:RestoreDBInstanceToPointInTime",
                    "rds:RestoreDBClusterToPointInTime",
                    "rds:DescribeDBEngineVersions",
                    "rds:DescribeOrderableDBInstanceOptions",
                    "rds:DescribeDBParameters",
                    "rds:DescribeDBClusterParameters",
                    "rds:ModifyDBParameterGroup",
                    "rds:ModifyDBClusterParameterGroup",
                    "rds:DescribeEvents",
                    "rds:DescribeEventSubscriptions",
                    "rds:CreateEventSubscription",
                    "rds:DeleteEventSubscription",
                    "rds:ModifyEventSubscription",
                    "rds:AddTagsToResource",
                    "rds:RemoveTagsFromResource",
                    "rds:ListTagsForResource",
                    "rds:DescribeDBLogFiles",
                    "rds:DownloadDBLogFilePortion",
                    "rds:DescribeDBInstances",
                    "rds:DescribeDBClusters",
                    "rds:DescribeDBEngineVersions",
                    "rds:DescribeDBParameterGroups",
                    "rds:DescribeDBClusterParameterGroups",
                    "rds:DescribeDBParameters",
                    "rds:DescribeDBClusterParameters",
                    "rds:DescribeOptionGroups",
                    "rds:DescribeDBSubnetGroups",
                    "rds:DescribeEventSubscriptions",
                    "rds:DescribeEvents",
                    "rds:DescribeOrderableDBInstanceOptions",
                    "rds:DescribePendingMaintenanceActions",
                    "rds:DescribeReservedDBInstances",
                    "rds:DescribeReservedDBInstancesOfferings",
                    "rds:DescribeSourceRegions",
                    "rds:DescribeValidDBInstanceModifications",
                    "rds:DescribeValidDBClusterModifications",
                    "rds:DescribeDBClusterSnapshotAttributes",
                    "rds:DescribeDBSnapshotAttributes",
                    "rds:DescribeDBClusterEndpoints",
                    "rds:DescribeDBClusterParameterGroups",
                    "rds:DescribeDBClusterParameters",
                    "rds:DescribeDBClusterSnapshots",
                    "rds:DescribeDBClusters",
                    "rds:DescribeDBInstances",
                    "rds:DescribeDBParameterGroups",
                    "rds:DescribeDBParameters",
                    "rds:DescribeDBSnapshots",
                    "rds:DescribeDBSubnetGroups",
                    "rds:DescribeEventCategories",
                    "rds:DescribeEventSubscriptions",
                    "rds:DescribeEvents",
                    "rds:DescribeOptionGroups",
                    "rds:DescribeOrderableDBInstanceOptions",
                    "rds:DescribePendingMaintenanceActions",
                    "rds:DescribeReservedDBInstances",
                    "rds:DescribeReservedDBInstancesOfferings",
                    "rds:DescribeSourceRegions",
                    "rds:DescribeValidDBInstanceModifications",
                    "rds:DescribeValidDBClusterModifications",
                    "rds:ListTagsForResource",
                    "rds:ModifyDBCluster",
                    "rds:ModifyDBClusterParameterGroup",
                    "rds:ModifyDBClusterSnapshotAttribute",
                    "rds:ModifyDBInstance",
                    "rds:ModifyDBParameterGroup",
                    "rds:ModifyDBSnapshotAttribute",
                    "rds:ModifyEventSubscription",
                    "rds:PromoteReadReplica",
                    "rds:PromoteReadReplicaDBCluster",
                    "rds:PurchaseReservedDBInstancesOffering",
                    "rds:RebootDBInstance",
                    "rds:RemoveTagsFromResource",
                    "rds:ResetDBClusterParameterGroup",
                    "rds:ResetDBParameterGroup",
                    "rds:RestoreDBClusterFromSnapshot",
                    "rds:RestoreDBClusterToPointInTime",
                    "rds:RestoreDBInstanceFromDBSnapshot",
                    "rds:RestoreDBInstanceToPointInTime",
                    "rds:RevokeDBSecurityGroupIngress",
                    "rds:StartDBCluster",
                    "rds:StartDBInstance",
                    "rds:StopDBCluster",
                    "rds:StopDBInstance"
                ],
                resources=["*"]
            )
        )

        # Create IAM role for database monitoring
        monitoring_role = iam.Role(
            self, "DatabaseMonitoringRole",
            assumed_by=iam.ServicePrincipal("monitoring.rds.amazonaws.com"),
            description="Role for database monitoring and metrics",
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonRDSEnhancedMonitoringRole")
            ]
        )

        # Create IAM role for database backup
        backup_role = iam.Role(
            self, "DatabaseBackupRole",
            assumed_by=iam.ServicePrincipal("backup.amazonaws.com"),
            description="Role for database backup operations",
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AWSBackupServiceRolePolicyForBackup")
            ]
        )

        # Output the role ARNs
        CfnOutput(
            self, "DatabaseAccessRoleARN",
            value=db_access_role.role_arn,
            description="Database access role ARN"
        )

        CfnOutput(
            self, "DatabaseMonitoringRoleARN",
            value=monitoring_role.role_arn,
            description="Database monitoring role ARN"
        )

        CfnOutput(
            self, "DatabaseBackupRoleARN",
            value=backup_role.role_arn,
            description="Database backup role ARN"
        ) 