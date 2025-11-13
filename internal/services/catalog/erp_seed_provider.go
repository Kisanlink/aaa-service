package catalog

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// ERPSeedProvider provides embedded seed data for ERP-module roles and permissions
// This is an example provider that demonstrates how other services can register their own roles
type ERPSeedProvider struct {
	*BaseSeedProvider
}

// NewERPSeedProvider creates a new ERP seed data provider
func NewERPSeedProvider() *ERPSeedProvider {
	return &ERPSeedProvider{
		BaseSeedProvider: NewBaseSeedProvider("erp-module", "ERP Module"),
	}
}

// GetResources returns all resources to be seeded for ERP-module
func (s *ERPSeedProvider) GetResources() []ResourceDefinition {
	return []ResourceDefinition{
		// Core ERP Resources
		{Name: "product", Type: "erp/catalog", Description: "Product resource"},
		{Name: "variant", Type: "erp/catalog", Description: "Product variant resource"},
		{Name: "product_price", Type: "erp/catalog", Description: "Product price resource"},
		{Name: "warehouse", Type: "erp/warehouse", Description: "Warehouse resource"},
		{Name: "inventory", Type: "erp/warehouse", Description: "Inventory resource"},
		{Name: "batch", Type: "erp/warehouse", Description: "Inventory batch resource"},
		{Name: "grn", Type: "erp/procurement", Description: "Goods Receipt Note resource"},
		{Name: "purchase_order", Type: "erp/procurement", Description: "Purchase order resource"},
		{Name: "sales", Type: "erp/sales", Description: "Sales transaction resource"},
		{Name: "return", Type: "erp/sales", Description: "Return resource"},
		{Name: "collaborator", Type: "erp/sales", Description: "Collaborator resource"},
		{Name: "collaborator_product", Type: "erp/sales", Description: "Collaborator-product relationship resource"},
		{Name: "tax", Type: "erp/finance", Description: "Tax resource"},
		{Name: "discount", Type: "erp/sales", Description: "Discount resource"},
		{Name: "bank_payment", Type: "erp/finance", Description: "Bank payment resource"},
		{Name: "refund_policy", Type: "erp/sales", Description: "Refund policy resource"},
		{Name: "attachment", Type: "erp/common", Description: "Attachment resource"},
		// Legacy/Additional Resources (keeping for backward compatibility)
		{Name: "invoice", Type: "erp/finance", Description: "Invoice resource"},
		{Name: "vendor", Type: "erp/procurement", Description: "Vendor resource"},
		{Name: "customer", Type: "erp/sales", Description: "Customer resource"},
		{Name: "payment", Type: "erp/finance", Description: "Payment resource"},
		{Name: "ledger", Type: "erp/accounting", Description: "Ledger resource"},
		{Name: "budget", Type: "erp/finance", Description: "Budget resource"},
	}
}

// GetActions returns all actions to be seeded for ERP-module
func (s *ERPSeedProvider) GetActions() []ActionDefinition {
	return []ActionDefinition{
		{Name: "create", Description: "Create a resource", Category: "general", IsStatic: true},
		{Name: "read", Description: "Read a resource", Category: "general", IsStatic: true},
		{Name: "update", Description: "Update a resource", Category: "general", IsStatic: true},
		{Name: "delete", Description: "Delete a resource", Category: "general", IsStatic: true},
		{Name: "list", Description: "List resources", Category: "general", IsStatic: true},
		{Name: "calculate", Description: "Calculate values (e.g., tax calculations)", Category: "general", IsStatic: false},
		{Name: "approve", Description: "Approve a resource", Category: "workflow", IsStatic: false},
		{Name: "reject", Description: "Reject a resource", Category: "workflow", IsStatic: false},
		{Name: "post", Description: "Post to ledger", Category: "accounting", IsStatic: false},
		{Name: "reconcile", Description: "Reconcile accounts", Category: "accounting", IsStatic: false},
		{Name: "export", Description: "Export data", Category: "reporting", IsStatic: false},
	}
}

// GetRoles returns all roles to be seeded with their permissions for ERP-module
func (s *ERPSeedProvider) GetRoles() []RoleDefinition {
	return []RoleDefinition{
		{
			Name:        "erp_accountant",
			Description: "Accountant role with permissions to manage ledgers and reconciliations",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"ledger:create", "ledger:read", "ledger:update", "ledger:list",
				"ledger:post", "ledger:reconcile",
				"invoice:read", "invoice:list",
				"payment:read", "payment:list", "payment:reconcile",
				"bank_payment:read", "bank_payment:list",
			},
		},
		{
			Name:        "erp_finance_manager",
			Description: "Finance Manager role with full financial management permissions",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				// All accountant permissions
				"ledger:*",
				"invoice:*",
				"payment:*",
				"bank_payment:*",
				// Budget management
				"budget:create", "budget:read", "budget:update", "budget:delete", "budget:list",
				"budget:approve", "budget:reject",
				// Reporting
				"*:export",
			},
		},
		{
			Name:        "erp_procurement_officer",
			Description: "Procurement Officer role with vendor and purchase order management",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"vendor:create", "vendor:read", "vendor:update", "vendor:list",
				"purchase_order:create", "purchase_order:read", "purchase_order:update", "purchase_order:list",
				"grn:read", "grn:list",
				"inventory:read", "inventory:list",
			},
		},
		{
			Name:        "erp_warehouse_manager",
			Description: "Warehouse Manager role with inventory management permissions",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"warehouse:create", "warehouse:read", "warehouse:update", "warehouse:delete", "warehouse:list",
				"inventory:create", "inventory:read", "inventory:update", "inventory:delete", "inventory:list",
				"batch:create", "batch:read", "batch:update", "batch:list",
				"grn:create", "grn:read", "grn:update", "grn:list",
				"purchase_order:read", "purchase_order:list", "purchase_order:approve",
			},
		},
		{
			Name:        "erp_sales_manager",
			Description: "Sales Manager role with customer, sales, and invoice management",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"customer:create", "customer:read", "customer:update", "customer:list",
				"collaborator:create", "collaborator:read", "collaborator:update", "collaborator:delete", "collaborator:list",
				"collaborator_product:create", "collaborator_product:read", "collaborator_product:update", "collaborator_product:delete", "collaborator_product:list",
				"sales:create", "sales:read", "sales:update", "sales:list",
				"return:create", "return:read", "return:update", "return:delete", "return:list",
				"invoice:create", "invoice:read", "invoice:update", "invoice:list",
				"invoice:approve",
				"payment:read", "payment:list",
				"bank_payment:read", "bank_payment:create", "bank_payment:list",
				"discount:create", "discount:read", "discount:update", "discount:delete", "discount:list",
				"refund_policy:create", "refund_policy:read", "refund_policy:update", "refund_policy:list",
			},
		},
		{
			Name:        "erp_catalog_manager",
			Description: "Catalog Manager role with product and pricing management",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"product:create", "product:read", "product:update", "product:delete", "product:list",
				"variant:create", "variant:read", "variant:update", "variant:delete", "variant:list",
				"product_price:create", "product_price:read", "product_price:update", "product_price:delete", "product_price:list",
				"tax:create", "tax:read", "tax:update", "tax:delete", "tax:list", "tax:calculate",
				"discount:read", "discount:list",
			},
		},
		{
			Name:        "erp_admin",
			Description: "ERP Administrator role with all ERP permissions",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				"*:*", // All permissions wildcard for ERP resources
			},
		},
		{
			Name:        "erp_auditor",
			Description: "ERP Auditor role with read-only access to all ERP resources",
			Scope:       models.RoleScopeOrg,
			Permissions: []string{
				"*:read",
				"*:list",
				"*:export",
			},
		},
	}
}

// Validate validates the seed data before execution
func (s *ERPSeedProvider) Validate(ctx context.Context) error {
	// First validate base provider
	if err := s.BaseSeedProvider.Validate(ctx); err != nil {
		return err
	}

	// Validate resources
	for _, resource := range s.GetResources() {
		if err := ValidateResource(resource); err != nil {
			return err
		}
	}

	// Validate actions
	for _, action := range s.GetActions() {
		if err := ValidateAction(action); err != nil {
			return err
		}
	}

	// Validate roles
	for _, role := range s.GetRoles() {
		if err := ValidateRole(role); err != nil {
			return err
		}
	}

	return nil
}
