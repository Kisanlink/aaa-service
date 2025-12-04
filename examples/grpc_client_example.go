package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Example demonstrating how to call the CatalogService.SeedRolesAndPermissions gRPC endpoint
func main() {
	// Connect to the gRPC server
	serverAddress := "localhost:50051" // Default gRPC port
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Create CatalogService client
	client := pb.NewCatalogServiceClient(conn)

	// Call SeedRolesAndPermissions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.SeedRolesAndPermissionsRequest{
		Force:          false, // Set to true to force re-seed
		OrganizationId: "",    // Empty for global roles
	}

	fmt.Println("Calling CatalogService.SeedRolesAndPermissions...")
	response, err := client.SeedRolesAndPermissions(ctx, req)
	if err != nil {
		log.Fatalf("Failed to seed roles and permissions: %v", err)
	}

	// Display results
	fmt.Printf("\n✅ Success!\n")
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("\nResults:\n")
	fmt.Printf("  - Actions Created: %d\n", response.ActionsCreated)
	fmt.Printf("  - Resources Created: %d\n", response.ResourcesCreated)
	fmt.Printf("  - Permissions Created: %d\n", response.PermissionsCreated)
	fmt.Printf("  - Roles Created: %d\n", response.RolesCreated)
	fmt.Printf("\nRoles:\n")
	for _, role := range response.CreatedRoles {
		fmt.Printf("  - %s\n", role)
	}
}
