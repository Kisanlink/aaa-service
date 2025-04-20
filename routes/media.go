package routes

import (
	"github.com/Kisanlink/aaa-service/media"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MediaRoutes(r *gin.RouterGroup, database *gorm.DB) {
	mediaServiceAWS, err := media.NewMediaServiceAWS()
	if err != nil {
		panic("Failed to create media service: " + err.Error())
	}
	r.POST("/upload/base64", media.UploadBase64FileHandlerAWS(mediaServiceAWS))
	r.POST("/upload", media.UploadSingleFileHandlerAWS(mediaServiceAWS))
	r.POST("/bulk-upload", media.UploadMultipleFilesHandlerAWS(mediaServiceAWS))
}
