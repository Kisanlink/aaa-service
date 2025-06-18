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
	data, err := client.ReadSchema()
	if err != nil {
		log.Printf("Failed to read SpiceDB schema: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to read schema"})
		return
	}
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Schema Read successfully", data)
}
