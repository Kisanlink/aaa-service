package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Kisanlink/aaa-service/config"
	"github.com/Kisanlink/aaa-service/database"
	docs "github.com/Kisanlink/aaa-service/docs"
	"github.com/Kisanlink/aaa-service/grpc"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

// @title AAA-Service API
// @version 1.0
// @description Authentication, Authorization, and Accounting (AAA) service providing RBAC-based access control. Supports both gRPC and REST API interfaces for seamless integration with client applications. Offers comprehensive user management, role-based permission control, and session accounting capabilities for secure system access.
func main() {
	database.ConnectDB()
	corsSetup := config.LoadConfig()
	database.HandleMigrationCommands()
	helper.InitLogger()
	docs.SwaggerInfo.BasePath = "/api/v1"

	// Create router
	r := gin.New()

	// Add CORS middleware first, before any routes
	r.Use(cors.New(config.GetCorsConfig(corsSetup)))

	// Add error handling middleware to ensure CORS headers are present in error responses
	r.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			// Ensure CORS headers are set even for error responses
			c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
			c.Header("Access-Control-Allow-Credentials", "true")
		}
	})

	// Setup routes after middleware
	routes.SetupRouter(database.DB)

	// Swagger documentation
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: fmt.Sprintf("%s/doc/doc.json", os.Getenv("URL")),
			CustomOptions: scalar.CustomOptions{
				PageTitle: "AAA-SERVICE API",
			},
			DarkMode: true,
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// Start gRPC server
	grpcServer, err := grpc.StartGRPCServer(database.DB)
	if err != nil {
		helper.Log.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcServer.GracefulStop()

	helper.Log.Println("Server is running on port:", corsSetup.Port)
	if err := r.Run(":" + corsSetup.Port); err != nil {
		helper.Log.Fatalf("Error starting server: %v", err)
	}
}
