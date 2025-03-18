package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/grpc_server"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped. Exiting application.")

}


