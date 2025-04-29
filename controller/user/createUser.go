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
	UserRepo     *repositories.UserRepository
	RoleRepo     *repositories.RoleRepository
	PermRepo     *repositories.PermissionRepository
	RolePermRepo *repositories.RolePermissionRepository
}

func NewUserServer(userRepo *repositories.UserRepository, roleRepo *repositories.RoleRepository, permRepo *repositories.PermissionRepository, rolePermRepo *repositories.RolePermissionRepository) *Server {
	return &Server{
		UserRepo:     userRepo,
		RoleRepo:     roleRepo,
		PermRepo:     permRepo,
		RolePermRepo: rolePermRepo,
	}
}

func (s *Server) RegisterUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if !helper.IsValidUsername(req.Username) {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"username '%s' contains invalid characters. Only alphanumeric (a-z, A-Z, 0-9), /, _, |, -, =, + are allowed, and spaces are prohibited",
			req.Username,
		)
	}
	if err := s.UserRepo.CheckIfUserExists(ctx, req.Username); err != nil {
		return nil, err
	}

	existingUser, _ := s.UserRepo.FindUserByMobile(ctx, req.MobileNumber)
	if existingUser != nil && existingUser.MobileNumber == req.MobileNumber {
		return nil, status.Error(codes.AlreadyExists, "Mobile number already in use")
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
	// Convert to string for validation
	mobileStr := strconv.FormatUint(req.MobileNumber, 10)

	// Check length exactly 10 digits
	if len(mobileStr) != 10 {
		return nil, status.Error(codes.InvalidArgument, "mobile number must be exactly 10 digits")
	}

	// Check for repeated digit numbers (like 1111111111, 2222222222, etc.)
	if isRepeatedDigitNumber(mobileStr) {
		return nil, status.Error(codes.InvalidArgument, "mobile number cannot be a repeated digit number")
	}
	// Check for sequential numbers (like 1234567890)
	if isSequentialNumber(mobileStr) {
		return nil, status.Error(codes.InvalidArgument, "mobile number cannot be sequential")
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
	}

	var otpResponse *pb.AadhaarOTPResponse
	if req.AadhaarNumber != "" {
		existingUser, _ := s.UserRepo.FindUserByAadhaar(ctx, req.AadhaarNumber)
		if existingUser != nil && existingUser.AadhaarNumber != nil && *existingUser.AadhaarNumber == req.AadhaarNumber {
			return nil, status.Error(codes.AlreadyExists, "Aadhaar number already exists")
		}
		newUser.AadhaarNumber = &req.AadhaarNumber

		client := &http.Client{}
		otpReq := struct {
			AadhaarNumber string `json:"aadhaar_number"`
		}{
			AadhaarNumber: req.AadhaarNumber,
		}

		jsonData, err := json.Marshal(otpReq)
		if err != nil {
			log.Printf("Failed to marshal Aadhaar OTP request: %v", err)
		} else {
			url := fmt.Sprintf("%s/api/v1/aadhaar/otp", os.Getenv("AADHAAR_VERIFICATION_ENDPOINT"))
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Failed to send Aadhaar OTP request: %v", err)
			} else {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Failed to read OTP response: %v", err)
				} else {
					// fmt.Printf("Original Aadhaar OTP Response: %s\n", string(body))
					var aadhaarResp struct {
						Message  string `json:"message"`
						Response struct {
							Timestamp     int64  `json:"timestamp"`
							TransactionID string `json:"transaction_id"`
							Data          struct {
								Entity      string `json:"@entity"`
								Message     string `json:"message"`
								ReferenceID int64  `json:"reference_id"`
							} `json:"data"`
							Code int `json:"code"`
						} `json:"response"`
					}

					if err := json.Unmarshal(body, &aadhaarResp); err != nil {
						log.Printf("Failed to parse OTP response: %v", err)
					} else {
						otpResponse = &pb.AadhaarOTPResponse{
							Timestamp:     aadhaarResp.Response.Timestamp,
							TransactionId: aadhaarResp.Response.TransactionID,
							Entity:        aadhaarResp.Response.Data.Entity,
							OtpMessage:    aadhaarResp.Response.Data.Message,
							ReferenceId:   strconv.FormatInt(aadhaarResp.Response.Data.ReferenceID, 10),
							StatusCode:    int32(aadhaarResp.Response.Code),
						}
					}
				}
			}
		}
	}

	createdUser, err := s.UserRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	if err := helper.SetAuthHeadersWithTokens(
		ctx,
		createdUser.ID,
		createdUser.Username,
		createdUser.IsValidated,
	); err != nil {
		return nil, err
	}

	minimalUser := &pb.MinimalUser{
		Id:           createdUser.ID,
		Username:     createdUser.Username,
		Password:     "",
		MobileNumber: createdUser.MobileNumber,
		CountryCode:  *createdUser.CountryCode,
		IsValidated:  createdUser.IsValidated,
		CreatedAt:    createdUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    createdUser.UpdatedAt.Format(time.RFC3339Nano),
		OtpResponse:  otpResponse,
	}

	response := &pb.CreateUserResponse{
		StatusCode:    http.StatusCreated,
		Success:       true,
		Message:       "User created successfully",
		Data:          minimalUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}
	if otpResponse != nil {
		response.Message = "User created and OTP sent successfully for Aadhaar verification"
	}

	return response, nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func isRepeatedDigitNumber(num string) bool {
	if len(num) == 0 {
		return false
	}
	firstDigit := num[0]
	for i := 1; i < len(num); i++ {
		if num[i] != firstDigit {
			return false
		}
	}
	return true
}
func isSequentialNumber(num string) bool {
	if len(num) <= 1 {
		return false
	}

	// Convert each character to its actual digit value
	digits := make([]int, len(num))
	for i, c := range num {
		digits[i] = int(c - '0')
	}

	ascending := true
	descending := true

	for i := 1; i < len(digits); i++ {
		// Check ascending sequence (1,2,3...)
		if digits[i] != (digits[i-1]+1)%10 { // Using %10 to handle 9->0 case
			ascending = false
		}
		// Check descending sequence (9,8,7...)
		if digits[i] != (digits[i-1]-1+10)%10 { // Using +10 to handle 0->9 case
			descending = false
		}
		// Early exit if both fail
		if !ascending && !descending {
			return false
		}
	}

	return ascending || descending
}
