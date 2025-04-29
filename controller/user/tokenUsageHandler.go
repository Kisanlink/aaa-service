package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type CreditUsageRequest struct {
	UserID          string  `json:"user_id" binding:"required"` // Using Base.ID
	TransactionType *string `json:"transaction_type"`           // "debit", "credit", or nil
	Tokens          *int    `json:"tokens"`                     // Required for transactions
}

func (s *Server) TokenUsageHandler(c *gin.Context) {
	ctx := c.Request.Context() // <- context here

	var req CreditUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request body",
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	if req.TransactionType == nil {
		tokens, err := s.UserRepo.GetTokensByUserID(ctx, req.UserID) // <- pass ctx
		if err != nil {
			c.JSON(http.StatusNotFound, model.Response{
				StatusCode: http.StatusNotFound,
				Success:    false,
				Message:    "User not found",
				Error:      err.Error(),
				TimeStamp:  time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusOK, model.Response{
			StatusCode: http.StatusOK,
			Success:    true,
			Message:    "Fetched user tokens successfully",
			Data: gin.H{
				"remaining_tokens": tokens,
			},
			TimeStamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	if req.Tokens == nil || *req.Tokens <= 0 {
		c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Tokens must be greater than zero for transactions",
			Error:      "invalid tokens value",
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	var user *model.User
	var err error

	switch *req.TransactionType {
	case "debit":
		user, err = s.UserRepo.DebitUserByID(ctx, req.UserID, *req.Tokens)
		if err != nil {
			status := http.StatusBadRequest
			msg := "Insufficient tokens"
			if err.Error() != "insufficient tokens" {
				status = http.StatusNotFound
				msg = "User not found"
			}
			c.JSON(status, model.Response{
				StatusCode: status,
				Success:    false,
				Message:    msg,
				Error:      err.Error(),
				TimeStamp:  time.Now().Format(time.RFC3339),
			})
			return
		}
	case "credit":
		user, err = s.UserRepo.CreditUserByID(ctx, req.UserID, *req.Tokens)
		if err != nil {
			c.JSON(http.StatusNotFound, model.Response{
				StatusCode: http.StatusNotFound,
				Success:    false,
				Message:    "User not found",
				Error:      err.Error(),
				TimeStamp:  time.Now().Format(time.RFC3339),
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid transaction type. Use 'debit', 'credit', or leave empty",
			Error:      "invalid transaction_type",
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Transaction successful",
		Data: gin.H{
			"remaining_tokens": user.Tokens,
		},
		TimeStamp: time.Now().Format(time.RFC3339),
	})
}
