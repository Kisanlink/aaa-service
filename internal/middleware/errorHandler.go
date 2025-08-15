package middleware

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		err := c.Errors[0].Err
		switch e := err.(type) {
		case *helper.CustomError:
			// If it's a CustomError, include both message and original error if available
			if e.Err != nil {
				c.JSON(e.Code, gin.H{
					"error": fmt.Sprintf("%s - %v", e.Message, e.Err),
				})
			} else {
				c.JSON(e.Code, gin.H{
					"error": e.Message,
				})
			}
		default:
			// Default case if it's not a CustomError, return detailed error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Internal Server Error: %v", e),
			})
		}
	}
}
