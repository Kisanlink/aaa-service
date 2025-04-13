package user

import (
	"context"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CheckAadhaarExist(ctx context.Context, req *pb.CheckAadhaarExistRequest) (*pb.CheckAadhaarExistResponse, error) {
	if req.AadhaarNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "Aadhaar number is required")
	}

	existingUser, err := s.UserRepo.FindUserByAadhaar(ctx, req.AadhaarNumber)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check Aadhaar existence: "+err.Error())
	}

	response := &pb.CheckAadhaarExistResponse{
		StatusCode: int32(codes.OK),
		Message:    "Aadhaar check completed successfully",
		IsExist:    existingUser != nil,
	}

	return response, nil
}
