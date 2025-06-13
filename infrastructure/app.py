#!/usr/bin/env python3
from aws_cdk import App
from stacks.database_stack import DatabaseStack
from stacks.iam_stack import IamStack
from stacks.container_stack import ContainerStack

app = App()

# Create IAM stack first
iam_stack = IamStack(app, "AaaServiceIamStack",
    env={
        "region": "ap-south-1",  # Mumbai region
        "account": app.node.try_get_context("account")
    }
)

# Create database stack with IAM role references
DatabaseStack(app, "AaaServiceDatabaseStack",
    env={
        "region": "ap-south-1",  # Mumbai region
        "account": app.node.try_get_context("account")
    }
)

# Create container stack
ContainerStack(app, "AaaServiceContainerStack",
    env={
        "region": "ap-south-1",  # Mumbai region
        "account": app.node.try_get_context("account")
    }
)

app.synth() 