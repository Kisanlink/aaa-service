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

// CreateUserRestApi creates a new user account
// @Summary Create a new user
// @Description Creates a new user account with the provided details. Optionally sends OTP for Aadhaar verification if Aadhaar number is provided.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body model.CreateUserRequest true "User creation request"
// @Success 201 {object} helper.Response{data=model.MinimalUser} "User created successfully"
// @Failure 400 {object} helper.Response "Invalid request body or validation failed"
// @Failure 409 {object} helper.Response "Username, mobile number or Aadhaar already exists"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /register [post]
func (s *UserHandler) CreateUserRestApi(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if !helper.IsValidUsername(req.Username) {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{
			fmt.Sprintf("username '%s' contains invalid characters. Only alphanumeric (a-z, A-Z, 0-9), /, _, |, -, =, + are allowed, and spaces are prohibited", req.Username),
		})
		return
	}

	if err := s.userService.CheckIfUserExists(req.Username); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Username already exists"})
		return
	}

	existingUser, _ := s.userService.FindUserByMobile(req.MobileNumber)
	if existingUser != nil && existingUser.MobileNumber == req.MobileNumber {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Mobile number already in use"})
		return
	}

	if req.Password == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Password is required"})
		return
	}

	if req.Username == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Username is required"})
		return
	}

	if req.MobileNumber == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Mobile number is required"})
		return
	}

	// Enhanced Mobile Number Validation
	mobileStr := strconv.FormatUint(req.MobileNumber, 10)

	// 1. Check for exactly 10 digits
	if len(mobileStr) != 10 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Mobile number must be exactly 10 digits"})
		return
	}

	// 2. Check if number starts with 0
	if mobileStr[0] == '0' {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Mobile number cannot start with 0"})
		return
	}

	// 3. Check for repeated digits (like 1111111111)
	if helper.IsRepeatedDigitNumber(mobileStr) {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Mobile number cannot be all repeated digits"})
		return
	}

	// 4. Check for sequential numbers (like 1234567890)
	if helper.IsSequentialNumber(mobileStr) {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Mobile number cannot be sequential"})
		return
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Failed to hash password"})
		return
	}

	newUser := &model.User{
		Username:     req.Username,
		Password:     hashedPassword,
		IsValidated:  false,
		MobileNumber: req.MobileNumber,
		CountryCode:  req.CountryCode,
	}

	var otpResponse *model.AadhaarOTPResponse
	if req.AadhaarNumber != nil && *req.AadhaarNumber != "" {
		existingUser, _ := s.userService.FindUserByAadhaar(*req.AadhaarNumber)
		if existingUser != nil && existingUser.AadhaarNumber != nil && *existingUser.AadhaarNumber == *req.AadhaarNumber {
			helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Aadhaar number already exists"})
			return
		}
		newUser.AadhaarNumber = req.AadhaarNumber

		client := &http.Client{}
		otpReq := struct {
			AadhaarNumber string `json:"aadhaar_number"`
		}{
			AadhaarNumber: *req.AadhaarNumber,
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
						otpResponse = &model.AadhaarOTPResponse{
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

	createdUser, err := s.userService.CreateUser(*newUser)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to create user"})
		return
	}

	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		createdUser.ID,
		createdUser.Username,
		createdUser.IsValidated,
	); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to set auth headers"})
		return
	}

	minimalUser := &model.MinimalUser{
		ID:           createdUser.ID,
		Username:     createdUser.Username,
		MobileNumber: createdUser.MobileNumber,
		CountryCode:  *createdUser.CountryCode,
		IsValidated:  createdUser.IsValidated,
		CreatedAt:    createdUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    createdUser.UpdatedAt.Format(time.RFC3339Nano),
		OtpResponse:  otpResponse,
	}

	message := "User created successfully"
	if otpResponse != nil {
		message = "User created and OTP sent successfully for Aadhaar verification"
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, message, minimalUser)
}
