package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	UserRepo *repositories.UserRepository
	RoleRepo *repositories.RoleRepository
	PermRepo *repositories.PermissionRepository
	RolePermRepo *repositories.RolePermissionRepository
}

func NewUserServer(userRepo *repositories.UserRepository,roleRepo *repositories.RoleRepository,permRepo *repositories.PermissionRepository,rolePermRepo *repositories.RolePermissionRepository) *Server {
	return &Server{
		UserRepo: userRepo,
		RoleRepo: roleRepo,
		PermRepo: permRepo,
		RolePermRepo: rolePermRepo,
	}
}

func (s *Server) RegisterUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	   if req.AadhaarNumber != "" {
		client := &http.Client{}
        otpReq := struct {
            AadhaarNumber string `json:"aadhaar_number"`
        }{
            AadhaarNumber: req.AadhaarNumber,
        }
        
        jsonData, err := json.Marshal(otpReq)
        if err != nil {
            return nil, status.Error(codes.Internal, "Failed to marshal Aadhaar OTP request")
        }
		url := fmt.Sprintf("%s/api/v1/aadhaar/otp", os.Getenv("AADHAAR_VERIFICATION_ENDPOINT"))
        resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            return nil, status.Error(codes.Internal, "Failed to send Aadhaar OTP request")
        }
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, status.Error(codes.Internal, "Failed to read OTP response")
        }
		fmt.Printf("Original Aadhaar OTP Response: %s\n", string(body))
		var otpResponse struct {
			Message  string `json:"message"`
			Response struct {
				Timestamp     int64  `json:"timestamp"`
				TransactionID string `json:"transaction_id"`
				Data         struct {
					Entity      string `json:"@entity"`
					Message    string `json:"message"`
					ReferenceID int64  `json:"reference_id"`
				} `json:"data"`
				Code int `json:"code"`
			} `json:"response"`
		}
        
        if err := json.Unmarshal(body, &otpResponse); err != nil {
            return nil, status.Error(codes.Internal, "Failed to parse OTP response")
        }
        
        if resp.StatusCode != http.StatusOK {
            return nil, status.Error(codes.Internal, "Failed to generate Aadhaar OTP")
        }
			
			hashedPassword, err := HashPassword(req.Password)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "Failed to hash password")				
			}
			
			newUser := &model.User{
				Username:     req.Username,
				Password:     hashedPassword,
				IsValidated:  false,
				MobileNumber: req.MobileNumber,
				CountryCode:  &req.CountryCode,
				AadhaarNumber: &req.AadhaarNumber,
			}
			
			createdUser, err := s.UserRepo.CreateUser(ctx, newUser)
			if err != nil {
				log.Printf("Failed to create user in background: %v", err)
			}
		if err := helper.SetAuthHeadersWithTokens(
			ctx,
			createdUser.ID,
			createdUser.Username,
			createdUser.IsValidated,
		); err != nil {
			return nil, err
		}
        // Return the full OTP response details along with success message
        return &pb.CreateUserResponse{
            StatusCode: int32(codes.OK),
            Message:    "OTP sent successfully for Aadhaar verification",
			User: &pb.MinimalUser{
				Id:          createdUser.ID,
				Username:    createdUser.Username,
				Password:    "", // Empty for security
				MobileNumber:createdUser.MobileNumber,
				CountryCode:*createdUser.CountryCode,
				IsValidated: createdUser.IsValidated,
				CreatedAt:   createdUser.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   createdUser.UpdatedAt.Format(time.RFC3339Nano),
			},
            Response: &pb.AadhaarOTPResponse{
                Timestamp:     otpResponse.Response.Timestamp,
                TransactionId: otpResponse.Response.TransactionID,
                Entity:       otpResponse.Response.Data.Entity,
                OtpMessage:    otpResponse.Response.Data.Message,
				ReferenceId:   strconv.FormatInt(otpResponse.Response.Data.ReferenceID, 10), // Convert to string
                StatusCode:    int32(otpResponse.Response.Code),
            },
        }, nil
    }

    // Normal user creation flow
    if err := s.UserRepo.CheckIfUserExists(ctx, req.Username); err != nil {
        return nil, err
    }
    
    if req.Password == "" {
        return nil, status.Error(codes.NotFound, "Password is required")
    }
    
    if req.Username == "" {
        return nil, status.Error(codes.NotFound, "username is required")
    }
	if req.MobileNumber == 0 { 
		return nil, status.Error(codes.NotFound, "mobile number is required") 
	}	
    
    hashedPassword, err := HashPassword(req.Password)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Failed to hash password")
    }
    
    newUser := &model.User{
        Username:     req.Username,
        Password:     hashedPassword,
        IsValidated:  false,
        MobileNumber: req.MobileNumber, // uint64 field
        CountryCode:  &req.CountryCode,
    }

    createdUser, err := s.UserRepo.CreateUser(ctx, newUser)
    if err != nil {
        return nil, err
    }    
    // roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, createdUser.ID)
    // if err != nil {
    //     return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
    // }
    
    // userRole := &pb.UserRoleResponse{
    //     Roles:       roles,
    //     Permissions: permissions,
    //     Actions:     actions,
    // }
    
    // Create response with headers (note: gRPC uses metadata for headers)
	if err := helper.SetAuthHeadersWithTokens(
		ctx,
		createdUser.ID,
		createdUser.Username,
		createdUser.IsValidated,
	); err != nil {
		return nil, err
	}
    
	   minimalUser := &pb.MinimalUser{
        Id:         createdUser.ID,
        Username:   createdUser.Username,
        Password:   "", // Empty for security
		MobileNumber:createdUser.MobileNumber,
		CountryCode:*createdUser.CountryCode,
        IsValidated: createdUser.IsValidated,
        CreatedAt:  createdUser.CreatedAt.Format(time.RFC3339Nano),
        UpdatedAt:  createdUser.UpdatedAt.Format(time.RFC3339Nano),
    }

    return &pb.CreateUserResponse{
        StatusCode: int32(codes.OK),
        Message:    "User created successfully",
        User:       minimalUser,
        Response:   nil,
    }, nil
}



func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func ConvertToPBUserRoles(userRoles []model.UserRole) []*pb.UserRole {
	var pbUserRoles []*pb.UserRole
	for _, userRole := range userRoles {
		pbUserRole := &pb.UserRole{
			Id:               userRole.ID,
			UserId:           userRole.UserID,
			RolePermissionId: userRole.RolePermissionID,
			CreatedAt:        userRole.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        userRole.UpdatedAt.Format(time.RFC3339Nano),
		}
		pbUserRoles = append(pbUserRoles, pbUserRole)
	}
	return pbUserRoles
}
func LowerCaseSlice(input []string) []string {
    for i, val := range input {
        input[i] = strings.ToLower(val)
    }
    return input
}