package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"

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

func GenerateSpiceDBSchema(roles []model.Role) []model.CreateSchema {
	// First, let's create a structure to organize the data
	type ResourceInfo struct {
		Relations   map[string]struct{}            // Track all relations needed
		Permissions map[string]map[string]struct{} // action -> roles
	}

	resourceMap := make(map[string]*ResourceInfo)

	// Process all roles and permissions
	for _, role := range roles {
		roleName := strings.ToLower(role.Name)

		for _, perm := range role.Permissions {
			resource := formatSpiceDBResource(perm.Resource)
			if resource == "" {
				continue
			}

			// Initialize resource if not exists
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
				action = strings.ToLower(action)

				if _, exists := resInfo.Permissions[action]; !exists {
					resInfo.Permissions[action] = make(map[string]struct{})
				}
				resInfo.Permissions[action][roleName] = struct{}{}
			}
		}
	}

	// Convert to the output format
	var schemaDefinitions []model.CreateSchema

	// Sort resources for consistent output
	resources := make([]string, 0, len(resourceMap))
	for resource := range resourceMap {
		resources = append(resources, resource)
	}
	sort.Strings(resources)

	for _, resource := range resources {
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
func formatSpiceDBResource(resource string) string {
	// Convert to lowercase and trim spaces
	resource = strings.ToLower(strings.TrimSpace(resource))

	// Remove all non-alphanumeric characters except underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	resource = reg.ReplaceAllString(resource, "")

	// Remove "db" prefix if it exists (we'll add it back properly)
	resource = strings.TrimPrefix(resource, "db")
	resource = strings.TrimPrefix(resource, "_")

	// Add single "db_" prefix
	resource = "db_" + resource

	// Ensure the name follows SpiceDB's pattern:
	// ^[a-z][a-z0-9_]{1,62}[a-z0-9]$

	// 1. Must end with alphanumeric (not underscore)
	if len(resource) > 0 && resource[len(resource)-1] == '_' {
		resource = resource[:len(resource)-1] + "0"
	}

	// 2. Length must be between 3 and 64 characters
	if len(resource) > 64 {
		resource = resource[:64]
		// Ensure it still ends properly
		if resource[len(resource)-1] == '_' {
			resource = resource[:63] + "0"
		}
	}

	// Final validation
	validReg := regexp.MustCompile(`^db_[a-z0-9_]{1,61}[a-z0-9]$`)
	if !validReg.MatchString(resource) {
		log.Printf("Invalid resource format after conversion: %s", resource)
		return ""
	}

	return resource
}

func NormalizeResourceType(resourceType string) string {
	// Replace slashes with underscores
	normalized := strings.ReplaceAll(resourceType, "/", "_")
	// Replace spaces with underscores (if any)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}
