package client

import (
	"context"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)
func CheckUserPermissions(userID string, roles []string, permissions []string, actions []string) (map[string]map[string]bool, error) {
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	result := map[string]map[string]bool{
		"user_actions": {},
		"role_permissions": {},
	}

	ctx := context.Background()

	// Check role-based permissions
	for _, role := range roles {
		for _, permission := range permissions {
			request := &pb.CheckPermissionRequest{
				Resource: &pb.ObjectReference{
					ObjectType: "assign_permission",
					ObjectId:   permission,
				},
				Permission: "allows_action", // Fixed permission check relation
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: "role",
						ObjectId:   role,
					},
				},
			}

			response, err := spicedb.CheckPermission(ctx, request)
			if err != nil {
				log.Printf("Failed to check permission %s for role %s: %s", permission, role, err)
				return nil, err
			}

			result["role_permissions"][permission] = response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
		}
	}

	// Check permission-based actions
	for _, permission := range permissions {
		for _, action := range actions {
			request := &pb.CheckPermissionRequest{
				Resource: &pb.ObjectReference{
					ObjectType: "action",
					ObjectId:   action,
				},
				Permission: action,
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: "assign_permission",
						ObjectId:   permission,
					},
				},
			}

			response, err := spicedb.CheckPermission(ctx, request)
			if err != nil {
				log.Printf("Failed to check action %s for permission %s: %s", action, permission, err)
				return nil, err
			}

			result["user_actions"][action] = response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
		}
	}

	// log.Printf("Permission and action check results for user %s: %v", userID, result)
	return result, nil
}
