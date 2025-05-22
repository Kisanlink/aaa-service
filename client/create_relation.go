package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/helper"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func CreateRelationship(
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
		return err // Fixed: removed nil
	}

	// Create the relationship
	_, err = spicedb.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
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
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	log.Printf(
		"Created relationship: %s#%s@user:%s",
		fmt.Sprintf("%s:%s", resourceType, resourceID),
		role,
		username,
	)

	return nil
}
