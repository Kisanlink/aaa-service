package rolepermission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Permission struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	PermissionID string `json:"permission_id"`
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type RolePermissionResponse struct {
	ID          string       `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
}

type QueryResult struct {
	ID          string `json:"id"`
	CreatedAt   time.Time
	Role        string
	Permissions string
}

func (s *ConnectRolePermissionServer) GetAllRolePermission(ctx context.Context, req *pb.GetConnRolePermissionallRequest) (*pb.GetConnRolePermissionallResponse, error) {
	queryResults, err := s.RolePermissionRepo.GetAllRolePermissions(ctx)
	if err != nil {
		return nil, err
	}
	if len(queryResults) == 0 {
		return &pb.GetConnRolePermissionallResponse{
			StatusCode: int32(http.StatusOK),
			Message:    "No RolePermissions found",
		}, nil
	}

	var connRolePermissions []*pb.ConnRolePermission
	for _, result := range queryResults {
		var role pb.ConnRole
		if err := json.Unmarshal([]byte(result.Role), &role); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to parse role JSON: %v", err))
		}

		var permissions []*pb.ConnPermissionOnRole
		if err := json.Unmarshal([]byte(result.Permissions), &permissions); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to parse permissions JSON: %v", err))
		}

		connRolePermissions = append(connRolePermissions, &pb.ConnRolePermission{
			Id:                result.ID,
			CreatedAt:         result.CreatedAt.Format(time.RFC3339),
			Role:              &role,
			PermissionOnRoles: permissions,
		})
	}

	return &pb.GetConnRolePermissionallResponse{
		StatusCode:         int32(http.StatusOK),
		Message:            "RolePermissions fetched successfully",
		ConnRolePermission: connRolePermissions,
	}, nil
}
