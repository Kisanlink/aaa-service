package media

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

type MediaServiceAWS struct {
	s3Client *s3.S3
	region   string
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Error      any    `json:"error"`
	TimeStamp  string `json:"timestamp"`
}

func NewMediaServiceAWS() (*MediaServiceAWS, error) {
	region := os.Getenv("AWS_REGION") // e.g., "ap-south-1"

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("invalid AWS credentials: %v", err)
	}

	return &MediaServiceAWS{
		s3Client: s3.New(sess),
		region:   region,
	}, nil
}

func newResponse(statusCode int, success bool, message string, data any, err any) Response {
	return Response{
		StatusCode: statusCode,
		Success:    success,
		Message:    message,
		Data:       data,
		Error:      err,
		TimeStamp:  time.Now().Format(time.RFC3339),
	}
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".tiff":
		return "application/octet-stream"
	default:
		return "application/json"
	}
}

func (m *MediaServiceAWS) UploadBase64FileAWS(bucketName, base64Data, fileName, folder string) Response {
	fileBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return newResponse(http.StatusBadRequest, false, "Failed to decode base64 string", nil, err.Error())
	}

	contentType := http.DetectContentType(fileBytes)
	if len(fileBytes) >= 512 {
		contentType = http.DetectContentType(fileBytes[:512])
	}

	var fileExtension string
	switch contentType {
	case "image/png":
		fileExtension = ".png"
	case "image/jpeg":
		fileExtension = ".jpg"
	case "image/gif":
		fileExtension = ".gif"
	case "application/pdf":
		fileExtension = ".pdf"
	case "audio/mpeg":
		fileExtension = ".mp3"
	case "audio/wav":
		fileExtension = ".wav"
	case "video/mp4":
		fileExtension = ".mp4"
	case "application/octet-stream":
		fileExtension = ".tiff"
	default:
		contentType = "application/json"
		fileExtension = ".png"
	}

	// Generate filename if not provided
	if fileName == "" {
		fileName = fmt.Sprintf("file_%d%s", time.Now().Unix(), fileExtension)
	} else {
		// Ensure the filename has the correct extension
		if !strings.HasSuffix(strings.ToLower(fileName), fileExtension) {
			fileName += fileExtension
		}
	}

	// Sanitize and make filename unique
	cleanName := generateUniqueFilename(fileName)
	objectName := fmt.Sprintf("%s/%s", folder, cleanName)

	_, err = m.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(bucketName),
		Key:                aws.String(objectName),
		Body:               bytes.NewReader(fileBytes),
		ContentLength:      aws.Int64(int64(len(fileBytes))),
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String("inline"),
	})

	if err != nil {
		return newResponse(http.StatusInternalServerError, false, "Failed to upload file to S3", nil, err.Error())
	}

	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, m.region, objectName)
	return newResponse(http.StatusOK, true, "File uploaded successfully", gin.H{"url": fileURL}, nil)
}

func (m *MediaServiceAWS) UploadFile(bucketName string, file []byte, fileName, folder string) Response {
	cleanName := generateUniqueFilename(fileName)
	objectName := fmt.Sprintf("%s/%s", folder, cleanName)
	contentType := getContentType(fileName)

	_, err := m.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(bucketName),
		Key:                aws.String(objectName),
		Body:               bytes.NewReader(file),
		ContentLength:      aws.Int64(int64(len(file))),
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String("inline"),
	})

	if err != nil {
		return newResponse(http.StatusInternalServerError, false, "Failed to upload file to S3", nil, err.Error())
	}

	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, m.region, objectName)
	return newResponse(http.StatusOK, true, "File uploaded successfully", gin.H{"url": fileURL}, nil)
}

func (m *MediaServiceAWS) UploadFiles(bucketName string, files [][]byte, fileNames []string, folder string) Response {
	if len(files) > 10 {
		return newResponse(http.StatusBadRequest, false, "Maximum 10 files allowed", nil, nil)
	}

	if len(files) != len(fileNames) {
		return newResponse(http.StatusBadRequest, false, "Number of files and names must match", nil, nil)
	}

	var urls []string
	for i, file := range files {
		cleanName := generateUniqueFilename(fileNames[i])
		objectName := fmt.Sprintf("%s/%s", folder, cleanName)
		contentType := getContentType(fileNames[i])

		_, err := m.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(bucketName),
			Key:                aws.String(objectName),
			Body:               bytes.NewReader(file),
			ContentLength:      aws.Int64(int64(len(file))),
			ContentType:        aws.String(contentType),
			ContentDisposition: aws.String("inline"),
		})

		if err != nil {
			return newResponse(http.StatusInternalServerError, false, "Failed to upload files", nil, err.Error())
		}

		urls = append(urls, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, m.region, objectName))
	}

	return newResponse(http.StatusOK, true, "Files uploaded successfully", gin.H{"urls": urls}, nil)
}

func sanitizeFilename(filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	sanitized := reg.ReplaceAllString(name, "_")
	return strings.ToLower(sanitized) + ext
}

func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	sanitized := sanitizeFilename(name)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", sanitized, timestamp, ext)
}
