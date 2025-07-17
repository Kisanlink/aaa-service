package models

import "github.com/Kisanlink/aaa-service/entities/models"

// Test data for TestNewAddress
var NewAddressTests = []struct {
	name       string
	street     string
	city       string
	state      string
	country    string
	postalCode string
}{
	{
		name:       "Valid address",
		street:     "123 Main St",
		city:       "New York",
		state:      "NY",
		country:    "USA",
		postalCode: "10001",
	},
	{
		name:       "Address with empty fields",
		street:     "",
		city:       "Los Angeles",
		state:      "CA",
		country:    "USA",
		postalCode: "90210",
	},
}

// Test data for TestAddressBeforeCreate
var AddressBeforeCreateTests = []struct {
	name        string
	street      string
	city        string
	state       string
	country     string
	postalCode  string
	shouldError bool
}{
	{
		name:        "Valid address creation",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: false,
	},
	{
		name:        "Address with missing required fields",
		street:      "",
		city:        "",
		state:       "",
		country:     "",
		postalCode:  "",
		shouldError: false, // Address creation doesn't require validation
	},
}

// Test data for TestAddressBeforeUpdate
var AddressBeforeUpdateTests = []struct {
	name        string
	street      string
	city        string
	state       string
	country     string
	postalCode  string
	shouldError bool
}{
	{
		name:        "Valid address update",
		street:      "456 Oak Ave",
		city:        "Los Angeles",
		state:       "CA",
		country:     "USA",
		postalCode:  "90210",
		shouldError: false,
	},
	{
		name:        "Address update with empty fields",
		street:      "",
		city:        "",
		state:       "",
		country:     "",
		postalCode:  "",
		shouldError: false,
	},
}

// Test data for TestAddressValidation
var AddressValidationTests = []struct {
	name        string
	street      string
	city        string
	state       string
	country     string
	postalCode  string
	addressType string
	shouldError bool
}{
	{
		name:        "Valid address validation",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		addressType: "home",
		shouldError: false,
	},
	{
		name:        "Address with invalid postal code",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "invalid",
		addressType: "work",
		shouldError: false, // Address validation is lenient
	},
}

// Helper function to create test address with specific data
func CreateTestAddress(street, city, state, country, postalCode string) *models.Address {
	address := models.NewAddress()
	address.Street = &street
	address.VTC = &city
	address.State = &state
	address.Country = &country
	address.Pincode = &postalCode
	return address
}

// Helper function to create test address with all fields
func CreateTestAddressWithAllFields(street, city, state, country, postalCode, addressType string) *models.Address {
	address := models.NewAddress()
	address.Street = &street
	address.VTC = &city
	address.State = &state
	address.Country = &country
	address.Pincode = &postalCode
	// Note: addressType is not a field in the Address model, skipping it
	return address
}

// Helper function to validate address fields
func ValidateAddress(address *models.Address, expectedStreet, expectedCity, expectedState, expectedCountry, expectedPostalCode string) bool {
	return address.Street != nil && *address.Street == expectedStreet &&
		address.VTC != nil && *address.VTC == expectedCity &&
		address.State != nil && *address.State == expectedState &&
		address.Country != nil && *address.Country == expectedCountry &&
		address.Pincode != nil && *address.Pincode == expectedPostalCode
}

func ValidateAddressWithAllFields(address *models.Address, expectedStreet, expectedCity, expectedState, expectedCountry, expectedPostalCode, expectedType string) bool {
	// Note: addressType is not a field in the Address model, so we only validate the other fields
	return ValidateAddress(address, expectedStreet, expectedCity, expectedState, expectedCountry, expectedPostalCode)
}
