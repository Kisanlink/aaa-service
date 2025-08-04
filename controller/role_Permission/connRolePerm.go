package rolepermission

import (
	"context"

	pb "github.com/Kisanlink/aaa-service/proto"
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
	// TODO: Implement this functionality after the model refactoring is complete
	// This requires RolePermission and PermissionOnRole models to be properly defined
	return &pb.CreateConnRolePermissionResponse{
		StatusCode: int32(codes.Unimplemented),
		Message:    "This functionality is temporarily disabled during model refactoring",
	}, nil
}
