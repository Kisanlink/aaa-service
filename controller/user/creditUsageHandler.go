package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type CreditUsageRequest struct {
	Transaction string `json:"transaction"` // "debit", "credit", or empty to just fetch credits
	Username    string `json:"username" binding:"required"`
	Credits     int    `json:"credits"` // Required only for credit/debit transactions
}

type CreditUsageResponse struct {
	StatusCode       int    `json:"status_code"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	RemainingCredits int    `json:"remaining_credits"`
	DataTimeStamp    string `json:"data_time_stamp"`
}

func (s *Server) CreditUsageHandler(c *gin.Context) {
	var req CreditUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// If transaction field is empty, just fetch the credits
	if req.Transaction == "" {
		credits, err := s.UserRepo.GetCreditsByUsername(c.Request.Context(), req.Username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		response := CreditUsageResponse{
			StatusCode:       http.StatusOK,
			Success:          true,
			Message:          "Fetched user credits successfully",
			RemainingCredits: credits,
			DataTimeStamp:    time.Now().Format(time.RFC3339),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Validate that the credits value is positive for credit or debit transactions
	if req.Credits <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credits must be greater than zero"})
		return
	}

	var user *model.User
	var err error

	switch req.Transaction {
	case "debit":
		user, err = s.UserRepo.DebitUser(c.Request.Context(), req.Username, req.Credits)
		if err != nil {
			if err.Error() == "Insufficient credits" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient credits"})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			}
			return
		}
	case "credit":
		user, err = s.UserRepo.CreditUser(c.Request.Context(), req.Username, req.Credits)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction type. Use 'debit', 'credit', or leave empty to fetch credits"})
		return
	}

	response := CreditUsageResponse{
		StatusCode:       http.StatusOK,
		Success:          true,
		Message:          "Transaction successful",
		RemainingCredits: user.Credits,
		DataTimeStamp:    time.Now().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}
