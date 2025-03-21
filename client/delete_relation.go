package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

// func DeleteUserRoleRelationship(userID string, roles []string, permissions []string, actions []string) (*pb.WriteRelationshipsResponse, error) {
// 	// Connect to SpiceDB
// 	spicedb, err := database.SpiceDB()
// 	if err != nil {
// 		log.Printf("Unable to connect to SpiceDB: %s", err)
// 		return nil, err
// 	}

// 	// Prepare relationship deletions
// 	var updates []*pb.RelationshipUpdate

// 	// Delete relationships for user-roles
// 	for _, role := range roles {
// 		relationship := &pb.Relationship{
// 			Resource: &pb.ObjectReference{
// 				ObjectType: "role",
// 				ObjectId:   role,
// 			},
// 			Relation: role, // This should match your role name
// 			Subject: &pb.SubjectReference{
// 				Object: &pb.ObjectReference{
// 					ObjectType: "user",
// 					ObjectId:   userID,
// 				},
// 			},
// 		}

// 		updates = append(updates, &pb.RelationshipUpdate{
// 			Operation:    pb.RelationshipUpdate_OPERATION_DELETE,
// 			Relationship: relationship,
// 		})
// 	}

// 	// Delete relationships for role-permissions
// 	for _, permission := range permissions {
// 		for _, role := range roles {
// 			relationship := &pb.Relationship{
// 				Resource: &pb.ObjectReference{
// 					ObjectType: "assign_permission",
// 					ObjectId:   permission,
// 				},
// 				Relation: "allows_action", // Matches your schema definition
// 				Subject: &pb.SubjectReference{
// 					Object: &pb.ObjectReference{
// 						ObjectType: "role",
// 						ObjectId:   role,
// 					},
// 				},
// 			}

// 			updates = append(updates, &pb.RelationshipUpdate{
// 				Operation:    pb.RelationshipUpdate_OPERATION_DELETE,
// 				Relationship: relationship,
// 			})
// 		}
// 	}

// 	// Delete relationships for permission-actions
// 	for _, action := range actions {
// 		for _, permission := range permissions {
// 			relationship := &pb.Relationship{
// 				Resource: &pb.ObjectReference{
// 					ObjectType: "action",
// 					ObjectId:   action,
// 				},
// 				Relation: action, // Using the action value as the relation (Dynamic)
// 				Subject: &pb.SubjectReference{
// 					Object: &pb.ObjectReference{
// 						ObjectType: "assign_permission",
// 						ObjectId:   permission,
// 					},
// 				},
// 			}

// 			updates = append(updates, &pb.RelationshipUpdate{
// 				Operation:    pb.RelationshipUpdate_OPERATION_DELETE,
// 				Relationship: relationship,
// 			})
// 		}
// 	}

// 	// Write relationship deletions to SpiceDB
// 	request := &pb.WriteRelationshipsRequest{
// 		Updates: updates,
// 	}
// 	res, err := spicedb.WriteRelationships(context.Background(), request)
// 	if err != nil {
// 		log.Printf("Failed to delete user-role, role-permission, and permission-action relationships: %s", err)
// 		return nil, err
// 	}
// 	return res, nil
// }

func DeleteUserRoleRelationship(userID string, roles []string, permissions []string, actions []string) (*pb.WriteRelationshipsResponse, error) {
    // Connect to SpiceDB
    spicedb, err := database.SpiceDB()
    if err != nil {
        log.Printf("Unable to connect to SpiceDB: %s", err)
        return nil, err
    }

    // Prepare relationship deletions
    var updates []*pb.RelationshipUpdate

    // Use a map to track unique relationships
    seenRelationships := make(map[string]bool)

    // Delete relationships for user-roles
    for _, role := range roles {
        relationshipKey := fmt.Sprintf("role:%s#%s@user:%s", role, role, userID)
        if !seenRelationships[relationshipKey] {
            relationship := &pb.Relationship{
                Resource: &pb.ObjectReference{
                    ObjectType: "role",
                    ObjectId:   role,
                },
                Relation: role, // This should match your role name
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
            seenRelationships[relationshipKey] = true
        }
    }

    // Delete relationships for role-permissions
    for _, permission := range permissions {
        for _, role := range roles {
            relationshipKey := fmt.Sprintf("assign_permission:%s#allows_action@role:%s", permission, role)
            if !seenRelationships[relationshipKey] {
                relationship := &pb.Relationship{
                    Resource: &pb.ObjectReference{
                        ObjectType: "assign_permission",
                        ObjectId:   permission,
                    },
                    Relation: "allows_action", // Matches your schema definition
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
                seenRelationships[relationshipKey] = true
            }
        }
    }

    // Delete relationships for permission-actions
    for _, action := range actions {
        for _, permission := range permissions {
            relationshipKey := fmt.Sprintf("action:%s#%s@assign_permission:%s", action, action, permission)
            if !seenRelationships[relationshipKey] {
                relationship := &pb.Relationship{
                    Resource: &pb.ObjectReference{
                        ObjectType: "action",
                        ObjectId:   action,
                    },
                    Relation: action, // Using the action value as the relation (Dynamic)
                    Subject: &pb.SubjectReference{
                        Object: &pb.ObjectReference{
                            ObjectType: "assign_permission",
                            ObjectId:   permission,
                        },
                    },
                }

                updates = append(updates, &pb.RelationshipUpdate{
                    Operation:    pb.RelationshipUpdate_OPERATION_DELETE,
                    Relationship: relationship,
                })
                seenRelationships[relationshipKey] = true
            }
        }
    }

    // Write relationship deletions to SpiceDB
    request := &pb.WriteRelationshipsRequest{
        Updates: updates,
    }
    res, err := spicedb.WriteRelationships(context.Background(), request)
    if err != nil {
        log.Printf("Failed to delete user-role, role-permission, and permission-action relationships: %s", err)
        return nil, err
    }
    return res, nil
}