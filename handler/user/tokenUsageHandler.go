package user

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// TokenUsageHandler manages token transactions for users
// @Summary Manage user tokens
// @Description Handles token transactions (credit/debit) or fetches token balance when no transaction type is specified
// @Tags Users
// @Accept json
// @Produce json
// @Param request body model.CreditUsageRequest true "Token transaction request"
// @Success 200 {object} object "Returns remaining tokens in all cases" example({"remaining_tokens": 100})
// @Failure 400 {object} helper.ErrorResponse "Invalid request, insufficient tokens, or invalid transaction type"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Router /token-transaction [post]
func (s *UserHandler) TokenUsageHandler(c *gin.Context) {
	var req model.CreditUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body", err.Error()})
		return
	}

	if req.TransactionType == nil {
		tokens, err := s.userService.GetTokensByUserID(req.UserID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
			return
		}
		helper.SendSuccessResponse(c.Writer, http.StatusOK, "Fetched user tokens successfully", map[string]interface{}{
			"remaining_tokens": tokens,
		})
		return
	}

	if req.Tokens == nil || *req.Tokens <= 0 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Tokens must be greater than zero for transactions"})
		return
	}

	var user *model.User
	var err error

	switch *req.TransactionType {
	case "debit":
		user, err = s.userService.DebitUserByID(req.UserID, *req.Tokens)
		if err != nil {
			if err.Error() == "insufficient tokens" {
				helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Insufficient tokens"})
			} else {
				helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
			}
			return
		}
	case "credit":
		user, err = s.userService.CreditUserByID(req.UserID, *req.Tokens)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
			return
		}
	default:
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid transaction type. Use 'debit', 'credit', or leave empty"})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Transaction successful", map[string]interface{}{
		"remaining_tokens": user.Tokens,
	})
}
