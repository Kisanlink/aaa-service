package helper

import (
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("askdwkdfmlermferflersmflesrmflersmflesrmflkes") // Replace with a secure key in production

// GenerateAccessToken generates a short-lived JWT access token
func GenerateAccessToken(userID string, userRoleId []model.UserRole, username string, isvalidate bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		"roleIds":    userRoleId,
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
func GenerateRefreshToken(userID string, userRoleId []model.UserRole, username string, isvalidate bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"exp":        time.Now().Add(time.Hour * 24 * 7).Unix(), // Expires in 7 days
		"roleIds":    userRoleId,
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
