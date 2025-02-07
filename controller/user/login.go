package user

// import (
// 	"context"
// 	"log"

// 	"golang.org/x/crypto/bcrypt"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	"github.com/Kisanlink/aaa-service/database"
// 	"github.com/Kisanlink/aaa-service/helper"
// 	"github.com/Kisanlink/aaa-service/model"
// 	"github.com/Kisanlink/aaa-service/pb"
// )

// type AuthServer struct {
//     pb.UnimplementedAuthServiceServer
// }

// func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
//     // Validate input
//     if req.Username == "" || req.Password == "" {
//         return nil, status.Error(codes.InvalidArgument, "Username and password are required")
//     }

//     // Fetch user from the database
//     var user model.User
//     if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
//         log.Printf("User not found: %v", err)
//         return nil, status.Error(codes.NotFound, "Account doesn't exist")
//     }

//     // Compare password hash
//     err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
//     if err != nil {
//         log.Printf("Invalid password: %v", err)
//         return nil, status.Error(codes.InvalidArgument, "Invalid username or password")
//     }

//     // Generate token
//     token, err := helper.GenerateToken(user.ID.String(), user.Username, string(user.Role))
//     if err != nil {
//         log.Printf("Failed to generate token: %v", err)
//         return nil, status.Error(codes.Internal, "Could not generate token")
//     }

//     // Prepare response
//     response := &pb.LoginResponse{
//         StatusCode:   200,
//         Message:      "Login successful",
//         Token:        token,
//         RefreshToken: "", // Add refresh token logic if needed
//         User: &pb.User{
//             Id:       user.ID.String(),
//             Username: user.Username,
//         },
//     }

//     return response, nil
// }
