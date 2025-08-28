package responses

import (
	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// ToUserInfo converts a User model to UserInfo response
func ToUserInfo(user *models.User, includeProfile, includeRoles, includeContacts bool) *UserInfo {
	if user == nil {
		return nil
	}

	userInfo := &UserInfo{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		Username:    user.Username,
		IsValidated: user.IsValidated,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Tokens:      user.Tokens,
		HasMPin:     user.HasMPin(),
	}

	// Include roles if requested and available
	if includeRoles && len(user.Roles) > 0 {
		userInfo.Roles = make([]UserRoleDetail, len(user.Roles))
		for i, userRole := range user.Roles {
			userInfo.Roles[i] = ToUserRoleDetail(&userRole)
		}
	}

	// Include profile if requested and available
	if includeProfile && user.Profile.ID != "" {
		userInfo.Profile = ToUserProfileInfo(&user.Profile)
	}

	// Include contacts if requested and available
	if includeContacts && len(user.Contacts) > 0 {
		userInfo.Contacts = make([]ContactInfo, len(user.Contacts))
		for i, contact := range user.Contacts {
			userInfo.Contacts[i] = ToContactInfo(&contact)
		}
	}

	return userInfo
}

// ToUserRoleDetail converts a UserRole model to UserRoleDetail response
func ToUserRoleDetail(userRole *models.UserRole) UserRoleDetail {
	return UserRoleDetail{
		ID:       userRole.ID,
		UserID:   userRole.UserID,
		RoleID:   userRole.RoleID,
		Role:     ToRoleDetail(&userRole.Role),
		IsActive: userRole.IsActive,
	}
}

// ToRoleDetail converts a Role model to RoleDetail response
func ToRoleDetail(role *models.Role) RoleDetail {
	return RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Scope:       string(role.Scope),
		IsActive:    role.IsActive,
		Version:     role.Version,
	}
}

// ToUserProfileInfo converts a UserProfile model to UserProfileInfo response
func ToUserProfileInfo(profile *models.UserProfile) *UserProfileInfo {
	if profile == nil || profile.ID == "" {
		return nil
	}

	profileInfo := &UserProfileInfo{
		ID:            profile.ID,
		Name:          profile.Name,
		CareOf:        profile.CareOf,
		DateOfBirth:   profile.DateOfBirth,
		YearOfBirth:   profile.YearOfBirth,
		AadhaarNumber: profile.AadhaarNumber,
		EmailHash:     profile.EmailHash,
		ShareCode:     profile.ShareCode,
	}

	// Include address if available
	if profile.Address.ID != "" {
		profileInfo.Address = ToAddressInfo(&profile.Address)
	}

	return profileInfo
}

// ToAddressInfo converts an Address model to AddressInfo response
func ToAddressInfo(address *models.Address) *AddressInfo {
	if address == nil || address.ID == "" {
		return nil
	}

	return &AddressInfo{
		ID:          address.ID,
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
	}
}

// ToContactInfo converts a Contact model to ContactInfo response
func ToContactInfo(contact *models.Contact) ContactInfo {
	return ContactInfo{
		ID:          contact.ID,
		Type:        contact.Type,
		Value:       contact.Value,
		IsPrimary:   contact.IsPrimary,
		IsVerified:  contact.IsVerified,
		Description: contact.Description,
	}
}
