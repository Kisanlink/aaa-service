package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
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

func ConvertAndDeduplicateRolePermissions(input map[string][]model.Permission) []model.RoleRes {
	var roles []model.RoleRes

	for roleName, permissions := range input {
		uniquePerms := make(map[string]model.Permission)

		// Deduplicate permissions
		for _, perm := range permissions {
			key := perm.Name + "|" + perm.Action + "|" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = model.Permission{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		// Convert map to slice
		var permSlice []model.Permission
		for _, perm := range uniquePerms {
			permSlice = append(permSlice, perm)
		}

		roles = append(roles, model.RoleRes{
			RoleName:    roleName,
			Permissions: permSlice,
		})
	}

	return roles
}
