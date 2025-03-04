package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kisanlink/aaa-service/client"
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
	roles := []string{"admin", "editor", "user"}
	permissions := []string{
        "create_user",
        "edit_user",
        "delete_user",
        "view_reports",
        "publish_content",
        "edit_content",
    }
    // updated, err := client.UpdateSchema(roles,permissions)
	// if err != nil {
	// 	log.Fatalf("Error reading schema: %v", err)
	// }
	// log.Printf("Updated Response: %+v", updated)
    // updated, err := client.CreateUserRoleRelationship("Alfiya",roles,permissions)
	// if err != nil {
	// 	log.Fatalf("Error reading schema: %v", err)
	// }
	// log.Printf("Updated Response: %+v", updated)
    // updated, err := client.ReadRelationshipsByUserID("Alfiya")
	// if err != nil {
	// 	log.Fatalf("Error reading schema: %v", err)
	// }
	// log.Printf("Updated Response: %+v", updated)
	results, err := client.CheckUserPermissions("Alfiya", roles, permissions)
	if err != nil {
		log.Fatalf("Failed to check permissions: %v", err)
	}
	
	for permission, hasPermission := range results {
		log.Printf("User has permission %s: %v", permission, hasPermission)
	}
	
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
