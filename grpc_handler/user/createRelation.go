package user

import (
	"context"
	"log"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateRelationship(ctx context.Context, req *pb.CreateRelationshipRequest) (*pb.CreateRelationshipResponse, error) {
	if req.RoleName == "" || req.UserId == "" || req.ResourceName == "" || req.PrincipalId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Missing required fields in request")
	}
	err := client.CreateRelationship(
		req.RoleName,
		req.UserId,
		req.ResourceName,
		req.PrincipalId,
	)

	if err != nil {
		log.Printf("Error creating relationships: %v", err)
		return nil, err
	}

	relationshipString := "user" + ":" + req.UserId + "#" + req.RoleName + "@" + req.ResourceName + ":" + req.PrincipalId
	return &pb.CreateRelationshipResponse{
		StatusCode:    200,
		Message:       "Relationship created successfully",
		Success:       true,
		Data:          relationshipString,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
