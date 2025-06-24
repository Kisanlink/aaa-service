package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// CreateRelation
// @Summary Create a relationship in SpiceDB
// @Description Creates a new relationship between a user and a resource in SpiceDB
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Param request body model.CreateRelationshipRequest true "Relationship creation request"
// @Success 200 {object} helper.Response{data=string} "Relationship created successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request format or missing fields"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /relation [post]
func (h *SpiceDBHandler) CreateRelation(c *gin.Context) {
	var req model.CreateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request format"})
		return
	}
	if req.RoleName == "" || req.UserID == "" || req.ResourceName == "" || req.PrincipalID == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing required fields in request"})
		return
	}
	err := client.CreateRelationship(
		req.RoleName,
		req.UserID,
		req.ResourceName,
		req.PrincipalID,
	)

	if err != nil {
		log.Printf("Error creating relationships: %v", err)
		return
	}

	relationshipString := "user" + ":" + req.UserID + "#" + req.RoleName + "@" + req.ResourceName + ":" + req.PrincipalID

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Relationship Created successfully", relationshipString)
}
