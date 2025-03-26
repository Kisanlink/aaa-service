package helper

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var jwtKey = []byte("askdwkdfmlermferflersmflesrmflersmflesrmflkes") // Replace with a secure key in production

// GenerateAccessToken generates a short-lived JWT access token
func GenerateAccessToken(userID string, username string, isvalidate bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		"username":   username,
		"isvalidate": isvalidate,
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateRefreshToken generates a long-lived JWT refresh token
func GenerateRefreshToken(userID string, username string, isvalidate bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"exp":        time.Now().Add(time.Hour * 24 * 7).Unix(), // Expires in 7 days
		"username":   username,
		"isvalidate": isvalidate,
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateToken validates the JWT token and returns the user ID
func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", err
	}

	return userID, nil
}

func SetAuthHeadersWithTokens(ctx context.Context, userID, username string, isValidated bool) error {
	// Generate tokens
	accessToken, err := GenerateAccessToken(userID, username, isValidated)
	if err != nil {
		return status.Error(codes.Internal, "Failed to generate access token")
	}

	refreshToken, err := GenerateRefreshToken(userID, username, isValidated)
	if err != nil {
		return status.Error(codes.Internal, "Failed to generate refresh token")
	}

	// Set headers
	header := metadata.Pairs(
		"token", accessToken,
		"refreshtoken", refreshToken,
		"userid", userID,
	)
	
	if err := grpc.SendHeader(ctx, header); err != nil {
		return status.Errorf(codes.Internal, "unable to send headers: %v", err)
	}
	return nil
}


func SetAuthHeadersWithTokensRest(c *gin.Context, userID, username string, isValidated bool) error {
	accessToken, err := GenerateAccessToken(userID, username, isValidated)
	if err != nil {
		return err
	}

	refreshToken, err := GenerateRefreshToken(userID, username, isValidated)
	if err != nil {
		return err
	}
	c.Header("token", accessToken)
	c.Header("refreshtoken", refreshToken)
	c.Header("userid", userID)
	
	return nil
}