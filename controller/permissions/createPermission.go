package permissions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/entities/models"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type PermissionServer struct {
	pb.UnimplementedPermissionServiceServer
	DB *gorm.DB
}

func NewPermissionServer(db *gorm.DB) *PermissionServer {
	return &PermissionServer{DB: db}
}

func (s *PermissionServer) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	permission := req
	if permission == nil {
		return &pb.CreatePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Permission cannot be nil",
		}, nil
	}
	if permission.Name == "" {
		return &pb.CreatePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Permission Name is required",
		}, nil
	}

	existingPermission := models.Permission{}
	result := s.DB.Table("permissions").Where("name = ?", permission.Name).First(&existingPermission)
	if result.Error == nil {
		return &pb.CreatePermissionResponse{
			StatusCode: int32(codes.AlreadyExists),
			Message:    fmt.Sprintf("Permission with name %s already exists", permission.Name),
		}, nil
	}

	newPermission := models.NewPermission(permission.Name, permission.Description)
	if err := s.DB.Table("permissions").Create(&newPermission).Error; err != nil {
		return &pb.CreatePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to create permission: %v", err),
		}, nil
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
