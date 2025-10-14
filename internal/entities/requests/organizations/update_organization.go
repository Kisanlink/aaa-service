// Package organizations provides request structures for organization-related operations.
package organizations

// UpdateOrganizationRequest represents the request for updating an existing organization.
// @Description Request body for updating an organization (all fields optional)
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Corp Name"`                                                                                                                                               // Organization name
	Type        *string `json:"type,omitempty" validate:"omitempty,oneof=enterprise small_business individual fpo cooperative agribusiness farmers_group shg ngo government input_supplier trader processing_unit research_institute" example:"cooperative"` // Organization type
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Updated description"`                                                                                                                                           // Organization description
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,org_id" example:"ORGN00000002"`                                                                                                                                                      // Parent organization ID
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`                                                                                                                                                                                          // Whether organization is active
}
