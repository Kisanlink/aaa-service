package organizations

// RemoveRoleFromGroupRequest represents the request for removing a role from a group within an organization
// This is typically handled via URL parameters, but we include this for consistency and validation
type RemoveRoleFromGroupRequest struct {
	// No additional fields needed as role_id, group_id, and org_id come from URL path
	// This struct exists for validation and consistency with other request patterns
}
