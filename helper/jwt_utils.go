package helper

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// ExtractUserIDFromToken extracts user ID from token without full validation (for middleware)
func ExtractUserIDFromToken(tokenString string) (string, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse without verification for quick extraction
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Try to get user ID from sub or user_id
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}
	if userID, ok := claims["user_id"].(string); ok && userID != "" {
		return userID, nil
	}

	return "", fmt.Errorf("user ID not found in token")
}

// GetTokenType extracts token type from token without full validation
func GetTokenType(tokenString string) (string, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if tokenType, ok := claims["token_type"].(string); ok {
		return tokenType, nil
	}

	return "access", nil // Default to access token
}

// HasPermission checks if the token contains a specific permission
func HasPermission(tokenContext *TokenContext, permission string) bool {
	for _, perm := range tokenContext.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// HasScope checks if the token contains a specific scope
func HasScope(tokenContext *TokenContext, scope string) bool {
	for _, s := range tokenContext.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasOrganizationAccess checks if the token has access to a specific organization
func HasOrganizationAccess(tokenContext *TokenContext, organizationID string) bool {
	if tokenContext.UserContext == nil {
		return false
	}

	// Check if user has roles in the organization
	for _, role := range tokenContext.UserContext.Roles {
		if role.OrganizationID != nil && *role.OrganizationID == organizationID && role.IsActive {
			return true
		}
	}

	// Check if user is member of the organization
	for _, org := range tokenContext.UserContext.Organizations {
		if org.ID == organizationID {
			return true
		}
	}

	return false
}

// HasGroupAccess checks if the token has access to a specific group
func HasGroupAccess(tokenContext *TokenContext, groupID string) bool {
	if tokenContext.UserContext == nil {
		return false
	}

	// Check if user has roles in the group
	for _, role := range tokenContext.UserContext.Roles {
		if role.GroupID != nil && *role.GroupID == groupID && role.IsActive {
			return true
		}
	}

	// Check if user is member of the group
	for _, group := range tokenContext.UserContext.Groups {
		if group.ID == groupID {
			return true
		}
	}

	return false
}

// HasRole checks if the token contains a specific role
func HasRole(tokenContext *TokenContext, roleName string) bool {
	if tokenContext.UserContext == nil {
		return false
	}

	for _, role := range tokenContext.UserContext.Roles {
		if role.Name == roleName && role.IsActive {
			return true
		}
	}
	return false
}

// HasActiveRole checks if the token contains any active role
func HasActiveRole(tokenContext *TokenContext) bool {
	if tokenContext.UserContext == nil {
		return false
	}

	for _, role := range tokenContext.UserContext.Roles {
		if role.IsActive {
			return true
		}
	}
	return false
}

// GetUserOrganizations returns all organization IDs the user has access to
func GetUserOrganizations(tokenContext *TokenContext) []string {
	if tokenContext.UserContext == nil {
		return nil
	}

	orgSet := make(map[string]bool)
	var organizations []string

	// From direct organization membership
	for _, org := range tokenContext.UserContext.Organizations {
		if !orgSet[org.ID] {
			organizations = append(organizations, org.ID)
			orgSet[org.ID] = true
		}
	}

	// From role-based organization access
	for _, role := range tokenContext.UserContext.Roles {
		if role.OrganizationID != nil && role.IsActive && !orgSet[*role.OrganizationID] {
			organizations = append(organizations, *role.OrganizationID)
			orgSet[*role.OrganizationID] = true
		}
	}

	return organizations
}

// GetUserGroups returns all group IDs the user has access to
func GetUserGroups(tokenContext *TokenContext) []string {
	if tokenContext.UserContext == nil {
		return nil
	}

	groupSet := make(map[string]bool)
	var groups []string

	// From direct group membership
	for _, group := range tokenContext.UserContext.Groups {
		if !groupSet[group.ID] {
			groups = append(groups, group.ID)
			groupSet[group.ID] = true
		}
	}

	// From role-based group access
	for _, role := range tokenContext.UserContext.Roles {
		if role.GroupID != nil && role.IsActive && !groupSet[*role.GroupID] {
			groups = append(groups, *role.GroupID)
			groupSet[*role.GroupID] = true
		}
	}

	return groups
}

// IsTokenExpired checks if the token is expired
func IsTokenExpired(tokenContext *TokenContext) bool {
	return tokenContext.ExpiresAt.Before(time.Now())
}

// GetTokenRemainingTime returns the remaining time before token expiration
func GetTokenRemainingTime(tokenContext *TokenContext) time.Duration {
	return time.Until(tokenContext.ExpiresAt)
}
