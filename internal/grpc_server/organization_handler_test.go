package grpc_server

import (
	"context"
	"errors"
	"testing"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock OrganizationService
type mockOrganizationService struct {
	getOrgFunc         func(ctx context.Context, orgID string) (interface{}, error)
	getOrgGroupsFunc   func(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error)
	createGroupFunc    func(ctx context.Context, orgID string, req interface{}) (interface{}, error)
	addUserToGroupFunc func(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error)
}

func (m *mockOrganizationService) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	if m.getOrgFunc != nil {
		return m.getOrgFunc(ctx, orgID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
	if m.getOrgGroupsFunc != nil {
		return m.getOrgGroupsFunc(ctx, orgID, limit, offset, includeInactive)
	}
	return []interface{}{}, nil
}

func (m *mockOrganizationService) CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	if m.createGroupFunc != nil {
		return m.createGroupFunc(ctx, orgID, req)
	}
	return map[string]interface{}{"id": "group-123"}, nil
}

func (m *mockOrganizationService) AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
	if m.addUserToGroupFunc != nil {
		return m.addUserToGroupFunc(ctx, orgID, groupID, userID, req)
	}
	return map[string]interface{}{}, nil
}

// Implement remaining interface methods as stubs
func (m *mockOrganizationService) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) UpdateOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool, orgType string) ([]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) ActivateOrganization(ctx context.Context, orgID string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) DeactivateOrganization(ctx context.Context, orgID string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) GetOrganizationStats(ctx context.Context, orgID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error {
	return errors.New("not implemented")
}

func (m *mockOrganizationService) GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockOrganizationService) CountOrganizations(ctx context.Context, includeInactive bool, orgType string) (int64, error) {
	return 0, errors.New("not implemented")
}

// Mock GroupService
type mockGroupService struct {
	getUserGroupsFunc         func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error)
	removeMemberFunc          func(ctx context.Context, groupID, principalID string, removedBy string) error
	getUserEffectiveRolesFunc func(ctx context.Context, orgID, userID string) (interface{}, error)
}

func (m *mockGroupService) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	if m.getUserGroupsFunc != nil {
		return m.getUserGroupsFunc(ctx, orgID, userID, limit, offset)
	}
	return []interface{}{}, nil
}

func (m *mockGroupService) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	if m.removeMemberFunc != nil {
		return m.removeMemberFunc(ctx, groupID, principalID, removedBy)
	}
	return nil
}

func (m *mockGroupService) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	if m.getUserEffectiveRolesFunc != nil {
		return m.getUserEffectiveRolesFunc(ctx, orgID, userID)
	}
	return []interface{}{}, nil
}

// Implement remaining interface methods as stubs
func (m *mockGroupService) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	return errors.New("not implemented")
}

func (m *mockGroupService) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	return errors.New("not implemented")
}

func (m *mockGroupService) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGroupService) CountGroupMembers(ctx context.Context, groupID string) (int64, error) {
	return 0, errors.New("not implemented")
}

func (m *mockGroupService) CountGroups(ctx context.Context, organizationID string, includeInactive bool) (int64, error) {
	return 0, errors.New("not implemented")
}

// Mock organization response
type mockOrgResponse struct {
	IsActive bool
}

func (m *mockOrgResponse) GetIsActive() bool {
	return m.IsActive
}

// Test AddUserToOrganization

func TestAddUserToOrganization_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
		getOrgGroupsFunc: func(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "members-group-123", "name": "Members"},
			}, nil
		},
		addUserToGroupFunc: func(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.OrganizationUser.UserId != "user-456" {
		t.Errorf("Expected user ID user-456, got %s", resp.OrganizationUser.UserId)
	}
}

func TestAddUserToOrganization_EmptyOrgID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService: &mockOrganizationService{},
		logger:     logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	if st, ok := status.FromError(err); ok {
		if st.Code() != codes.InvalidArgument {
			t.Errorf("Expected code InvalidArgument, got %v", st.Code())
		}
	}
}

func TestAddUserToOrganization_EmptyUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService: &mockOrganizationService{},
		logger:     logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}
}

func TestAddUserToOrganization_NoGroupService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService:   &mockOrganizationService{},
		groupService: nil,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 503 {
		t.Errorf("Expected status code 503, got %d", resp.StatusCode)
	}

	if st, ok := status.FromError(err); ok {
		if st.Code() != codes.Unavailable {
			t.Errorf("Expected code Unavailable, got %v", st.Code())
		}
	}
}

func TestAddUserToOrganization_OrgNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return nil, errors.New("organization not found")
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-999",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", resp.StatusCode)
	}
}

func TestAddUserToOrganization_InactiveOrg(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: false}, nil
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}
}

func TestAddUserToOrganization_CreateMembersGroup(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
		getOrgGroupsFunc: func(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
			return []interface{}{}, nil // No groups exist
		},
		createGroupFunc: func(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
			return map[string]interface{}{"id": "new-members-group"}, nil
		},
		addUserToGroupFunc: func(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
			if groupID != "new-members-group" {
				t.Errorf("Expected group ID new-members-group, got %s", groupID)
			}
			return map[string]interface{}{}, nil
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestAddUserToOrganization_AlreadyMember(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
		getOrgGroupsFunc: func(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "members-group-123", "name": "Members"},
			}, nil
		},
		addUserToGroupFunc: func(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
			return nil, errors.New("user is already a member of this group")
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.AddUserToOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.AddUserToOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error for idempotent operation, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Message != "User is already a member of organization" {
		t.Errorf("Expected message about existing membership, got %s", resp.Message)
	}
}

// Test ValidateOrganizationAccess

func TestValidateOrganizationAccess_BasicMembershipSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123", "name": "Members"},
			}, nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Error("Expected access to be allowed")
	}
}

func TestValidateOrganizationAccess_DenyNonMember(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{}, nil // No groups
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected access to be denied")
	}

	if len(resp.Reasons) == 0 {
		t.Error("Expected denial reason")
	}
}

func TestValidateOrganizationAccess_EmptyUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService: &mockOrganizationService{},
		logger:     logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "",
		OrganizationId: "org-123",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.Allowed {
		t.Error("Expected access to be denied")
	}
}

func TestValidateOrganizationAccess_InactiveOrg(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: false}, nil
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected access to be denied for inactive org")
	}
}

func TestValidateOrganizationAccess_WithPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123"},
			}, nil
		},
		getUserEffectiveRolesFunc: func(ctx context.Context, orgID, userID string) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{
					"permissions": []interface{}{"resource:read", "resource:write"},
				},
			}, nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
		ResourceType:   "resource",
		Action:         "read",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Error("Expected access to be allowed with permission")
	}
}

func TestValidateOrganizationAccess_DenyWithoutPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123"},
			}, nil
		},
		getUserEffectiveRolesFunc: func(ctx context.Context, orgID, userID string) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{
					"permissions": []interface{}{"other:read"},
				},
			}, nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
		ResourceType:   "resource",
		Action:         "write",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Allowed {
		t.Error("Expected access to be denied without permission")
	}
}

func TestValidateOrganizationAccess_WildcardPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123"},
			}, nil
		},
		getUserEffectiveRolesFunc: func(ctx context.Context, orgID, userID string) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{
					"permissions": []interface{}{"*:*"},
				},
			}, nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.ValidateOrganizationAccessRequest{
		UserId:         "user-456",
		OrganizationId: "org-123",
		ResourceType:   "anything",
		Action:         "write",
	}

	resp, err := handler.ValidateOrganizationAccess(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Allowed {
		t.Error("Expected access to be allowed with wildcard permission")
	}
}

// Test RemoveUserFromOrganization

func TestRemoveUserFromOrganization_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123"},
				map[string]interface{}{"id": "group-456"},
			}, nil
		},
		removeMemberFunc: func(ctx context.Context, groupID, principalID string, removedBy string) error {
			return nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestRemoveUserFromOrganization_EmptyOrgID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService: &mockOrganizationService{},
		logger:     logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	if resp.Success {
		t.Error("Expected success to be false")
	}
}

func TestRemoveUserFromOrganization_NoGroupService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	handler := &OrganizationHandler{
		orgService:   &mockOrganizationService{},
		groupService: nil,
		logger:       logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 503 {
		t.Errorf("Expected status code 503, got %d", resp.StatusCode)
	}
}

func TestRemoveUserFromOrganization_OrgNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return nil, errors.New("organization not found")
		},
	}

	mockGroupService := &mockGroupService{}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "org-999",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", resp.StatusCode)
	}
}

func TestRemoveUserFromOrganization_NotAMember(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{}, nil // User has no groups
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if !resp.Success {
		t.Error("Expected success to be true (idempotent)")
	}

	if resp.Message != "User is not a member of organization" {
		t.Errorf("Expected message about not being a member, got %s", resp.Message)
	}
}

func TestRemoveUserFromOrganization_PartialFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockOrgService := &mockOrganizationService{
		getOrgFunc: func(ctx context.Context, orgID string) (interface{}, error) {
			return &mockOrgResponse{IsActive: true}, nil
		},
	}

	callCount := 0
	mockGroupService := &mockGroupService{
		getUserGroupsFunc: func(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": "group-123"},
				map[string]interface{}{"id": "group-456"},
			}, nil
		},
		removeMemberFunc: func(ctx context.Context, groupID, principalID string, removedBy string) error {
			callCount++
			if callCount == 2 {
				return errors.New("removal failed")
			}
			return nil
		},
	}

	handler := &OrganizationHandler{
		orgService:   mockOrgService,
		groupService: mockGroupService,
		logger:       logger,
	}

	req := &pb.RemoveUserFromOrganizationRequest{
		OrganizationId: "org-123",
		UserId:         "user-456",
	}

	resp, err := handler.RemoveUserFromOrganization(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for partial failure, got nil")
	}

	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	if resp.Success {
		t.Error("Expected success to be false")
	}
}
