package media

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func UploadBase64FileHandlerAWS(mediaService *MediaServiceAWS) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			BucketName  string `json:"bucket_name" binding:"required"`
			ImageBase64 string `json:"image_base64" binding:"required"`
			FileName    string `json:"file_name" binding:"required"`
			Folder      string `json:"folder" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			errorMsg := "Missing required fields: "
			var missingFields []string

			if request.BucketName == "" {
				missingFields = append(missingFields, "bucket_name")
			}
			if request.ImageBase64 == "" {
				missingFields = append(missingFields, "image_base64")
			}
			if request.FileName == "" {
				missingFields = append(missingFields, "file_name")
			}
			if request.Folder == "" {
				missingFields = append(missingFields, "folder")
			}

			errorMsg += strings.Join(missingFields, ", ")
			c.JSON(http.StatusBadRequest, newResponse(http.StatusBadRequest, false, errorMsg, nil, nil))
			return
		}

		response := mediaService.UploadBase64FileAWS(request.BucketName, request.ImageBase64, request.FileName, request.Folder)
		c.JSON(response.StatusCode, response)
	}
}

func UploadSingleFileHandlerAWS(mediaService *MediaServiceAWS) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate required fields
		var missingFields []string

		bucketName := c.PostForm("bucket_name")
		if bucketName == "" {
			missingFields = append(missingFields, "bucket_name")
		}

		folder := c.PostForm("folder")
		if folder == "" {
			missingFields = append(missingFields, "folder")
		}

		file, err := c.FormFile("file")
		if err != nil {
			missingFields = append(missingFields, "file")
		}

		if len(missingFields) > 0 {
			errorMsg := "Missing required fields: " + strings.Join(missingFields, ", ")
			c.JSON(http.StatusBadRequest, newResponse(http.StatusBadRequest, false, errorMsg, nil, nil))
			return
		}

		// Process file upload
		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, newResponse(http.StatusInternalServerError, false, "Unable to open file", nil, err.Error()))
			return
		}
		defer openedFile.Close()

		fileBytes, err := io.ReadAll(openedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, newResponse(http.StatusInternalServerError, false, "Unable to read file", nil, err.Error()))
			return
		}

		response := mediaService.UploadFile(bucketName, fileBytes, file.Filename, folder)
		c.JSON(response.StatusCode, response)
	}
}

func UploadMultipleFilesHandlerAWS(mediaService *MediaServiceAWS) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate required fields
		var missingFields []string

		bucketName := c.PostForm("bucket_name")
		if bucketName == "" {
			missingFields = append(missingFields, "bucket_name")
		}

		folder := c.PostForm("folder")
		if folder == "" {
			missingFields = append(missingFields, "folder")
		}

		form, err := c.MultipartForm()
		if err != nil {
			missingFields = append(missingFields, "files")
		} else if len(form.File["files"]) == 0 {
			missingFields = append(missingFields, "files")
		}

		if len(missingFields) > 0 {
			errorMsg := "Missing required fields: " + strings.Join(missingFields, ", ")
			c.JSON(http.StatusBadRequest, newResponse(http.StatusBadRequest, false, errorMsg, nil, nil))
			return
		}

		files := form.File["files"]
		if len(files) > 10 {
			c.JSON(http.StatusBadRequest, newResponse(http.StatusBadRequest, false, "Maximum 10 files allowed", nil, nil))
			return
		}

		// Process files
		var fileBytes [][]byte
		var fileNames []string

		for _, file := range files {
			openedFile, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, newResponse(http.StatusInternalServerError, false, "Unable to open file", nil, err.Error()))
				return
			}

			bytes, err := io.ReadAll(openedFile)
			openedFile.Close()
			if err != nil {
				c.JSON(http.StatusInternalServerError, newResponse(http.StatusInternalServerError, false, "Unable to read file", nil, err.Error()))
				return
			}

			fileBytes = append(fileBytes, bytes)
			fileNames = append(fileNames, file.Filename)
		}

		response := mediaService.UploadFiles(bucketName, fileBytes, fileNames, folder)
		c.JSON(response.StatusCode, response)
	}
}
