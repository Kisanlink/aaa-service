package helper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func PrettyJSON(body interface{}) {

	marshaled, err := json.MarshalIndent(body, "", "   ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", string(marshaled))
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
// // IsZeroValued checks if a value is the zero value for its type
// func IsZeroValued(i interface{}) bool {
// 	if i == nil {
// 		return true
// 	}

// 	v := reflect.ValueOf(i)
	
// 	// Handle pointer types
// 	if v.Kind() == reflect.Ptr {
// 		if v.IsNil() {
// 			return true
// 		}
// 		v = v.Elem() // Dereference the pointer
// 	}

// 	switch v.Kind() {
// 	case reflect.String:
// 		return v.String() == ""
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return v.Int() == 0
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return v.Uint() == 0
// 	case reflect.Float32, reflect.Float64:
// 		return v.Float() == 0
// 	case reflect.Bool:
// 		return !v.Bool()
// 	case reflect.Struct:
// 		// Special case for time.Time
// 		if t, ok := i.(time.Time); ok {
// 			return t.IsZero()
// 		}
// 		// For other structs, check all fields
// 		return isStructZero(v)
// 	case reflect.Array, reflect.Slice, reflect.Map:
// 		return v.Len() == 0
// 	default:
// 		return false
// 	}
// }

// func isStructZero(v reflect.Value) bool {
// 	for i := 0; i < v.NumField(); i++ {
// 		field := v.Field(i)
// 		if !field.IsZero() {
// 			return false
// 		}
// 	}
// 	return true
// }


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