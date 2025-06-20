package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/Kisanlink/aaa-service/model"
	"golang.org/x/crypto/bcrypt"
)

func PrettyJSON(body interface{}) {

	marshaled, err := json.MarshalIndent(body, "", "   ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", string(marshaled))
}

// Helper function to safely handle nil strings
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func LowerCaseSlice(input []string) []string {
	for i, val := range input {
		input[i] = strings.ToLower(val)
	}
	return input
}

func UpperCaseSlice(input []string) []string {
	for i, val := range input {
		input[i] = strings.ToUpper(val)
	}
	return input
}

func IsValidUsername(username string) bool {
	// Regular expression to match only allowed characters
	// ^ asserts position at start of the string
	// [a-zA-Z0-9/_|\\-=+] matches allowed characters
	// + matches the previous token one or more times
	// $ asserts position at end of the string
	validUsernamePattern := `^[a-zA-Z0-9/_|\-=+]+$`
	match, _ := regexp.MatchString(validUsernamePattern, username)
	return match
}
func IsZeroValued[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func IsRepeatedDigitNumber(num string) bool {
	if len(num) == 0 {
		return false
	}
	firstDigit := num[0]
	for i := 1; i < len(num); i++ {
		if num[i] != firstDigit {
			return false
		}
	}
	return true
}
func IsSequentialNumber(num string) bool {
	if len(num) <= 1 {
		return false
	}

	// Convert each character to its actual digit value
	digits := make([]int, len(num))
	for i, c := range num {
		digits[i] = int(c - '0')
	}

	ascending := true
	descending := true

	for i := 1; i < len(digits); i++ {
		// Check ascending sequence (1,2,3...)
		if digits[i] != (digits[i-1]+1)%10 { // Using %10 to handle 9->0 case
			ascending = false
		}
		// Check descending sequence (9,8,7...)
		if digits[i] != (digits[i-1]-1+10)%10 { // Using +10 to handle 0->9 case
			descending = false
		}
		// Early exit if both fail
		if !ascending && !descending {
			return false
		}
	}

	return ascending || descending
}
func GenerateSpiceDBSchema(roles []model.Role, resources []model.Resource) []model.CreateSchema {
	// First, let's create a structure to organize the data
	type ResourceInfo struct {
		Relations   map[string]struct{}            // Track all relations needed
		Permissions map[string]map[string]struct{} // action -> roles
	}

	resourceMap := make(map[string]*ResourceInfo)

	// First pass: create resource entries for all resources
	for _, res := range resources {
		resource := SanitizeDBName(res.Name)
		if resource == "" || resource == "db_" {
			continue
		}

		if _, exists := resourceMap[resource]; !exists {
			resourceMap[resource] = &ResourceInfo{
				Relations:   make(map[string]struct{}),
				Permissions: make(map[string]map[string]struct{}),
			}
		}
	}

	// Process all roles and permissions
	for _, role := range roles {
		roleName := SanitizeDBName(role.Name)
		if roleName == "" {
			continue
		}

		for _, perm := range role.Permissions {
			resource := SanitizeDBName(perm.Resource)
			if resource == "db_" {
				continue
			}

			// Initialize resource if not exists (shouldn't happen due to first pass)
			if _, exists := resourceMap[resource]; !exists {
				resourceMap[resource] = &ResourceInfo{
					Relations:   make(map[string]struct{}),
					Permissions: make(map[string]map[string]struct{}),
				}
			}
			resInfo := resourceMap[resource]

			// Add relation (role becomes the relation)
			resInfo.Relations[roleName] = struct{}{}

			// Add permissions for each action
			for _, action := range perm.Actions {
				actionName := SanitizeDBName(action)
				if actionName == "" {
					continue
				}

				if _, exists := resInfo.Permissions[actionName]; !exists {
					resInfo.Permissions[actionName] = make(map[string]struct{})
				}
				resInfo.Permissions[actionName][roleName] = struct{}{}
			}
		}
	}

	// Second pass: handle inheritance for resources with more than two words
	for resource, resInfo := range resourceMap {
		parts := strings.Split(resource, "_")
		if len(parts) <= 2 {
			continue // Only process resources with more than two words
		}

		parentResource := strings.Join(parts[:2], "_")
		if parentInfo, exists := resourceMap[parentResource]; exists {
			// Inherit relations from parent
			for relation := range parentInfo.Relations {
				resInfo.Relations[relation] = struct{}{}
			}

			// Inherit permissions from parent
			for action, roles := range parentInfo.Permissions {
				if _, exists := resInfo.Permissions[action]; !exists {
					resInfo.Permissions[action] = make(map[string]struct{})
				}
				for role := range roles {
					resInfo.Permissions[action][role] = struct{}{}
				}
			}
		}
	}

	// Convert to the output format
	var schemaDefinitions []model.CreateSchema

	// Sort resources for consistent output
	resourceNames := make([]string, 0, len(resourceMap))
	for resource := range resourceMap {
		resourceNames = append(resourceNames, resource)
	}
	sort.Strings(resourceNames)

	for _, resource := range resourceNames {
		resInfo := resourceMap[resource]

		// Prepare relations slice
		relations := make([]string, 0, len(resInfo.Relations))
		for relation := range resInfo.Relations {
			relations = append(relations, relation)
		}
		sort.Strings(relations)

		// Prepare permissions
		var permissionData []model.Data

		// Sort actions for consistent output
		actions := make([]string, 0, len(resInfo.Permissions))
		for action := range resInfo.Permissions {
			actions = append(actions, action)
		}
		sort.Strings(actions)

		for _, action := range actions {
			// Get roles for this action
			roles := make([]string, 0, len(resInfo.Permissions[action]))
			for role := range resInfo.Permissions[action] {
				roles = append(roles, role)
			}
			sort.Strings(roles)

			// Create permission entry
			permissionData = append(permissionData, model.Data{
				Action: action,
				Roles:  roles,
			})
		}

		schemaDefinitions = append(schemaDefinitions, model.CreateSchema{
			Resource:  resource,
			Relations: relations,
			Data:      permissionData,
		})
	}

	return schemaDefinitions
}

func NormalizeResourceType(resourceType string) string {
	// Replace slashes with underscores
	normalized := strings.ReplaceAll(resourceType, "/", "_")
	// Replace spaces with underscores (if any)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

// SanitizeDBName converts a string into a valid database name with efficient handling of:
// - Multiple slashes converted to single underscore
// - Multiple spaces converted to single underscore
// - Trimmed trailing/leading special chars
// - Proper handling of all special characters
func SanitizeDBName(input string) string {
	var result strings.Builder
	lastWasSpecial := false // Track if last character was special

	for _, r := range strings.TrimSpace(input) {
		switch {
		case r == '/' || r == '-':
			if !lastWasSpecial {
				result.WriteRune('_')
				lastWasSpecial = true
			}
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			result.WriteRune(r)
			lastWasSpecial = false
		case r == '_':
			// Preserve existing underscores
			result.WriteRune(r)
			lastWasSpecial = false
		case unicode.IsSpace(r):
			if !lastWasSpecial {
				result.WriteRune('_')
				lastWasSpecial = true
			}
		default:
			// Skip all other special characters
			continue
		}
	}

	// Get the sanitized string
	s := result.String()

	// Trim any remaining underscores at ends
	s = strings.Trim(s, "_")

	// Ensure we don't return an empty string
	if s == "" {
		return "invalid_name"
	}

	return s
}

// testCases := []string{
// "db/aaa/users",
// "my@database/name",
// "  spaces  /front/back  ",
// "special!chars#here",
// "UPPER/lower/CASE",
// "multiple///slashes",
// "   leading/trailing///   ",
// "normal-name_123",
// "multiple---dashes",
// "mixed-hyphen_underscore/and space",
// "",
// "just    spaces",
// }

// for _, tc := range testCases {
// 	fmt.Printf("Original: '%s'\nSanitized: '%s'\n\n", tc, SanitizeDBName(tc))
// }

var (
	ErrEmptyName                 = errors.New("name cannot be empty")
	ErrInvalidFormat             = errors.New("name must be lowercase letters separated by single underscores (like 'good_vibes')")
	ErrInvalidChars              = errors.New("name contains invalid characters (only lowercase a-z and _ allowed)")
	ErrCapInvalidChars           = errors.New("name contains invalid characters (only lowercase A-Z and _ allowed)")
	ErrConsecutiveUnderscores    = errors.New("name contains consecutive underscores")
	ErrLeadingTrailingUnderscore = errors.New("name cannot start or end with underscore")
)

func OnlyValidName(input string) error {
	if input == "" {
		return ErrEmptyName
	}

	// Check for invalid characters (only a-z and _ allowed)
	for _, r := range input {
		if !(r >= 'a' && r <= 'z') && r != '_' {
			return ErrInvalidChars
		}
	}

	// Cannot start/end with underscore
	if strings.HasPrefix(input, "_") || strings.HasSuffix(input, "_") {
		return ErrLeadingTrailingUnderscore
	}

	// No consecutive underscores
	if strings.Contains(input, "__") {
		return ErrConsecutiveUnderscores
	}

	// If underscores exist, validate parts between them
	if strings.Contains(input, "_") {
		parts := strings.Split(input, "_")
		for _, part := range parts {
			if part == "" {
				return ErrConsecutiveUnderscores
			}
			// Each part must be lowercase letters only
			for _, r := range part {
				if !(r >= 'a' && r <= 'z') {
					return ErrInvalidChars
				}
			}
		}
	}

	return nil // "read" is now valid
}

// testCases := []string{
// 		"good_vibes", // only valid case
// 		"db/aaa/users",
// 		"my@database/name",
// 		"  spaces  /front/back  ",
// 		"special!chars#here",
// 		"UPPER/lower/CASE",
// 		"multiple///slashes",
// 		"   leading/trailing///   ",
// 		"normal-name_123",
// 		"multiple---dashes",
// 		"mixed-hyphen_underscore/and space",
// 		"",
// 		"just    spaces",
// 		"good__vibes", // consecutive _
// 		"_good_vibes", // starts with _
// 		"good_vibes_", // ends with _
// 		"Good_Vibes",  // uppercase
// 		"goodVibes",   // camelCase
// 		"good-vibes",  // hyphen
// "good"
// 	}

// 	for _, tc := range testCases {
// 		err := OnlyValidName(tc)
// 		if err != nil {
// 			fmt.Printf("Invalid: '%s' - %v\n", tc, err)
// 		} else {
// 			fmt.Printf("Valid:   '%s'\n", tc)
// 		}
// 	}

func OnlyValidCapitalName(input string) error {
	if input == "" {
		return ErrEmptyName
	}

	// Check for invalid characters (only A-Z and _ allowed)
	for _, r := range input {
		if !(r >= 'A' && r <= 'Z') && r != '_' {
			return ErrCapInvalidChars
		}
	}

	// Cannot start/end with underscore
	if strings.HasPrefix(input, "_") || strings.HasSuffix(input, "_") {
		return ErrLeadingTrailingUnderscore
	}

	// No consecutive underscores
	if strings.Contains(input, "__") {
		return ErrConsecutiveUnderscores
	}

	// If underscores exist, validate parts between them
	if strings.Contains(input, "_") {
		parts := strings.Split(input, "_")
		for _, part := range parts {
			if part == "" {
				return ErrConsecutiveUnderscores
			}
			// Each part must be uppercase letters only
			for _, r := range part {
				if !(r >= 'A' && r <= 'Z') {
					return ErrCapInvalidChars
				}
			}
		}
	}

	return nil
}
