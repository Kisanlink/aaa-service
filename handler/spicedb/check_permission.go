package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// CheckPermissionSpiceDB checks if a user has permission to perform an action on a resource
// @Summary Check user permission
// @Description Checks if a user has permission to perform a specific action on a resource
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Param request body model.CheckPermissionRequest true "Permission Check Request"
// @Success 200 {object} helper.Response{data=bool} "Permission check result"
// @Failure 400 {object} helper.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} helper.ErrorResponse "Failed to check permission"
// @Router /check-permission [post]
func (h *SpiceDBHandler) CheckPermissionSpiceDB(c *gin.Context) {
	var req model.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request parameters: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request parameters"})
		return
	}

	// Call the CheckPermission function
	hasPermission, err := client.CheckPermission(
		req.Username,
		req.Action,
		req.ResourceType,
		req.ResourceID,
	)
	if err != nil {
		log.Printf("Permission check failed: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission check successful", hasPermission)
}
