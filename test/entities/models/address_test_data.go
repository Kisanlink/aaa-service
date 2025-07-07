package models

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
		name:       "US Address",
		street:     "123 Main St",
		city:       "New York",
		state:      "NY",
		country:    "USA",
		postalCode: "10001",
	},
	{
		name:       "UK Address",
		street:     "456 Oxford St",
		city:       "London",
		state:      "England",
		country:    "UK",
		postalCode: "W1C 1AP",
	},
	{
		name:       "Indian Address",
		street:     "789 MG Road",
		city:       "Mumbai",
		state:      "Maharashtra",
		country:    "India",
		postalCode: "400001",
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
		name:        "Valid address",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: false,
	},
	{
		name:        "Empty street",
		street:      "",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: true,
	},
	{
		name:        "Empty city",
		street:      "123 Main St",
		city:        "",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: true,
	},
	{
		name:        "Empty country",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "",
		postalCode:  "10001",
		shouldError: true,
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
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: false,
	},
	{
		name:        "Empty street update",
		street:      "",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: true,
	},
	{
		name:        "Empty city update",
		street:      "123 Main St",
		city:        "",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		shouldError: true,
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
		name:        "Valid home address",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		addressType: "home",
		shouldError: false,
	},
	{
		name:        "Valid work address",
		street:      "456 Business Ave",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10002",
		addressType: "work",
		shouldError: false,
	},
	{
		name:        "Invalid address type",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		addressType: "invalid",
		shouldError: true,
	},
	{
		name:        "Empty street",
		street:      "",
		city:        "New York",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		addressType: "home",
		shouldError: true,
	},
	{
		name:        "Empty city",
		street:      "123 Main St",
		city:        "",
		state:       "NY",
		country:     "USA",
		postalCode:  "10001",
		addressType: "home",
		shouldError: true,
	},
	{
		name:        "Empty country",
		street:      "123 Main St",
		city:        "New York",
		state:       "NY",
		country:     "",
		postalCode:  "10001",
		addressType: "home",
		shouldError: true,
	},
}

// Helper function to create test address with specific data
func CreateTestAddress(street, city, state, country, postalCode string) *Address {
	return NewAddress(street, city, state, country, postalCode)
}

// Helper function to create test address with all fields
func CreateTestAddressWithAllFields(street, city, state, country, postalCode, addressType string) *Address {
	address := NewAddress(street, city, state, country, postalCode)
	address.Type = addressType
	return address
}

// Helper function to validate address fields
func ValidateAddressFields(address *Address, expectedStreet, expectedCity, expectedState, expectedCountry, expectedPostalCode string) bool {
	return address.Street == expectedStreet &&
		address.City == expectedCity &&
		address.State == expectedState &&
		address.Country == expectedCountry &&
		address.PostalCode == expectedPostalCode
}

// Helper function to validate address with all fields
func ValidateAddressWithAllFields(address *Address, expectedStreet, expectedCity, expectedState, expectedCountry, expectedPostalCode, expectedType string) bool {
	return address.Street == expectedStreet &&
		address.City == expectedCity &&
		address.State == expectedState &&
		address.Country == expectedCountry &&
		address.PostalCode == expectedPostalCode &&
		address.Type == expectedType
}
