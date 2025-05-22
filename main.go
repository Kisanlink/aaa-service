package main

import (
	"log"

	"github.com/Kisanlink/aaa-service/config"
	"github.com/Kisanlink/aaa-service/database"
	docs "github.com/Kisanlink/aaa-service/docs"
	"github.com/Kisanlink/aaa-service/grpc"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/gin-contrib/cors"
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
	r := routes.SetupRouter(database.DB)
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Use(cors.New(config.GetCorsConfig(corsSetup)))
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
