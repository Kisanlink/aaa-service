package user

import (
	"bytes"
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
	"github.com/gin-gonic/gin"
)



type CreateUserRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required"`
	MobileNumber  uint64 `json:"mobile_number" binding:"required"`
	CountryCode   string `json:"country_code"`
	AadhaarNumber string `json:"aadhaar_number"`
}

type AadhaarOTPResponse struct {
	Timestamp     int64  `json:"timestamp"`
	TransactionID string `json:"transaction_id"`
	Entity        string `json:"entity"`
	OtpMessage    string `json:"otp_message"`
	ReferenceID   string `json:"reference_id"`
	StatusCode    int32  `json:"status_code"`
}

type MinimalUser struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	MobileNumber uint64 `json:"mobile_number"`
	CountryCode  string `json:"country_code"`
	IsValidated  bool   `json:"is_validated"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateUserResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string            `json:"message"`
	User       *MinimalUser      `json:"user,omitempty"`
	Response   *AadhaarOTPResponse `json:"response,omitempty"`
}



func (s *Server) CreateUserRestApi(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Aadhaar verification flow
	if req.AadhaarNumber != "" {
		client := &http.Client{}
		otpReq := struct {
			AadhaarNumber string `json:"aadhaar_number"`
		}{
			AadhaarNumber: req.AadhaarNumber,
		}

		jsonData, err := json.Marshal(otpReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal Aadhaar OTP request"})
			return
		}

		url := fmt.Sprintf("%s/api/v1/aadhaar/otp", os.Getenv("AADHAAR_VERIFICATION_ENDPOINT"))
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send Aadhaar OTP request"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read OTP response"})
			return
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OTP response"})
			return
		}

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Aadhaar OTP"})
			return
		}

		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
			return
		}

		newUser := &model.User{
			Username:      req.Username,
			Password:      hashedPassword,
			IsValidated:   false,
			MobileNumber:  req.MobileNumber,
			CountryCode:   &req.CountryCode,
			AadhaarNumber: &req.AadhaarNumber,
		}

		createdUser, err := s.UserRepo.CreateUser(c.Request.Context(), newUser)
		if err != nil {
			log.Printf("Failed to create user in background: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Set auth headers with tokens
		if err := helper.SetAuthHeadersWithTokens(
			c.Request.Context(),
			createdUser.ID,
			createdUser.Username,
			createdUser.IsValidated,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set auth headers"})
			return
		}

		// Return the full OTP response details along with success message
		response := CreateUserResponse{
			StatusCode:http.StatusCreated,
			Success: true,
			Message:    "OTP sent successfully for Aadhaar verification",
			User: &MinimalUser{
				ID:           createdUser.ID,
				Username:     createdUser.Username,
				MobileNumber: createdUser.MobileNumber,
				CountryCode:  *createdUser.CountryCode,
				IsValidated:  createdUser.IsValidated,
				CreatedAt:    createdUser.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:    createdUser.UpdatedAt.Format(time.RFC3339Nano),
			},
			Response: &AadhaarOTPResponse{
				Timestamp:     otpResponse.Response.Timestamp,
				TransactionID: otpResponse.Response.TransactionID,
				Entity:        otpResponse.Response.Data.Entity,
				OtpMessage:    otpResponse.Response.Data.Message,
				ReferenceID:   strconv.FormatInt(otpResponse.Response.Data.ReferenceID, 10),
				StatusCode:    int32(otpResponse.Response.Code),
			},
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Normal user creation flow
	if err := s.UserRepo.CheckIfUserExists(c.Request.Context(), req.Username); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if req.MobileNumber == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := &model.User{
		Username:     req.Username,
		Password:     hashedPassword,
		IsValidated:  false,
		MobileNumber: req.MobileNumber,
		CountryCode:  &req.CountryCode,
	}

	createdUser, err := s.UserRepo.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Set auth headers with tokens
	if err := helper.SetAuthHeadersWithTokens(
		c.Request.Context(),
		createdUser.ID,
		createdUser.Username,
		createdUser.IsValidated,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set auth headers"})
		return
	}

	response := CreateUserResponse{
		StatusCode:http.StatusCreated,
		Message:    "User created successfully",
		User: &MinimalUser{
			ID:           createdUser.ID,
			Username:     createdUser.Username,
			MobileNumber: createdUser.MobileNumber,
			CountryCode:  *createdUser.CountryCode,
			IsValidated:  createdUser.IsValidated,
			CreatedAt:    createdUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    createdUser.UpdatedAt.Format(time.RFC3339Nano),
		},
	}

	c.JSON(http.StatusOK, response)
}

