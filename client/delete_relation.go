package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/helper"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func DeleteRelationship(
	role string, // "viewer", "owner", "farmer"
	username string, // "alice"
	resourceType string, // "db/farmers"
	resourceID string, // "123"
) error {
	// Validate parameters
	if role == "" || username == "" || resourceType == "" || resourceID == "" {
		return fmt.Errorf("role, username, resourceType and resourceID are required")
	}
	normalizedResourceType := helper.NormalizeResourceType(resourceType)

	// Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return err
	}

	// Delete the relationship
	_, err = spicedb.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_DELETE,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: normalizedResourceType,
						ObjectId:   resourceID,
					},
					Relation: role,
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: "user",
							ObjectId:   username,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	log.Printf(
		"Deleted relationship: %s#%s@user:%s",
		fmt.Sprintf("%s:%s", resourceType, resourceID),
		role,
		username,
	)

	return nil
}

func DeleteRelationships(
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
		return err
	}

	// Read the schema to get all definitions and their relations
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
				// Delete relationship for this definition and role
				err = DeleteRelationship(relation, username, definition, resourceID)
				if err != nil {
					return fmt.Errorf("failed to delete relationship for %s: %w", definition, err)
				}
			}
		}
	}

	return nil
}
