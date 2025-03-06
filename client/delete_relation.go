package client

import (
	"context"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)



func DeleteUserRoleRelationship(userID string, roles []string, permissions []string) (*pb.WriteRelationshipsResponse, error) {
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	var updates []*pb.RelationshipUpdate

	for _, role := range roles {
		relationship := &pb.Relationship{
			Resource: &pb.ObjectReference{
				ObjectType: "role",
				ObjectId:   role,
			},
			Relation: role, 
			Subject: &pb.SubjectReference{
				Object: &pb.ObjectReference{
					ObjectType: "user",
					ObjectId:   userID,
				},
			},
		}

		updates = append(updates, &pb.RelationshipUpdate{
			Operation:    pb.RelationshipUpdate_OPERATION_DELETE, 
			Relationship: relationship,
		})
	}

	for _, permission := range permissions {
		for _, role := range roles {
			relationship := &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: "assign_permission",
					ObjectId:   permission,
				},
				Relation: permission, 
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: "role",
						ObjectId:   role,
					},
				},
			}

			updates = append(updates, &pb.RelationshipUpdate{
				Operation:    pb.RelationshipUpdate_OPERATION_DELETE, 
				Relationship: relationship,
			})
		}
	}

	request := &pb.WriteRelationshipsRequest{
		Updates: updates,
	}
	res, err := spicedb.WriteRelationships(context.Background(), request)
	if err != nil {
		log.Printf("Failed to delete user-role and role-permission relationships: %s", err)
		return nil, err
	}

	// log.Printf("User-role and role-permission relationships deleted successfully: %s", res)
	return res, nil
}

