package rolepermission

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	pb "github.com/Kisanlink/aaa-service/proto"
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
	var queryResults []QueryResult

	err := s.DB.Raw(`
         SELECT
            rp.id AS id,
            rp.created_at AS created_at,
            json_build_object(
                'id', r.id,
                'name', r.name,
                'description', r.description,
                'created_at', r.created_at
            ) AS role,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', por.id,
                        'created_at', por.created_at,
                        'name', p.name,
                        'description', p.description,
                        'permission_id', p.id
                    )
                ) FILTER (WHERE por.id IS NOT NULL),
                '[]'
            ) AS permissions
        FROM
            role_permissions rp
        JOIN
            roles r ON rp.role_id = r.id
        LEFT JOIN
            permission_on_roles por ON rp.id = por.user_role_id
        LEFT JOIN
            permissions p ON por.permission_id = p.id
        GROUP BY
            rp.id, rp.created_at, r.id, r.name, r.description, r.created_at;
    `).Scan(&queryResults).Error

	if err != nil {
		return nil, err
	}

	if len(queryResults) == 0 {
		return &pb.GetConnRolePermissionallResponse{
			StatusCode: http.StatusOK,
			Message:    "No RolePermissions found",
		}, nil
	}

	var rolePermissionResponses []RolePermissionResponse
	for _, result := range queryResults {
		var role Role
		if err := json.Unmarshal([]byte(result.Role), &role); err != nil {
			return nil, err
		}

		var permissions []Permission
		if err := json.Unmarshal([]byte(result.Permissions), &permissions); err != nil {
			return nil, err
		}

		rolePermissionResponses = append(rolePermissionResponses, RolePermissionResponse{
			ID:          result.ID,
			CreatedAt:   result.CreatedAt,
			Role:        role,
			Permissions: permissions,
		})
	}

	var connRolePermissions []*pb.ConnRolePermission
	for _, resp := range rolePermissionResponses {
		permissions := make([]*pb.ConnPermissionOnRole, len(resp.Permissions))
		for i, perm := range resp.Permissions {
			permissions[i] = &pb.ConnPermissionOnRole{
				Id:        perm.ID,
				CreatedAt: perm.CreatedAt,
				Permission: &pb.ConnPermission{
					Id:          perm.PermissionID,
					Name:        perm.Name,
					Description: perm.Description,
				},
			}
		}
		connRolePermissions = append(connRolePermissions, &pb.ConnRolePermission{
			Id:        resp.ID,
			CreatedAt: resp.CreatedAt.Format(time.RFC3339),
			Role: &pb.ConnRole{
				Id:          resp.Role.ID,
				Name:        resp.Role.Name,
				Description: resp.Role.Description,
			},
			PermissionOnRoles: permissions,
		})
	}

	return &pb.GetConnRolePermissionallResponse{
		StatusCode:         http.StatusOK,
		Message:            "RolePermissions fetched successfully",
		ConnRolePermission: connRolePermissions,
	}, nil
}
