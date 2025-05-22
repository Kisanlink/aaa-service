package client

import (
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/database"
)

func CreateRelationships(
	roles []string, // array of roles like ["admin", "farmer"]
	username string, // "alice"
	resourceID string, // "123"
) error {
	// Validate parameters
	if len(roles) == 0 || username == "" || resourceID == "" {
		return fmt.Errorf("roles, username, and resourceID are required")
	}

	// Connect to SpiceDB
	client, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return err // Fixed: removed nil
	}
	definitions, err := ReadDefinitionSchema(client)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	// Create a set of roles for O(1) lookup
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	// Check each definition for the roles
	for definition, relations := range definitions {
		// Skip the user definition
		if definition == "user" {
			continue
		}

		// Check if any of the requested roles exist in this definition
		for _, relation := range relations {
			if roleSet[relation] {
				// Create relationship for this definition and role
				err = CreateRelationship(relation, username, definition, resourceID)
				if err != nil {
					return fmt.Errorf("failed to create relationship for %s: %w", definition, err)
				}
			}
		}
	}

	return nil
}
