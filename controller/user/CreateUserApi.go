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
	CountryCode   string `json:"country_code" binding:"required"`
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
	ID           string              `json:"id"`
	Username     string              `json:"username"`
	MobileNumber uint64              `json:"mobile_number"`
	CountryCode  string              `json:"country_code"`
	IsValidated  bool                `json:"is_validated"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	OtpResponse  *AadhaarOTPResponse `json:"otp_response,omitempty"`
}

type CreateUserResponse struct {
	StatusCode    int          `json:"status_code"`
	Success       bool         `json:"success"`
	Message       string       `json:"message"`
	Data          *MinimalUser `json:"data"`
	DataTimeStamp string       `json:"data_time_stamp"`
}

func (s *Server) CreateUserRestApi(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Invalid request body",
			"data":        nil,
		})
		return
	}

	if !helper.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     fmt.Sprintf("username '%s' contains invalid characters. Only alphanumeric (a-z, A-Z, 0-9), /, _, |, -, =, + are allowed, and spaces are prohibited", req.Username),
			"data":        nil,
		})
		return
	}

	if err := s.UserRepo.CheckIfUserExists(c.Request.Context(), req.Username); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status_code": http.StatusConflict,
			"success":     false,
			"message":     "Username already exists",
			"data":        nil,
		})
		return
	}

	existingUser, _ := s.UserRepo.FindUserByMobile(c.Request.Context(), req.MobileNumber)
	if existingUser != nil && existingUser.MobileNumber == req.MobileNumber {
		c.JSON(http.StatusConflict, gin.H{
			"status_code": http.StatusConflict,
			"success":     false,
			"message":     "Mobile number already in use",
			"data":        nil,
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Password is required",
			"data":        nil,
		})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Username is required",
			"data":        nil,
		})
		return
	}

	if req.MobileNumber == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Mobile number is required",
			"data":        nil,
		})
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Failed to hash password",
			"data":        nil,
		})
		return
	}

	newUser := &model.User{
		Username:     req.Username,
		Password:     hashedPassword,
		IsValidated:  false,
		MobileNumber: req.MobileNumber,
		CountryCode:  &req.CountryCode,
	}

	var otpResponse *AadhaarOTPResponse
	if req.AadhaarNumber != "" {
		existingUser, _ := s.UserRepo.FindUserByAadhaar(c.Request.Context(), req.AadhaarNumber)
		if existingUser != nil && existingUser.AadhaarNumber != nil && *existingUser.AadhaarNumber == req.AadhaarNumber {
			c.JSON(http.StatusConflict, gin.H{
				"status_code": http.StatusConflict,
				"success":     false,
				"message":     "Aadhaar number already exists",
				"data":        nil,
			})
			return
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
						otpResponse = &AadhaarOTPResponse{
							Timestamp:     aadhaarResp.Response.Timestamp,
							TransactionID: aadhaarResp.Response.TransactionID,
							Entity:        aadhaarResp.Response.Data.Entity,
							OtpMessage:    aadhaarResp.Response.Data.Message,
							ReferenceID:   strconv.FormatInt(aadhaarResp.Response.Data.ReferenceID, 10),
							StatusCode:    int32(aadhaarResp.Response.Code),
						}
					}
				}
			}
		}
	}

	createdUser, err := s.UserRepo.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to create user",
			"data":        nil,
		})
		return
	}

	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		createdUser.ID,
		createdUser.Username,
		createdUser.IsValidated,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to set auth headers",
			"data":        nil,
		})
		return
	}

	minimalUser := &MinimalUser{
		ID:           createdUser.ID,
		Username:     createdUser.Username,
		MobileNumber: createdUser.MobileNumber,
		CountryCode:  *createdUser.CountryCode,
		IsValidated:  createdUser.IsValidated,
		CreatedAt:    createdUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    createdUser.UpdatedAt.Format(time.RFC3339Nano),
		OtpResponse:  otpResponse,
	}

	response := &CreateUserResponse{
		StatusCode:    http.StatusCreated,
		Success:       true,
		Message:       "User created successfully",
		Data:          minimalUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}
	if otpResponse != nil {
		response.Message = "User created and OTP sent successfully for Aadhaar verification"
	}

	c.JSON(http.StatusCreated, response)
}
