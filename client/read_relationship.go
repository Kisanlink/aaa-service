package client

import (
	"context"
	"io"
	"log"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func ReadRelationshipsByUserID(userID string) ([]*pb.Relationship, error) {
	//  Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	//  Define the subject filter (user)
	subjectFilter := &pb.SubjectFilter{
		SubjectType: "user",
		OptionalSubjectId: userID,
	}

	//  Prepare the request to read relationships
	request := &pb.ReadRelationshipsRequest{
		Consistency: &pb.Consistency{
			Requirement: &pb.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
		RelationshipFilter: &pb.RelationshipFilter{
			OptionalSubjectFilter: subjectFilter,
		},
	}

	//  Call the ReadRelationships API
	stream, err := spicedb.ReadRelationships(context.Background(), request)
	if err != nil {
		log.Printf("Failed to read relationships for user %s: %s", userID, err)
		return nil, err
	}

	//  Collect all relationships
	var relationships []*pb.Relationship
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error while reading relationships: %s", err)
			return nil, err
		}

		// Append the relationship to the list
		relationships = append(relationships, resp.Relationship)
	}

	log.Printf("Found %d relationships for user %s", len(relationships), userID)
	return relationships, nil
}

