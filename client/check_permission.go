package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/helper"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func CheckPermission(
	username string, // "alice"
	action string, // "edit"
	resourceType string, // "db/farmers"
	resourceID string, // "123" it is userid
) (bool, error) {
	// Validate parameters
	if username == "" || action == "" || resourceType == "" || resourceID == "" {
		return false, fmt.Errorf("username, action, resourceType and resourceID are required")
	}
	normalizedResourceType := helper.NormalizeResourceType(resourceType)

	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return false, err
	}
	// Check permission
	resp, err := spicedb.CheckPermission(context.Background(), &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: normalizedResourceType,
			ObjectId:   resourceID,
		},
		Permission: action, // Must match schema permissions
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   username,
			},
		},
	})

	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}
