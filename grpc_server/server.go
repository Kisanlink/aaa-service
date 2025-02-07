package grpc_server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Kisanlink/aaa-service/controller/permissions"
	"github.com/Kisanlink/aaa-service/controller/roles"
	"github.com/Kisanlink/aaa-service/controller/user"
	pb "github.com/Kisanlink/aaa-service/pb"
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
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &GreeterServer{})
	userServer := &user.Server{DB: db}
	pb.RegisterUserServiceServer(s, userServer)
	roleServer := roles.NewRoleServer(db)
	pb.RegisterRoleServiceServer(s, roleServer)
	permissionServer := permissions.NewPermissionServer(db)
	pb.RegisterPermissionServiceServer(s, permissionServer)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s, nil
}
