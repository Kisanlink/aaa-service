package roles

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"gorm.io/gorm"
)

type RoleServer struct {
	pb.UnimplementedRoleServiceServer
	DB *gorm.DB
}

func NewRoleServer(db *gorm.DB) *RoleServer {
	return &RoleServer{DB: db}
}

func (s *RoleServer) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	if req.Name == "" {
		return &pb.CreateRoleResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Role Name is required",
		}, nil
	}
	if s.DB == nil {
		return &pb.CreateRoleResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Database connection is not initialized",
		}, nil
	}

	// Check if a role with the same name already exists
	existingRole := model.Role{}
	result := s.DB.Table("roles").Where("name = ?", req.Name).First(&existingRole)
	if result.Error == nil {
		return &pb.CreateRoleResponse{
			StatusCode: http.StatusConflict,
			Message:    fmt.Sprintf("Role with name '%s' already exists", req.Name),
		}, nil
	}

	// Create a new role
	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.DB.Table("roles").Create(&newRole).Error; err != nil {
		return &pb.CreateRoleResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to create role: %v", err),
		}, nil
	}

	// Convert the created role to a protobuf-compatible structure
	pbRole := &pb.Role{
		Id:          newRole.ID, // Assuming ID is a UUID or similar
		Name:        newRole.Name,
		Description: newRole.Description,
	}

	// Return the success response with the created role
	return &pb.CreateRoleResponse{
		StatusCode: http.StatusCreated,
		Message:    "Role created successfully",
		Role:       pbRole, // Assign the protobuf-compatible role here
	}, nil
}
