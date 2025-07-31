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
	pb "github.com/Kisanlink/aaa-service/proto"
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

	// Register v1 services
	pb.RegisterGreeterServer(s, &GreeterServer{})
	userServer := &user.Server{DB: db}
	pb.RegisterUserServiceServer(s, userServer)
	roleServer := roles.NewRoleServer(db)
	pb.RegisterRoleServiceServer(s, roleServer)
	permissionServer := permissions.NewPermissionServer(db)
	pb.RegisterPermissionServiceServer(s, permissionServer)
	connectRolePermissionServer := rolepermission.NewConnectRolePermissionServer(db)
	pb.RegisterConnectRolePermissionServiceServer(s, connectRolePermissionServer)

	// Register v2 services
	userServerV2 := user.NewUserServerV2(db)
	pb.RegisterUserServiceV2Server(s, userServerV2)

	// TODO: Add v2 role and permission services when implemented
	// roleServerV2 := roles.NewRoleServerV2(db)
	// pb.RegisterRoleServiceV2Server(s, roleServerV2)
	// permissionServerV2 := permissions.NewPermissionServerV2(db)
	// pb.RegisterPermissionServiceV2Server(s, permissionServerV2)

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
