package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// ReadSpiceDB schema
// @Summary Read spice db schema
// @Description Read schema from SpiceDB
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=map[string][]string} "Schema Read successfully"
// @Failure 500 {object} helper.ErrorResponse "Failed to Read SpiceDB schema"
// @Router /read/schema [get]
func (h *SpiceDBHandler) ReadSpiceDB(c *gin.Context) {
	// Generate SpiceDB schema definitions
	data, err := client.ReadSchema()
	if err != nil {
		log.Printf("Failed to update SpiceDB schema: %v", err)

	}
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Schema Read successfully", data)
}
