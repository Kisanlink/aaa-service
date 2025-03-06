package grpc_server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Kisanlink/aaa-service/controller/permissions"
	rolepermission "github.com/Kisanlink/aaa-service/controller/role_Permission"
	"github.com/Kisanlink/aaa-service/controller/roles"
	"github.com/Kisanlink/aaa-service/controller/user"

	// "github.com/Kisanlink/aaa-service/middleware"
	pb "github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
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
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	s :=  grpc.NewServer()
	// grpc.UnaryInterceptor(middleware.AuthInterceptor(db))
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	connectRolePermissionRepo := repositories.NewRolePermissionRepository(db)
	pb.RegisterGreeterServer(s, &GreeterServer{})
	userServer := user.NewUserServer(userRepo)
	pb.RegisterUserServiceServer(s, userServer)
	roleServer := roles.NewRoleServer(roleRepo,permissionRepo)
	pb.RegisterRoleServiceServer(s, roleServer)
	permissionServer := permissions.NewPermissionServer(permissionRepo,roleRepo)
	pb.RegisterPermissionServiceServer(s, permissionServer)
	connectRolePermissionServer := rolepermission.NewConnectRolePermissionServer(connectRolePermissionRepo, roleRepo, permissionRepo)
	pb.RegisterConnectRolePermissionServiceServer(s, connectRolePermissionServer)
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(),
	)
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
