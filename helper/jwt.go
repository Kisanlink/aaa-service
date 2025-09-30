package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/config"
	"github.com/Kisanlink/aaa-service/internal/entities/models"
	jwt "github.com/golang-jwt/jwt/v4"
)

// OrganizationContext represents organization information in JWT
type OrganizationContext struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupContext represents group information in JWT
type GroupContext struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}

// RoleContext represents enhanced role information in JWT
type RoleContext struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Scope          string               `json:"scope"`
	OrganizationID *string              `json:"organization_id,omitempty"`
	GroupID        *string              `json:"group_id,omitempty"`
	IsActive       bool                 `json:"is_active"`
	Organization   *OrganizationContext `json:"organization,omitempty"`
	Group          *GroupContext        `json:"group,omitempty"`
}

// UserContext represents comprehensive user context in JWT
type UserContext struct {
	ID            string                `json:"id"`
	Username      *string               `json:"username,omitempty"`
	PhoneNumber   string                `json:"phone_number"`
	CountryCode   string                `json:"country_code"`
	IsValidated   bool                  `json:"is_validated"`
	Status        *string               `json:"status,omitempty"`
	Roles         []RoleContext         `json:"roles"`
	Organizations []OrganizationContext `json:"organizations"`
	Groups        []GroupContext        `json:"groups"`
}

// GenerateAccessToken generates a JWT access token with comprehensive organizational context
func GenerateAccessToken(userID string, userRoles []models.UserRole, username string, isValidated bool) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	now := time.Now()
	iat := now.Add(-cfg.Leeway / 2)
	nbf := now.Add(-cfg.Leeway / 2)
	exp := now.Add(cfg.TTL)

	// Build enhanced role context with organization and group information
	roleContexts := make([]RoleContext, len(userRoles))
	organizationMap := make(map[string]OrganizationContext)
	groupMap := make(map[string]GroupContext)

	for i, userRole := range userRoles {
		roleContext := RoleContext{
			ID:       userRole.RoleID,
			Name:     userRole.Role.Name,
			Scope:    string(userRole.Role.Scope),
			IsActive: userRole.IsActive && userRole.Role.IsActive,
		}

		// Add organization context if role is organization-scoped
		if userRole.Role.OrganizationID != nil {
			roleContext.OrganizationID = userRole.Role.OrganizationID
			// Note: In a real implementation, you'd fetch organization details
			// For now, we'll include the ID and populate name if available
		}

		// Add group context if role is group-scoped
		if userRole.Role.GroupID != nil {
			roleContext.GroupID = userRole.Role.GroupID
			// Note: In a real implementation, you'd fetch group details
		}

		roleContexts[i] = roleContext
	}

	// Extract unique organizations and groups from roles
	organizations := make([]OrganizationContext, 0, len(organizationMap))
	for _, org := range organizationMap {
		organizations = append(organizations, org)
	}

	groups := make([]GroupContext, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, group)
	}

	// Build comprehensive user context
	userContext := UserContext{
		ID:            userID,
		Username:      &username,
		IsValidated:   isValidated,
		Roles:         roleContexts,
		Organizations: organizations,
		Groups:        groups,
	}

	claims := jwt.MapClaims{
		// Standard JWT claims
		"sub": userID,
		"iss": cfg.Issuer,
		"aud": cfg.Audience,
		"iat": iat.Unix(),
		"nbf": nbf.Unix(),
		"exp": exp.Unix(),
		"jti": generateJTI(), // Unique token identifier

		// Enhanced user context
		"user_context": userContext,

		// Security and session information
		"session_id":    generateSessionID(),
		"token_type":    "access",
		"token_version": "2.0",

		// Legacy fields for backward compatibility
		"user_id":    userID,
		"roleIds":    userRoles, // Keep original format for compatibility
		"username":   username,
		"isvalidate": isValidated,

		// Additional security context
		"permissions":    extractPermissions(userRoles),
		"scopes":         extractScopes(userRoles),
		"tenant_context": extractTenantContext(userRoles),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GenerateRefreshToken generates a long-lived JWT refresh token with minimal context
func GenerateRefreshToken(userID string, userRoles []models.UserRole, username string, isValidated bool) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	now := time.Now()
	iat := now.Add(-cfg.Leeway / 2)
	nbf := now.Add(-cfg.Leeway / 2)
	exp := now.Add(7 * 24 * time.Hour) // 7 days

	// Refresh tokens contain minimal information for security
	claims := jwt.MapClaims{
		// Standard JWT claims
		"sub": userID,
		"iss": cfg.Issuer,
		"aud": cfg.Audience,
		"iat": iat.Unix(),
		"nbf": nbf.Unix(),
		"exp": exp.Unix(),
		"jti": generateJTI(),

		// Minimal context for refresh tokens
		"token_type":    "refresh",
		"token_version": "2.0",
		"session_id":    generateSessionID(),

		// Legacy fields for backward compatibility
		"user_id":    userID,
		"username":   username,
		"isvalidate": isValidated,
		"roleIds":    userRoles, // Minimal role info for compatibility
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken validates the JWT token and returns the user ID (sub preferred)
func ValidateToken(tokenString string) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}
	if uid, ok := claims["user_id"].(string); ok && uid != "" {
		return uid, nil
	}
	return "", err
}

// ValidateTokenWithContext validates the JWT token and returns comprehensive token information
func ValidateTokenWithContext(tokenString string) (*TokenContext, error) {
	cfg := config.LoadJWTConfigFromEnv()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	// Extract user ID
	var userID string
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		userID = sub
	} else if uid, ok := claims["user_id"].(string); ok && uid != "" {
		userID = uid
	} else {
		return nil, err
	}

	// Build token context
	tokenContext := &TokenContext{
		UserID:       userID,
		TokenType:    getStringClaim(claims, "token_type", "access"),
		TokenVersion: getStringClaim(claims, "token_version", "1.0"),
		SessionID:    getStringClaim(claims, "session_id", ""),
		JTI:          getStringClaim(claims, "jti", ""),
		IssuedAt:     time.Unix(int64(claims["iat"].(float64)), 0),
		ExpiresAt:    time.Unix(int64(claims["exp"].(float64)), 0),
	}

	// Extract user context if available (v2.0 tokens)
	if userContextData, ok := claims["user_context"]; ok {
		if userContextMap, ok := userContextData.(map[string]any); ok {
			tokenContext.UserContext = parseUserContext(userContextMap)
		}
	}

	// Extract permissions and scopes
	if permissions, ok := claims["permissions"].([]any); ok {
		tokenContext.Permissions = convertToStringSlice(permissions)
	}

	if scopes, ok := claims["scopes"].([]any); ok {
		tokenContext.Scopes = convertToStringSlice(scopes)
	}

	return tokenContext, nil
}

// TokenContext represents comprehensive token information
type TokenContext struct {
	UserID       string       `json:"user_id"`
	TokenType    string       `json:"token_type"`
	TokenVersion string       `json:"token_version"`
	SessionID    string       `json:"session_id"`
	JTI          string       `json:"jti"`
	IssuedAt     time.Time    `json:"issued_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	UserContext  *UserContext `json:"user_context,omitempty"`
	Permissions  []string     `json:"permissions"`
	Scopes       []string     `json:"scopes"`
}

// Helper functions for JWT generation and parsing

// generateJTI generates a unique JWT ID
func generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("jti_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// extractPermissions extracts permission names from user roles
func extractPermissions(userRoles []models.UserRole) []string {
	permissionSet := make(map[string]bool)
	var permissions []string

	for _, userRole := range userRoles {
		if !userRole.IsActive || !userRole.Role.IsActive {
			continue
		}

		for _, permission := range userRole.Role.Permissions {
			if permission.IsActive && !permissionSet[permission.Name] {
				permissions = append(permissions, permission.Name)
				permissionSet[permission.Name] = true
			}
		}
	}

	return permissions
}

// extractScopes extracts authorization scopes from user roles
func extractScopes(userRoles []models.UserRole) []string {
	scopeSet := make(map[string]bool)
	var scopes []string

	for _, userRole := range userRoles {
		if !userRole.IsActive || !userRole.Role.IsActive {
			continue
		}

		// Add role-based scopes
		roleScope := fmt.Sprintf("role:%s", userRole.Role.Name)
		if !scopeSet[roleScope] {
			scopes = append(scopes, roleScope)
			scopeSet[roleScope] = true
		}

		// Add organization-based scopes
		if userRole.Role.OrganizationID != nil {
			orgScope := fmt.Sprintf("org:%s", *userRole.Role.OrganizationID)
			if !scopeSet[orgScope] {
				scopes = append(scopes, orgScope)
				scopeSet[orgScope] = true
			}
		}

		// Add group-based scopes
		if userRole.Role.GroupID != nil {
			groupScope := fmt.Sprintf("group:%s", *userRole.Role.GroupID)
			if !scopeSet[groupScope] {
				scopes = append(scopes, groupScope)
				scopeSet[groupScope] = true
			}
		}

		// Add scope-based scopes (global vs org)
		scopeType := fmt.Sprintf("scope:%s", string(userRole.Role.Scope))
		if !scopeSet[scopeType] {
			scopes = append(scopes, scopeType)
			scopeSet[scopeType] = true
		}
	}

	return scopes
}

// extractTenantContext extracts tenant/organization context from user roles
func extractTenantContext(userRoles []models.UserRole) map[string]any {
	tenantContext := make(map[string]any)
	organizationIDs := make(map[string]bool)
	groupIDs := make(map[string]bool)

	for _, userRole := range userRoles {
		if !userRole.IsActive || !userRole.Role.IsActive {
			continue
		}

		if userRole.Role.OrganizationID != nil {
			organizationIDs[*userRole.Role.OrganizationID] = true
		}

		if userRole.Role.GroupID != nil {
			groupIDs[*userRole.Role.GroupID] = true
		}
	}

	// Convert maps to slices
	var orgs []string
	for orgID := range organizationIDs {
		orgs = append(orgs, orgID)
	}

	var groups []string
	for groupID := range groupIDs {
		groups = append(groups, groupID)
	}

	if len(orgs) > 0 {
		tenantContext["organizations"] = orgs
	}

	if len(groups) > 0 {
		tenantContext["groups"] = groups
	}

	return tenantContext
}

// getStringClaim safely extracts a string claim with a default value
func getStringClaim(claims jwt.MapClaims, key, defaultValue string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return defaultValue
}

// parseUserContext parses user context from JWT claims
func parseUserContext(userContextMap map[string]any) *UserContext {
	userContext := &UserContext{}

	if id, ok := userContextMap["id"].(string); ok {
		userContext.ID = id
	}

	if username, ok := userContextMap["username"].(string); ok {
		userContext.Username = &username
	}

	if phoneNumber, ok := userContextMap["phone_number"].(string); ok {
		userContext.PhoneNumber = phoneNumber
	}

	if countryCode, ok := userContextMap["country_code"].(string); ok {
		userContext.CountryCode = countryCode
	}

	if isValidated, ok := userContextMap["is_validated"].(bool); ok {
		userContext.IsValidated = isValidated
	}

	if status, ok := userContextMap["status"].(string); ok {
		userContext.Status = &status
	}

	// Parse roles
	if rolesData, ok := userContextMap["roles"].([]any); ok {
		userContext.Roles = parseRoleContexts(rolesData)
	}

	// Parse organizations
	if orgsData, ok := userContextMap["organizations"].([]any); ok {
		userContext.Organizations = parseOrganizationContexts(orgsData)
	}

	// Parse groups
	if groupsData, ok := userContextMap["groups"].([]any); ok {
		userContext.Groups = parseGroupContexts(groupsData)
	}

	return userContext
}

// parseRoleContexts parses role contexts from JWT claims
func parseRoleContexts(rolesData []any) []RoleContext {
	var roles []RoleContext

	for _, roleData := range rolesData {
		if roleMap, ok := roleData.(map[string]any); ok {
			role := RoleContext{}

			if id, ok := roleMap["id"].(string); ok {
				role.ID = id
			}

			if name, ok := roleMap["name"].(string); ok {
				role.Name = name
			}

			if scope, ok := roleMap["scope"].(string); ok {
				role.Scope = scope
			}

			if orgID, ok := roleMap["organization_id"].(string); ok {
				role.OrganizationID = &orgID
			}

			if groupID, ok := roleMap["group_id"].(string); ok {
				role.GroupID = &groupID
			}

			if isActive, ok := roleMap["is_active"].(bool); ok {
				role.IsActive = isActive
			}

			roles = append(roles, role)
		}
	}

	return roles
}

// parseOrganizationContexts parses organization contexts from JWT claims
func parseOrganizationContexts(orgsData []any) []OrganizationContext {
	var organizations []OrganizationContext

	for _, orgData := range orgsData {
		if orgMap, ok := orgData.(map[string]any); ok {
			org := OrganizationContext{}

			if id, ok := orgMap["id"].(string); ok {
				org.ID = id
			}

			if name, ok := orgMap["name"].(string); ok {
				org.Name = name
			}

			organizations = append(organizations, org)
		}
	}

	return organizations
}

// parseGroupContexts parses group contexts from JWT claims
func parseGroupContexts(groupsData []any) []GroupContext {
	var groups []GroupContext

	for _, groupData := range groupsData {
		if groupMap, ok := groupData.(map[string]any); ok {
			group := GroupContext{}

			if id, ok := groupMap["id"].(string); ok {
				group.ID = id
			}

			if name, ok := groupMap["name"].(string); ok {
				group.Name = name
			}

			if orgID, ok := groupMap["organization_id"].(string); ok {
				group.OrganizationID = orgID
			}

			groups = append(groups, group)
		}
	}

	return groups
}

// convertToStringSlice converts []any to []string
func convertToStringSlice(data []any) []string {
	var result []string
	for _, item := range data {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}
