package rolepermission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type ConnectRolePermissionServer struct {
	pb.UnimplementedConnectRolePermissionServiceServer
	DB *gorm.DB
}

func NewConnectRolePermissionServer(db *gorm.DB) *ConnectRolePermissionServer {
	return &ConnectRolePermissionServer{DB: db}
}

// func (s *ConnectRolePermissionServer) CreateConnectRolePermission(ctx context.Context, req *pb.CreateConnRolePermissionRequest) (*pb.CreateConnRolePermissionResponse, error) {
// 	if len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
// 		return &pb.CreateConnRolePermissionResponse{
// 			StatusCode: int32(codes.InvalidArgument),
// 			Message:    "Both role_ids and permission_ids are required",
// 		}, nil
// 	}

// 	var pbRolePermissions []*pb.ConnRolePermission

// 	for _, roleID := range req.RoleIds {
// 		rolePermission := model.RolePermission{
// 			RoleID: roleID,
// 		}

// 		if err := s.DB.Table("role_permissions").Create(&rolePermission).Error; err != nil {
// 			return &pb.CreateConnRolePermissionResponse{
// 				StatusCode: int32(codes.Internal),
// 				Message:    fmt.Sprintf("Failed to create RolePermission: %v", err),
// 			}, nil
// 		}

// 		// Associate permissions with the role
// 		for _, permissionID := range req.PermissionIds {
// 			permissionOnRole := model.PermissionOnRole{
// 				PermissionID: permissionID,
// 				UserRoleID:   rolePermission.ID, // Link to the RolePermission entry
// 			}

// 			if err := s.DB.Table("permission_on_roles").Create(&permissionOnRole).Error; err != nil {
// 				return &pb.CreateConnRolePermissionResponse{
// 					StatusCode: int32(codes.Internal),
// 					Message:    fmt.Sprintf("Failed to create PermissionOnRole: %v", err),
// 				}, nil
// 			}
// 		}

// 		// Prepare the response
// 		pbRolePermission := &pb.ConnRolePermission{
// 			Id:        rolePermission.ID,
// 			CreatedAt: rolePermission.CreatedAt.String(),
// 			UpdatedAt: rolePermission.UpdatedAt.String(),
// 		}

// 		pbRolePermissions = append(pbRolePermissions, pbRolePermission)
// 	}

//		return &pb.CreateConnRolePermissionResponse{
//			StatusCode: http.StatusCreated,
//			Message:    "RolePermissions created successfully",
//			ConnRolePermission: pbRolePermissions,
//		}, nil
//	}
// func (s *ConnectRolePermissionServer) CreateConnectRolePermission(ctx context.Context, req *pb.CreateConnRolePermissionRequest) (*pb.CreateConnRolePermissionResponse, error) {
// 	if len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
// 		return &pb.CreateConnRolePermissionResponse{
// 			StatusCode: int32(codes.InvalidArgument),
// 			Message:    "Both role_ids and permission_ids are required",
// 		}, nil
// 	}

// 	// Initialize a single ConnRolePermission object
// 	var connRolePermission pb.ConnRolePermission
// 	var permissionOnRoles []*pb.ConnPermissionOnRole

// 	for _, roleID := range req.RoleIds {
// 		rolePermission := model.RolePermission{
// 			RoleID: roleID,
// 		}
// 		if err := s.DB.Table("role_permissions").Create(&rolePermission).Error; err != nil {
// 			return &pb.CreateConnRolePermissionResponse{
// 				StatusCode: int32(codes.Internal),
// 				Message:    fmt.Sprintf("Failed to create RolePermission: %v", err),
// 			}, nil
// 		}

// 		// Associate permissions with the role
// 		for _, permissionID := range req.PermissionIds {
// 			permissionOnRole := model.PermissionOnRole{
// 				PermissionID: permissionID,
// 				UserRoleID:   rolePermission.ID, // Link to the RolePermission entry
// 			}
// 			if err := s.DB.Table("permission_on_roles").Create(&permissionOnRole).Error; err != nil {
// 				return &pb.CreateConnRolePermissionResponse{
// 					StatusCode: int32(codes.Internal),
// 					Message:    fmt.Sprintf("Failed to create PermissionOnRole: %v", err),
// 				}, nil
// 			}

// 			// Prepare the PermissionOnRole for the response
// 			pbPermissionOnRole := &pb.ConnPermissionOnRole{
// 				Id:         permissionOnRole.ID,
// 				CreatedAt:  permissionOnRole.CreatedAt.String(),
// 				UpdatedAt:  permissionOnRole.UpdatedAt.String(),
// 				UserRoleId: permissionOnRole.UserRoleID,
// 				Permission: &pb.ConnPermission{
// 					Id:          permissionOnRole.PermissionID,
// 					Name:        "SamplePermissionName",        // Replace with actual name if available
// 					Description: "SamplePermissionDescription", // Replace with actual description if available
// 				},
// 			}
// 			permissionOnRoles = append(permissionOnRoles, pbPermissionOnRole)
// 		}

// 		// Populate the ConnRolePermission object
// 		connRolePermission.Id = rolePermission.ID
// 		connRolePermission.CreatedAt = rolePermission.CreatedAt.String()
// 		connRolePermission.UpdatedAt = rolePermission.UpdatedAt.String()
// 		connRolePermission.Role = &pb.ConnRole{
// 			Id:          roleID,
// 			Name:        "SampleRoleName",        // Replace with actual name if available
// 			Description: "SampleRoleDescription", // Replace with actual description if available
// 		}
// 		connRolePermission.PermissionOnRoles = permissionOnRoles
// 	}

// 	return &pb.CreateConnRolePermissionResponse{
// 		StatusCode:         http.StatusCreated,
// 		Message:            "RolePermissions created successfully",
// 		ConnRolePermission: &connRolePermission,
// 	}, nil
// }

func (s *ConnectRolePermissionServer) CreateConnectRolePermission(ctx context.Context, req *pb.CreateConnRolePermissionRequest) (*pb.CreateConnRolePermissionResponse, error) {
	if len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
		return &pb.CreateConnRolePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Both role_ids and permission_ids are required",
		}, nil
	}

	// Initialize a single ConnRolePermission object
	var connRolePermission pb.ConnRolePermission
	var permissionOnRoles []*pb.ConnPermissionOnRole

	for _, roleID := range req.RoleIds {
		// Fetch the role details from the database
		var role model.Role
		result := s.DB.Table("roles").Where("id = ?", roleID).First(&role)
		if result.Error != nil {
			return &pb.CreateConnRolePermissionResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Role with ID %s not found", roleID),
			}, nil
		}

		// Create a RolePermission entry
		rolePermission := model.RolePermission{
			RoleID: roleID,
		}
		if err := s.DB.Table("role_permissions").Create(&rolePermission).Error; err != nil {
			return &pb.CreateConnRolePermissionResponse{
				StatusCode: int32(codes.Internal),
				Message:    fmt.Sprintf("Failed to create RolePermission: %v", err),
			}, nil
		}

		// Associate permissions with the role
		for _, permissionID := range req.PermissionIds {
			// Fetch the permission details from the database
			var permission model.Permission
			result := s.DB.Table("permissions").Where("id = ?", permissionID).First(&permission)
			if result.Error != nil {
				return &pb.CreateConnRolePermissionResponse{
					StatusCode: int32(codes.NotFound),
					Message:    fmt.Sprintf("Permission with ID %s not found", permissionID),
				}, nil
			}

			// Create a PermissionOnRole entry
			permissionOnRole := model.PermissionOnRole{
				PermissionID: permissionID,
				UserRoleID:   rolePermission.ID, // Link to the RolePermission entry
			}
			if err := s.DB.Table("permission_on_roles").Create(&permissionOnRole).Error; err != nil {
				return &pb.CreateConnRolePermissionResponse{
					StatusCode: int32(codes.Internal),
					Message:    fmt.Sprintf("Failed to create PermissionOnRole: %v", err),
				}, nil
			}

			// Prepare the PermissionOnRole for the response
			pbPermissionOnRole := &pb.ConnPermissionOnRole{
				Id:         permissionOnRole.ID,
				CreatedAt:  permissionOnRole.CreatedAt.String(),
				UpdatedAt:  permissionOnRole.UpdatedAt.String(),
				UserRoleId: permissionOnRole.UserRoleID,
				Permission: &pb.ConnPermission{
					Id:          permission.ID,
					Name:        permission.Name,
					Description: permission.Description,
				},
			}
			permissionOnRoles = append(permissionOnRoles, pbPermissionOnRole)
		}

		// Populate the ConnRolePermission object with real data
		connRolePermission.Id = rolePermission.ID
		connRolePermission.CreatedAt = rolePermission.CreatedAt.String()
		connRolePermission.UpdatedAt = rolePermission.UpdatedAt.String()
		connRolePermission.Role = &pb.ConnRole{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
		connRolePermission.PermissionOnRoles = permissionOnRoles
	}

	return &pb.CreateConnRolePermissionResponse{
		StatusCode:         http.StatusCreated,
		Message:            "RolePermissions created successfully",
		ConnRolePermission: &connRolePermission,
	}, nil
}
