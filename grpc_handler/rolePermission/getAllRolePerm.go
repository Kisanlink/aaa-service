package rolepermission

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetAllRolePermission(ctx context.Context, req *pb.GetConnRolePermissionallRequest) (*pb.GetConnRolePermissionallResponse, error) {
	rolePermissions, err := s.rolePermService.GetAllRolePermissions()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
	}

	var responseData []*pb.ConnRolePermissionResponse
	for _, rp := range rolePermissions {
		if helper.IsZeroValued(rp.Role) || helper.IsZeroValued(rp.Permission) || rp.Permission.ID == "" {
			continue
		}

		responseData = append(responseData, &pb.ConnRolePermissionResponse{
			Id:        rp.ID,
			CreatedAt: rp.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: rp.UpdatedAt.Format(time.RFC3339Nano),
			Role: &pb.ConnRole{
				Id:          rp.Role.ID,
				Name:        rp.Role.Name,
				Description: rp.Role.Description,
				Source:      rp.Role.Source,
				CreatedAt:   rp.Role.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   rp.Role.UpdatedAt.Format(time.RFC3339Nano),
			},
			Permissions: []*pb.ConnPermission{{
				Id:             rp.Permission.ID,
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				Resource:       rp.Permission.Resource,
				Source:         rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime.Format(time.RFC3339Nano),
				ValidEndTime:   rp.Permission.ValidEndTime.Format(time.RFC3339Nano),
				CreatedAt:      rp.Permission.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:      rp.Permission.UpdatedAt.Format(time.RFC3339Nano),
			}},
			IsActive: rp.IsActive,
		})
	}

	return &pb.GetConnRolePermissionallResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role with Permissions fetched successfully",
		Data:          responseData,
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
	}, nil
}
