package services

import (
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	aaaresponses "github.com/Kisanlink/aaa-service/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
)

// ResponseTransformerImpl implements the ResponseTransformer interface
type ResponseTransformerImpl struct{}

// NewResponseTransformer creates a new ResponseTransformer instance
func NewResponseTransformer() interfaces.ResponseTransformer {
	return &ResponseTransformerImpl{}
}

// TransformUser transforms a User model into a StandardUserResponse
func (rt *ResponseTransformerImpl) TransformUser(user *models.User, options interfaces.TransformOptions) interface{} {
	if user == nil {
		return nil
	}

	response := &aaaresponses.StandardUserResponse{
		ID:          user.GetID(),
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		Username:    user.Username,
		IsValidated: user.IsValidated,
		IsActive:    rt.determineUserActiveStatus(user),
		Status:      user.Status,
		Tokens:      user.Tokens,
		HasMPin:     user.HasMPin(),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		DeletedAt:   user.DeletedAt,
	}

	// Include profile if requested
	if options.IncludeProfile && user.Profile.GetID() != "" {
		response.Profile = rt.transformUserProfile(&user.Profile, options)
	}

	// Include contacts if requested
	if options.IncludeContacts && len(user.Contacts) > 0 {
		if contacts := rt.transformContacts(user.Contacts, options); contacts != nil {
			response.Contacts = contacts
		}
	}

	// Include roles if requested
	if options.IncludeRole && len(user.Roles) > 0 {
		if roles := rt.TransformUserRoles(user.Roles, options); roles != nil {
			roleSlice := make([]*aaaresponses.StandardUserRoleResponse, 0, len(roles))
			for _, role := range roles {
				if roleResp, ok := role.(*aaaresponses.StandardUserRoleResponse); ok {
					roleSlice = append(roleSlice, roleResp)
				}
			}
			response.Roles = roleSlice
		}
	}

	return response
}

// TransformUsers transforms a slice of User models into StandardUserResponse slice
func (rt *ResponseTransformerImpl) TransformUsers(users []models.User, options interfaces.TransformOptions) []interface{} {
	if len(users) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(users))
	for _, user := range users {
		if options.ExcludeDeleted && user.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !rt.determineUserActiveStatus(&user) {
			continue
		}

		response := rt.TransformUser(&user, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformRole transforms a Role model into a StandardRoleResponse
func (rt *ResponseTransformerImpl) TransformRole(role *models.Role, options interfaces.TransformOptions) interface{} {
	if role == nil {
		return nil
	}

	response := &aaaresponses.StandardRoleResponse{
		ID:             role.GetID(),
		Name:           role.Name,
		Description:    role.Description,
		Scope:          string(role.Scope),
		IsActive:       role.IsActive,
		Version:        role.Version,
		OrganizationID: role.OrganizationID,
		GroupID:        role.GroupID,
		ParentID:       role.ParentID,
		CreatedAt:      role.CreatedAt,
		UpdatedAt:      role.UpdatedAt,
		DeletedAt:      role.DeletedAt,
	}

	// Include permissions if available
	if len(role.Permissions) > 0 {
		if permissions := rt.transformPermissions(role.Permissions, options); permissions != nil {
			response.Permissions = permissions
		}
	}

	// Include children roles if available
	if len(role.Children) > 0 {
		if children := rt.TransformRoles(role.Children, options); children != nil {
			childSlice := make([]*aaaresponses.StandardRoleResponse, 0, len(children))
			for _, child := range children {
				if childResp, ok := child.(*aaaresponses.StandardRoleResponse); ok {
					childSlice = append(childSlice, childResp)
				}
			}
			response.Children = childSlice
		}
	}

	return response
}

// TransformRoles transforms a slice of Role models into StandardRoleResponse slice
func (rt *ResponseTransformerImpl) TransformRoles(roles []models.Role, options interfaces.TransformOptions) []interface{} {
	if len(roles) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(roles))
	for _, role := range roles {
		if options.ExcludeDeleted && role.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !role.IsActive {
			continue
		}

		response := rt.TransformRole(&role, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformUserRole transforms a UserRole model into a StandardUserRoleResponse
func (rt *ResponseTransformerImpl) TransformUserRole(userRole *models.UserRole, options interfaces.TransformOptions) interface{} {
	if userRole == nil {
		return nil
	}

	response := &aaaresponses.StandardUserRoleResponse{
		ID:         userRole.GetID(),
		UserID:     userRole.UserID,
		RoleID:     userRole.RoleID,
		IsActive:   userRole.IsActive,
		AssignedAt: userRole.CreatedAt, // Using CreatedAt as AssignedAt
		CreatedAt:  userRole.CreatedAt,
		UpdatedAt:  userRole.UpdatedAt,
		DeletedAt:  userRole.DeletedAt,
	}

	// Include user if requested
	if options.IncludeUser && userRole.User.GetID() != "" {
		// Create options without IncludeRole to avoid circular references
		userOptions := options
		userOptions.IncludeRole = false
		if user := rt.TransformUser(&userRole.User, userOptions); user != nil {
			if userResp, ok := user.(*aaaresponses.StandardUserResponse); ok {
				response.User = userResp
			}
		}
	}

	// Include role if requested
	if options.IncludeRole && userRole.Role.GetID() != "" {
		if role := rt.TransformRole(&userRole.Role, options); role != nil {
			if roleResp, ok := role.(*aaaresponses.StandardRoleResponse); ok {
				response.Role = roleResp
			}
		}
	}

	return response
}

// TransformUserRoles transforms a slice of UserRole models into StandardUserRoleResponse slice
func (rt *ResponseTransformerImpl) TransformUserRoles(userRoles []models.UserRole, options interfaces.TransformOptions) []interface{} {
	if len(userRoles) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(userRoles))
	for _, userRole := range userRoles {
		if options.ExcludeDeleted && userRole.DeletedAt != nil {
			continue
		}
		if options.OnlyActiveRoles && !userRole.IsActive {
			continue
		}

		response := rt.TransformUserRole(&userRole, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// Helper methods for transforming nested objects

// transformUserProfile transforms a UserProfile model into a StandardUserProfileResponse
func (rt *ResponseTransformerImpl) transformUserProfile(profile *models.UserProfile, options interfaces.TransformOptions) *aaaresponses.StandardUserProfileResponse {
	if profile == nil || profile.GetID() == "" {
		return nil
	}

	response := &aaaresponses.StandardUserProfileResponse{
		ID:          profile.GetID(),
		UserID:      profile.UserID,
		Name:        profile.Name,
		CareOf:      profile.CareOf,
		DateOfBirth: profile.DateOfBirth,
		YearOfBirth: profile.YearOfBirth,
		Photo:       profile.Photo,
		EmailHash:   profile.EmailHash,
		ShareCode:   profile.ShareCode,
		Message:     profile.Message,
		AddressID:   profile.AddressID,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
		DeletedAt:   profile.DeletedAt,
	}

	// Mask Aadhaar number for security
	if profile.AadhaarNumber != nil && *profile.AadhaarNumber != "" {
		masked := rt.maskAadhaarNumber(*profile.AadhaarNumber)
		response.AadhaarNumber = &masked
	}

	return response
}

// transformContacts transforms a slice of Contact models into StandardContactResponse slice
func (rt *ResponseTransformerImpl) transformContacts(contacts []models.Contact, options interfaces.TransformOptions) []*aaaresponses.StandardContactResponse {
	if len(contacts) == 0 {
		return []*aaaresponses.StandardContactResponse{}
	}

	result := make([]*aaaresponses.StandardContactResponse, 0, len(contacts))
	for _, contact := range contacts {
		if options.ExcludeDeleted && contact.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !contact.IsActive {
			continue
		}

		response := &aaaresponses.StandardContactResponse{
			ID:          contact.GetID(),
			UserID:      contact.UserID,
			Type:        contact.Type,
			Value:       contact.Value,
			CountryCode: contact.CountryCode,
			IsPrimary:   contact.IsPrimary,
			IsVerified:  contact.IsVerified,
			IsActive:    contact.IsActive,
			VerifiedAt:  rt.parseVerifiedAt(contact.VerifiedAt),
			CreatedAt:   contact.CreatedAt,
			UpdatedAt:   contact.UpdatedAt,
			DeletedAt:   contact.DeletedAt,
		}

		result = append(result, response)
	}

	return result
}

// transformPermissions transforms a slice of Permission models into StandardPermissionResponse slice
func (rt *ResponseTransformerImpl) transformPermissions(permissions []models.Permission, options interfaces.TransformOptions) []*aaaresponses.StandardPermissionResponse {
	if len(permissions) == 0 {
		return []*aaaresponses.StandardPermissionResponse{}
	}

	result := make([]*aaaresponses.StandardPermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		if options.ExcludeDeleted && permission.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !permission.IsActive {
			continue
		}

		response := &aaaresponses.StandardPermissionResponse{
			ID:          permission.GetID(),
			Name:        permission.Name,
			Description: permission.Description,
			ResourceID:  permission.ResourceID,
			ActionID:    permission.ActionID,
			IsActive:    permission.IsActive,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
			DeletedAt:   permission.DeletedAt,
		}

		result = append(result, response)
	}

	return result
}

// Helper methods for business logic

// determineUserActiveStatus determines if a user is active based on status and deleted_at
func (rt *ResponseTransformerImpl) determineUserActiveStatus(user *models.User) bool {
	// User is not active if soft deleted
	if user.DeletedAt != nil {
		return false
	}

	// User is active if status is "active" or if status is nil/empty (backward compatibility)
	if user.Status == nil {
		return true // Default to active for backward compatibility
	}

	return *user.Status == "active"
}

// maskAadhaarNumber masks an Aadhaar number for security (shows only last 4 digits)
func (rt *ResponseTransformerImpl) maskAadhaarNumber(aadhaar string) string {
	if len(aadhaar) < 4 {
		return "****"
	}

	// Remove any spaces or special characters
	cleaned := strings.ReplaceAll(aadhaar, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if len(cleaned) >= 12 {
		return "XXXX-XXXX-" + cleaned[len(cleaned)-4:]
	}

	return "****" + cleaned[len(cleaned)-4:]
}

// ensureConsistentState ensures is_active and deleted_at fields are logically consistent
func (rt *ResponseTransformerImpl) ensureConsistentState(isActive bool, deletedAt *time.Time) (bool, *time.Time) {
	// If deleted_at is set, is_active should be false
	if deletedAt != nil {
		return false, deletedAt
	}

	// If not deleted, return the provided is_active status
	return isActive, nil
}

// parseVerifiedAt parses a string timestamp into *time.Time for VerifiedAt field
func (rt *ResponseTransformerImpl) parseVerifiedAt(verifiedAt *string) *time.Time {
	if verifiedAt == nil || *verifiedAt == "" {
		return nil
	}

	// Try to parse RFC3339 format first
	if parsed, err := time.Parse(time.RFC3339, *verifiedAt); err == nil {
		return &parsed
	}

	// Try other common formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, *verifiedAt); err == nil {
			return &parsed
		}
	}

	// If parsing fails, return nil
	return nil
}

// TransformOrganization transforms an Organization model into a StandardOrganizationResponse
func (rt *ResponseTransformerImpl) TransformOrganization(org *models.Organization, options interfaces.TransformOptions) interface{} {
	if org == nil {
		return nil
	}

	response := &aaaresponses.StandardOrganizationResponse{
		ID:          org.GetID(),
		Name:        org.Name,
		Description: &org.Description,
		Type:        nil, // Organization model doesn't have Type field
		IsActive:    org.IsActive,
		ParentID:    org.ParentID,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		DeletedAt:   org.DeletedAt,
	}

	// Include children if available
	if len(org.Children) > 0 {
		if children := rt.TransformOrganizations(org.Children, options); children != nil {
			childSlice := make([]*aaaresponses.StandardOrganizationResponse, 0, len(children))
			for _, child := range children {
				if childResp, ok := child.(*aaaresponses.StandardOrganizationResponse); ok {
					childSlice = append(childSlice, childResp)
				}
			}
			response.Children = childSlice
		}
	}

	return response
}

// TransformOrganizations transforms a slice of Organization models into StandardOrganizationResponse slice
func (rt *ResponseTransformerImpl) TransformOrganizations(orgs []models.Organization, options interfaces.TransformOptions) []interface{} {
	if len(orgs) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(orgs))
	for _, org := range orgs {
		if options.ExcludeDeleted && org.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !org.IsActive {
			continue
		}

		response := rt.TransformOrganization(&org, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformGroup transforms a Group model into a StandardGroupResponse
func (rt *ResponseTransformerImpl) TransformGroup(group *models.Group, options interfaces.TransformOptions) interface{} {
	if group == nil {
		return nil
	}

	response := &aaaresponses.StandardGroupResponse{
		ID:             group.GetID(),
		Name:           group.Name,
		Description:    &group.Description,
		Type:           nil, // Group model doesn't have Type field
		IsActive:       group.IsActive,
		OrganizationID: &group.OrganizationID,
		ParentID:       group.ParentID,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
		DeletedAt:      group.DeletedAt,
	}

	// Include organization if requested and available
	if options.IncludeUser && group.Organization != nil && group.Organization.GetID() != "" { // Reusing IncludeUser for organization
		if org := rt.TransformOrganization(group.Organization, options); org != nil {
			if orgResp, ok := org.(*aaaresponses.StandardOrganizationResponse); ok {
				response.Organization = orgResp
			}
		}
	}

	// Include children if available
	if len(group.Children) > 0 {
		if children := rt.TransformGroups(group.Children, options); children != nil {
			childSlice := make([]*aaaresponses.StandardGroupResponse, 0, len(children))
			for _, child := range children {
				if childResp, ok := child.(*aaaresponses.StandardGroupResponse); ok {
					childSlice = append(childSlice, childResp)
				}
			}
			response.Children = childSlice
		}
	}

	return response
}

// TransformGroups transforms a slice of Group models into StandardGroupResponse slice
func (rt *ResponseTransformerImpl) TransformGroups(groups []models.Group, options interfaces.TransformOptions) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(groups))
	for _, group := range groups {
		if options.ExcludeDeleted && group.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !group.IsActive {
			continue
		}

		response := rt.TransformGroup(&group, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformContact transforms a Contact model into a StandardContactResponse
func (rt *ResponseTransformerImpl) TransformContact(contact *models.Contact, options interfaces.TransformOptions) interface{} {
	if contact == nil {
		return nil
	}

	response := &aaaresponses.StandardContactResponse{
		ID:          contact.GetID(),
		UserID:      contact.UserID,
		Type:        contact.Type,
		Value:       contact.Value,
		CountryCode: contact.CountryCode,
		IsPrimary:   contact.IsPrimary,
		IsVerified:  contact.IsVerified,
		IsActive:    contact.IsActive,
		VerifiedAt:  rt.parseVerifiedAt(contact.VerifiedAt),
		VerifiedBy:  contact.VerifiedBy,
		Description: contact.Description,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
		DeletedAt:   contact.DeletedAt,
	}

	return response
}

// TransformContacts transforms a slice of Contact models into StandardContactResponse slice
func (rt *ResponseTransformerImpl) TransformContacts(contacts []models.Contact, options interfaces.TransformOptions) []interface{} {
	if len(contacts) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(contacts))
	for _, contact := range contacts {
		if options.ExcludeDeleted && contact.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !contact.IsActive {
			continue
		}

		response := rt.TransformContact(&contact, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformAddress transforms an Address model into a StandardAddressResponse
func (rt *ResponseTransformerImpl) TransformAddress(address *models.Address, options interfaces.TransformOptions) interface{} {
	if address == nil {
		return nil
	}

	response := &aaaresponses.StandardAddressResponse{
		ID:          address.GetID(),
		House:       address.House,
		Street:      address.Street,
		Landmark:    address.Landmark,
		PostOffice:  address.PostOffice,
		Subdistrict: address.Subdistrict,
		District:    address.District,
		VTC:         address.VTC,
		State:       address.State,
		Country:     address.Country,
		Pincode:     address.Pincode,
		FullAddress: address.FullAddress,
		CreatedAt:   address.CreatedAt,
		UpdatedAt:   address.UpdatedAt,
		DeletedAt:   address.DeletedAt,
	}

	return response
}

// TransformAddresses transforms a slice of Address models into StandardAddressResponse slice
func (rt *ResponseTransformerImpl) TransformAddresses(addresses []models.Address, options interfaces.TransformOptions) []interface{} {
	if len(addresses) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(addresses))
	for _, address := range addresses {
		if options.ExcludeDeleted && address.DeletedAt != nil {
			continue
		}

		response := rt.TransformAddress(&address, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}

// TransformPermission transforms a Permission model into a StandardPermissionResponse
func (rt *ResponseTransformerImpl) TransformPermission(permission *models.Permission, options interfaces.TransformOptions) interface{} {
	if permission == nil {
		return nil
	}

	response := &aaaresponses.StandardPermissionResponse{
		ID:          permission.GetID(),
		Name:        permission.Name,
		Description: permission.Description,
		ResourceID:  permission.ResourceID,
		ActionID:    permission.ActionID,
		IsActive:    permission.IsActive,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
		DeletedAt:   permission.DeletedAt,
	}

	return response
}

// TransformPermissions transforms a slice of Permission models into StandardPermissionResponse slice
func (rt *ResponseTransformerImpl) TransformPermissions(permissions []models.Permission, options interfaces.TransformOptions) []interface{} {
	if len(permissions) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(permissions))
	for _, permission := range permissions {
		if options.ExcludeDeleted && permission.DeletedAt != nil {
			continue
		}
		if options.ExcludeInactive && !permission.IsActive {
			continue
		}

		response := rt.TransformPermission(&permission, options)
		if response != nil {
			result = append(result, response)
		}
	}

	return result
}
