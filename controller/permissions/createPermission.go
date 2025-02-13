package permissions

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionServer struct {
	pb.UnimplementedPermissionServiceServer
	PermissionRepo *repositories.PermissionRepository
}

func NewPermissionServer(permissionRepo *repositories.PermissionRepository) *PermissionServer {
	return &PermissionServer{
		PermissionRepo: permissionRepo,
	}
}

func (s *PermissionServer) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Permission cannot be nil")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Permission Name is required")
	}
	if err := s.PermissionRepo.CheckIfPermissionExists(ctx, req.Name); err != nil {
		return nil, err
	}

	newPermission := model.Permission{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.PermissionRepo.CreatePermission(ctx, &newPermission); err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:          newPermission.ID,
		Name:        newPermission.Name,
		Description: newPermission.Description,
	}
	return &pb.CreatePermissionResponse{
		StatusCode: http.StatusCreated,
		Message:    "Permission created successfully",
		Permission: pbPermission,
	}, nil
}
