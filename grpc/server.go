package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Kisanlink/aaa-service/grpc_handler/user"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/Kisanlink/aaa-service/services"
	pb "github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type GreeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received request from: %v", req.Name)
	return &pb.HelloResponse{Message: "Hello " + req.Name}, nil
}

func StartGRPCServer(db *gorm.DB) (*grpc.Server, error) {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
	// grpc.UnaryInterceptor(middleware.AuthInterceptor()),
	)
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	userService := services.NewUserService(userRepo, roleRepo)
	roleService := services.NewRoleService(roleRepo)

	// pb.RegisterGreeterServer(s, &GreeterServer{})
	userServer := user.NewUserServer(userService, roleService)
	pb.RegisterUserServiceServer(s, userServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Printf("GRPC Server is running on port %s", port)

	return s, nil
}
