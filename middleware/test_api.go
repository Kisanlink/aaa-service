package middleware

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// PermissionTestHandler handles permission test endpoints
type PermissionTestHandler struct {
	// Add any dependencies if needed
}

// NewPermissionTestHandler creates a new PermissionTestHandler
func NewPermissionTestHandler() *PermissionTestHandler {
	return &PermissionTestHandler{}
}

// @Summary Test GET permission
// @Description Tests the GeneralPermissionCheck middleware with GET request
// @Tags Middleware
// @Accept json
// @Produce json
// @Param x-user-id header string true "User ID"
// @Param x-resource-name header string true "Resource Name"
// @Param x-principal-id header string true "Resource ID"
// @Param x-action header string true "Action"
// @Success 200 {object} helper.Response{data=string} "Success message"
// @Failure 400 {object} helper.ErrorResponse "Missing required headers"
// @Failure 403 {object} helper.ErrorResponse "Permission denied"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /test-permission-get [get]
func (h *PermissionTestHandler) TestPermissionGET(c *gin.Context) {
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission granted for GET request", nil)
}

// @Summary Test POST permission
// @Description Tests the CanCreatePermission middleware with POST request
// @Tags Middleware
// @Accept json
// @Produce json
// @Param x-user-id header string true "User ID"
// @Param x-resource-name header string true "Resource Name"
// @Param x-action header string true "Action"
// @Param x-role-name header string true "Role Name"
// @Success 200 {object} helper.Response{data=string} "Success message"
// @Failure 400 {object} helper.ErrorResponse "Missing required headers"
// @Failure 403 {object} helper.ErrorResponse "Permission denied"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /test-permission-post [post]
func (h *PermissionTestHandler) TestPermissionPOST(c *gin.Context) {
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission granted for POST request", nil)
}
