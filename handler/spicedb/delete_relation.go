package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// DeleteRelation
// @Summary Delete a relationship in SpiceDB
// @Description Deletes an existing relationship between a user and a resource in SpiceDB
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Param request body model.CreateRelationshipRequest true "Relationship deletion request"
// @Success 200 {object} helper.Response{data=string} "Relationship deleted successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request format or missing fields"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /relation [delete]
func (h *SpiceDBHandler) DeleteRelation(c *gin.Context) {
	var req model.CreateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request format"})
		return
	}
	if req.Relation == "" || req.Username == "" || req.Resource == "" || req.ResourceID == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing required fields in request"})
		return
	}
	err := client.DeleteRelationship(
		req.Relation,
		req.Username,
		req.Resource,
		req.ResourceID,
	)

	if err != nil {
		log.Printf("Error creating relationships: %v", err)
		return
	}

	relationshipString := "user" + ":" + req.Username + "#" + req.Relation + "@" + req.Resource + ":" + req.ResourceID

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Relationship deleted successfully", relationshipString)
}
