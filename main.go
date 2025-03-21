package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kisanlink/aaa-service/controller/user"
	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/grpc_server"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, skipping...")
	}
}

func main() {
	database.ConnectDB()

	// Start the gRPC server
	grpcServer, err := grpc_server.StartGRPCServer(database.DB)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

	log.Println("Application started successfully")
	// start gin server
	router := gin.Default()
	userRepo := repositories.NewUserRepository(database.DB)
	s := user.Server{UserRepo: userRepo} 
	router.POST("/api/v1/user/login", s.LoginRestApi)
	go func() {
		if err := router.Run(":3000"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start Gin server: %v", err)
		}
	}()

	log.Println("Gin server running on port 8080")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped. Exiting application.")

}


