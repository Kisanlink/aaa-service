package grpc_server

import (
	"context"
	"log"

	pb "github.com/Kisanlink/aaa-service/pb"
)

type GreeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received request from: %v", req.Name)
	return &pb.HelloResponse{Message: "Hello " + req.Name}, nil
}

// func StartGRPCServer() {
// 	lis, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}

// 	s := grpc.NewServer()

// 	pb.RegisterGreeterServer(s, &Server{})

// 	log.Println("gRPC server is running on port 50051")

// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
