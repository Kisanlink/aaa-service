package main

import (
	"log"
	"net"
	"net/http"

	"github.com/Kisanlink/aaa-service/grpc_server"
	pb "github.com/Kisanlink/aaa-service/pb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &grpc_server.GreeterServer{})
		log.Println("gRPC server listening on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start Gin HTTP server
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the Gin server!",
		})
	})

	log.Println("Gin server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run Gin server: %v", err)
	}
}
