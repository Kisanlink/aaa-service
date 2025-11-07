.PHONY: help deploy-beta deploy-staging deploy-prod delete-beta delete-staging delete-prod status-beta status-staging status-prod validate

# AWS Configuration
AWS_REGION ?= ap-south-1
STACK_PREFIX = aaa-service
TEMPLATE_FILE = ../cloudformation.yaml

# Helper function to get stack status
define check_stack
	@aws cloudformation describe-stacks \
		--stack-name $(1) \
		--region $(AWS_REGION) \
		--query 'Stacks[0].StackStatus' \
		--output text 2>/dev/null || echo "DOES_NOT_EXIST"
endef

help:
	@echo "AAA Service Infrastructure Deployment"
	@echo "======================================"
	@echo ""
	@echo "Available targets:"
	@echo "  deploy-beta       - Deploy beta environment"
	@echo "  deploy-staging    - Deploy staging environment"
	@echo "  deploy-prod       - Deploy production environment"
	@echo ""
	@echo "  status-beta       - Check beta stack status"
	@echo "  status-staging    - Check staging stack status"
	@echo "  status-prod       - Check production stack status"
	@echo ""
	@echo "  delete-beta       - Delete beta stack"
	@echo "  delete-staging    - Delete staging stack"
	@echo "  delete-prod       - Delete production stack"
	@echo ""
	@echo "  validate          - Validate CloudFormation template"
	@echo ""
	@echo "Environment variables:"
	@echo "  AWS_REGION        - AWS region (default: ap-south-1)"

validate:
	@echo "Validating CloudFormation template..."
	@aws cloudformation validate-template \
		--template-body file://$(TEMPLATE_FILE) \
		--region $(AWS_REGION)
	@echo "✅ Template is valid"

# Beta Environment
deploy-beta: validate
	@echo "Deploying beta environment..."
	@aws cloudformation deploy \
		--stack-name $(STACK_PREFIX)-beta \
		--template-file $(TEMPLATE_FILE) \
		--parameter-overrides file://parameters/beta.json \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION) \
		--no-fail-on-empty-changeset
	@echo "✅ Beta deployment complete"
	@$(MAKE) status-beta

status-beta:
	@echo "Beta Stack Status:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-beta \
		--region $(AWS_REGION) \
		--query 'Stacks[0].{Status:StackStatus,Updated:LastUpdatedTime}' \
		--output table 2>/dev/null || echo "Stack does not exist"
	@echo ""
	@echo "Beta Stack Outputs:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-beta \
		--region $(AWS_REGION) \
		--query 'Stacks[0].Outputs[?OutputKey==`ALBDNS` || OutputKey==`HTTPEndpoint`]' \
		--output table 2>/dev/null || echo "No outputs available"

delete-beta:
	@echo "⚠️  Deleting beta stack..."
	@read -p "Are you sure you want to delete the beta stack? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		aws cloudformation delete-stack \
			--stack-name $(STACK_PREFIX)-beta \
			--region $(AWS_REGION); \
		echo "Waiting for stack deletion..."; \
		aws cloudformation wait stack-delete-complete \
			--stack-name $(STACK_PREFIX)-beta \
			--region $(AWS_REGION); \
		echo "✅ Beta stack deleted"; \
	else \
		echo "Cancelled"; \
	fi

# Staging Environment
deploy-staging: validate
	@echo "Deploying staging environment..."
	@aws cloudformation deploy \
		--stack-name $(STACK_PREFIX)-staging \
		--template-file $(TEMPLATE_FILE) \
		--parameter-overrides file://parameters/staging.json \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION) \
		--no-fail-on-empty-changeset
	@echo "✅ Staging deployment complete"
	@$(MAKE) status-staging

status-staging:
	@echo "Staging Stack Status:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-staging \
		--region $(AWS_REGION) \
		--query 'Stacks[0].{Status:StackStatus,Updated:LastUpdatedTime}' \
		--output table 2>/dev/null || echo "Stack does not exist"
	@echo ""
	@echo "Staging Stack Outputs:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-staging \
		--region $(AWS_REGION) \
		--query 'Stacks[0].Outputs[?OutputKey==`ALBDNS` || OutputKey==`HTTPEndpoint`]' \
		--output table 2>/dev/null || echo "No outputs available"

delete-staging:
	@echo "⚠️  Deleting staging stack..."
	@read -p "Are you sure you want to delete the staging stack? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		aws cloudformation delete-stack \
			--stack-name $(STACK_PREFIX)-staging \
			--region $(AWS_REGION); \
		echo "Waiting for stack deletion..."; \
		aws cloudformation wait stack-delete-complete \
			--stack-name $(STACK_PREFIX)-staging \
			--region $(AWS_REGION); \
		echo "✅ Staging stack deleted"; \
	else \
		echo "Cancelled"; \
	fi

# Production Environment
deploy-prod: validate
	@echo "⚠️  Deploying PRODUCTION environment..."
	@read -p "Are you sure you want to deploy to PRODUCTION? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		aws cloudformation deploy \
			--stack-name $(STACK_PREFIX)-prod \
			--template-file $(TEMPLATE_FILE) \
			--parameter-overrides file://parameters/prod.json \
			--capabilities CAPABILITY_NAMED_IAM \
			--region $(AWS_REGION) \
			--no-fail-on-empty-changeset; \
		echo "✅ Production deployment complete"; \
		$(MAKE) status-prod; \
	else \
		echo "Cancelled"; \
	fi

status-prod:
	@echo "Production Stack Status:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-prod \
		--region $(AWS_REGION) \
		--query 'Stacks[0].{Status:StackStatus,Updated:LastUpdatedTime}' \
		--output table 2>/dev/null || echo "Stack does not exist"
	@echo ""
	@echo "Production Stack Outputs:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_PREFIX)-prod \
		--region $(AWS_REGION) \
		--query 'Stacks[0].Outputs[?OutputKey==`ALBDNS` || OutputKey==`HTTPEndpoint`]' \
		--output table 2>/dev/null || echo "No outputs available"

delete-prod:
	@echo "🛑 DANGER: Deleting PRODUCTION stack..."
	@echo "This will destroy all production data!"
	@read -p "Type 'DELETE-PRODUCTION' to confirm: " confirm; \
	if [ "$$confirm" = "DELETE-PRODUCTION" ]; then \
		aws cloudformation delete-stack \
			--stack-name $(STACK_PREFIX)-prod \
			--region $(AWS_REGION); \
		echo "Waiting for stack deletion..."; \
		aws cloudformation wait stack-delete-complete \
			--stack-name $(STACK_PREFIX)-prod \
			--region $(AWS_REGION); \
		echo "✅ Production stack deleted"; \
	else \
		echo "Cancelled - confirmation did not match"; \
	fi
