package routes

import (
	"github.com/Kisanlink/aaa-service/handler/spicedb"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SpiceDBRoutes(r *gin.RouterGroup, database *gorm.DB, s spicedb.SpiceDBHandler) {
	r.GET("/update/schema", s.UpdateSpiceDb)
}
