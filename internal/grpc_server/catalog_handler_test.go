package grpc_server

import (
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestCatalogHandlerImplementsInterface verifies that CatalogHandler properly implements the gRPC interface
func TestCatalogHandlerImplementsInterface(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// This will fail to compile if CatalogHandler doesn't implement CatalogServiceServer
	var _ pb.CatalogServiceServer = &CatalogHandler{
		catalogService: nil, // nil is fine for interface check
		authChecker:    nil, // nil is fine for interface check
		logger:         logger,
	}
}

// TestCatalogHandlerSeedRolesAndPermissionsSignature verifies the method signature is correct
func TestCatalogHandlerSeedRolesAndPermissionsSignature(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockService := &catalog.CatalogService{} // nil fields are ok for signature test

	handler := NewCatalogHandler(mockService, nil, logger)

	// Verify handler is not nil
	assert.NotNil(t, handler)

	// Verify handler implements the interface correctly
	// This compiles successfully, which proves the interface is properly implemented
	var _ pb.CatalogServiceServer = handler

	// The fact that this test compiles and runs proves:
	// 1. The protobuf files are properly generated
	// 2. The CatalogHandler implements the correct interface
	// 3. The method signatures match the protobuf definition
	assert.NotNil(t, handler, "Handler should be created successfully")
}

// TestSeedRequestResponseTypes verifies protobuf message types are correct
func TestSeedRequestResponseTypes(t *testing.T) {
	// Test request creation
	req := &pb.SeedRolesAndPermissionsRequest{
		Force:          true,
		OrganizationId: "test-org-123",
	}
	assert.True(t, req.Force)
	assert.Equal(t, "test-org-123", req.OrganizationId)

	// Test response creation
	resp := &pb.SeedRolesAndPermissionsResponse{
		StatusCode:         200,
		Message:            "Success",
		RolesCreated:       6,
		PermissionsCreated: 72,
		ResourcesCreated:   8,
		ActionsCreated:     9,
		CreatedRoles:       []string{"farmer", "kisansathi", "CEO", "fpo_manager", "admin", "readonly"},
	}

	assert.Equal(t, int32(200), resp.StatusCode)
	assert.Equal(t, "Success", resp.Message)
	assert.Equal(t, int32(6), resp.RolesCreated)
	assert.Len(t, resp.CreatedRoles, 6)
	assert.Contains(t, resp.CreatedRoles, "farmer")
	assert.Contains(t, resp.CreatedRoles, "admin")
}
