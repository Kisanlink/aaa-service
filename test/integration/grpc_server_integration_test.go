//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCServer_Connection(t *testing.T) {
	addr := os.Getenv("INTEGRATION_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Test that connection is established
	if conn.GetState().String() != "READY" {
		t.Errorf("Expected connection state READY, got %s", conn.GetState().String())
	}
}

// Uncomment and implement when you have gRPC services
// func TestGRPCServer_HealthService(t *testing.T) {
//     addr := os.Getenv("INTEGRATION_GRPC_ADDR")
//     if addr == "" {
//         addr = "localhost:50052"
//     }
//
//     conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
//     if err != nil {
//         t.Fatalf("Failed to connect to gRPC server: %v", err)
//     }
//     defer conn.Close()
//
//     // client := pb.NewHealthServiceClient(conn)
//     // resp, err := client.CheckDatabaseHealth(context.Background(), &pb.HealthRequest{})
//     // if err != nil {
//     //     t.Fatalf("Health check failed: %v", err)
//     // }
//     // if resp.Status != pb.HealthStatus_HEALTHY {
//     //     t.Errorf("Expected HEALTHY, got %v", resp.Status)
//     // }
// }
