package organizations

// RemoveUserFromGroupRequest represents the request for removing a user from a group within an organization
// This is typically handled via URL parameters, but we include this for consistency and validation
type RemoveUserFromGroupRequest struct {
	// No additional fields needed as user_id, group_id, and org_id come from URL path
	// This struct exists for validation and consistency with other request patterns
}
