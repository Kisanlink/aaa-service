package client

import (
	"context"
	"io"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func ReadUserPermissionsActionsAndRoles(userID string) (map[string][]string, error) {
	// Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	// Prepare to collect all relationships
	userData := map[string][]string{
		"roles":      {},
		"permissions": {},
		"actions":     {},
	}

	// Read User -> Roles
	if err := readRelationships(spicedb, "user", userID, userData, "roles"); err != nil {
		return nil, err
	}

	// Read Role -> Permissions
	for _, role := range userData["roles"] {
		if err := readRelationships(spicedb, "role", role, userData, "permissions"); err != nil {
			return nil, err
		}
	}

	// Read Permission -> Actions
	for _, permission := range userData["permissions"] {
		if err := readRelationships(spicedb, "assign_permission", permission, userData, "actions"); err != nil {
			return nil, err
		}
	}

	log.Printf("User %s has Roles: %v, Permissions: %v, Actions: %v", userID, userData["roles"], userData["permissions"], userData["actions"])
	return userData, nil
}

// Helper function to read relationships
func readRelationships(spicedb pb.PermissionsServiceClient, objectType, objectID string, userData map[string][]string, dataType string) error {
	request := &pb.ReadRelationshipsRequest{
		Consistency: &pb.Consistency{
			Requirement: &pb.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
		RelationshipFilter: &pb.RelationshipFilter{
			ResourceType: objectType,
			OptionalResourceId: objectID,
		},
	}

	stream, err := spicedb.ReadRelationships(context.Background(), request)
	if err != nil {
		log.Printf("Failed to read %s relationships for %s %s: %s", dataType, objectType, objectID, err)
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error while reading %s relationships: %s", dataType, err)
			return err
		}
        log.Printf("Result %s",resp)
		userData[dataType] = append(userData[dataType], resp.Relationship.Resource.ObjectId)
	}

	return nil
}
