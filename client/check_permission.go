package client

import (
	"context"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)


func CheckUserPermissions(userID string, roles []string, permissions []string) (map[string]bool, error) {
	//  Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	// Prepare the result map
	permissionResults := make(map[string]bool)

	//  Check permissions for the user through roles
	for _, permission := range permissions {
		// Assume the permission is tied to the "assign_permission" resource type
		request := &pb.CheckPermissionRequest{
			Resource: &pb.ObjectReference{
				ObjectType: "assign_permission",
				ObjectId:   permission,
			},
			Permission: permission,
			Subject: &pb.SubjectReference{
				Object: &pb.ObjectReference{
					ObjectType: "user",
					ObjectId:   userID,
				},
			},
		}

		response, err := spicedb.CheckPermission(context.Background(), request)
		if err != nil {
			log.Printf("Failed to check permission %s for user %s: %s", permission, userID, err)
			return nil, err
		}

		// Store the result
		permissionResults[permission] = response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
	}

	//  Check permissions for the user through roles
	for _, role := range roles {
		for _, permission := range permissions {
			request := &pb.CheckPermissionRequest{
				Resource: &pb.ObjectReference{
					ObjectType: "assign_permission",
					ObjectId:   permission,
				},
				Permission: permission,
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: "role",
						ObjectId:   role,
					},
				},
			}

			response, err := spicedb.CheckPermission(context.Background(), request)
			if err != nil {
				log.Printf("Failed to check permission %s for role %s: %s", permission, role, err)
				return nil, err
			}

			// If the role has the permission, the user inherits it
			if response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				permissionResults[permission] = true
			}
		}
	}

	log.Printf("Permission check results for user %s: %v", userID, permissionResults)
	return permissionResults, nil
}