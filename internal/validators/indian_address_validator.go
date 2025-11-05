package validators

import (
	"fmt"
	"regexp"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// IndianAddressValidator validates Indian address format and fields
type IndianAddressValidator struct {
	pincodeRegex *regexp.Regexp
	validStates  map[string]bool
}

// NewIndianAddressValidator creates a new Indian address validator
func NewIndianAddressValidator() *IndianAddressValidator {
	return &IndianAddressValidator{
		// Indian pincode: 6 digits, first digit cannot be 0
		pincodeRegex: regexp.MustCompile(`^[1-9][0-9]{5}$`),
		validStates:  getValidIndianStates(),
	}
}

// getValidIndianStates returns a map of valid Indian states and union territories
func getValidIndianStates() map[string]bool {
	return map[string]bool{
		// States
		"Andhra Pradesh":    true,
		"Arunachal Pradesh": true,
		"Assam":             true,
		"Bihar":             true,
		"Chhattisgarh":      true,
		"Goa":               true,
		"Gujarat":           true,
		"Haryana":           true,
		"Himachal Pradesh":  true,
		"Jharkhand":         true,
		"Karnataka":         true,
		"Kerala":            true,
		"Madhya Pradesh":    true,
		"Maharashtra":       true,
		"Manipur":           true,
		"Meghalaya":         true,
		"Mizoram":           true,
		"Nagaland":          true,
		"Odisha":            true,
		"Punjab":            true,
		"Rajasthan":         true,
		"Sikkim":            true,
		"Tamil Nadu":        true,
		"Telangana":         true,
		"Tripura":           true,
		"Uttar Pradesh":     true,
		"Uttarakhand":       true,
		"West Bengal":       true,
		// Union Territories
		"Andaman and Nicobar Islands": true,
		"Chandigarh":                  true,
		"Dadra and Nagar Haveli":      true,
		"Daman and Diu":               true,
		"Delhi":                       true,
		"Jammu and Kashmir":           true,
		"Ladakh":                      true,
		"Lakshadweep":                 true,
		"Puducherry":                  true,
	}
}

// ValidateIndianAddress validates an address according to Indian format rules
func (v *IndianAddressValidator) ValidateIndianAddress(addr *models.Address) error {
	if addr == nil {
		return fmt.Errorf("address cannot be nil")
	}

	// Validate required fields for Indian address
	if addr.State == nil || *addr.State == "" {
		return fmt.Errorf("state is required for Indian address")
	}

	// Validate state is a valid Indian state
	if !v.validStates[*addr.State] {
		return fmt.Errorf("invalid Indian state: %s", *addr.State)
	}

	// Validate pincode format if provided
	if addr.Pincode != nil && *addr.Pincode != "" {
		if !v.pincodeRegex.MatchString(*addr.Pincode) {
			return fmt.Errorf("invalid pincode format: must be 6 digits, cannot start with 0 (got: %s)", *addr.Pincode)
		}
	}

	// Validate at least one location field is present
	if (addr.House == nil || *addr.House == "") &&
		(addr.Street == nil || *addr.Street == "") &&
		(addr.VTC == nil || *addr.VTC == "") &&
		(addr.District == nil || *addr.District == "") {
		return fmt.Errorf("at least one location field (house, street, vtc, or district) is required")
	}

	return nil
}

// ValidatePincode validates only the pincode format
func (v *IndianAddressValidator) ValidatePincode(pincode string) error {
	if pincode == "" {
		return fmt.Errorf("pincode cannot be empty")
	}

	if !v.pincodeRegex.MatchString(pincode) {
		return fmt.Errorf("invalid pincode format: must be 6 digits, cannot start with 0 (got: %s)", pincode)
	}

	return nil
}

// ValidateState validates only the state name
func (v *IndianAddressValidator) ValidateState(state string) error {
	if state == "" {
		return fmt.Errorf("state cannot be empty")
	}

	if !v.validStates[state] {
		return fmt.Errorf("invalid Indian state: %s", state)
	}

	return nil
}

// IsValidState checks if the given state is a valid Indian state
func (v *IndianAddressValidator) IsValidState(state string) bool {
	return v.validStates[state]
}

// IsValidPincode checks if the given pincode is valid
func (v *IndianAddressValidator) IsValidPincode(pincode string) bool {
	return v.pincodeRegex.MatchString(pincode)
}

// GetValidStates returns the list of valid Indian states
func (v *IndianAddressValidator) GetValidStates() []string {
	states := make([]string, 0, len(v.validStates))
	for state := range v.validStates {
		states = append(states, state)
	}
	return states
}
