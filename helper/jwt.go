package helper

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/config"
	"github.com/Kisanlink/aaa-service/internal/entities/models"
	jwt "github.com/golang-jwt/jwt/v4"
)

// GenerateAccessToken generates a JWT access token using centralized config
func GenerateAccessToken(userID string, userRoleId []models.UserRole, username string, isvalidate bool) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	now := time.Now()
	iat := now.Add(-cfg.Leeway / 2)
	nbf := now.Add(-cfg.Leeway / 2)
	exp := now.Add(cfg.TTL)

	claims := jwt.MapClaims{
		// Standard claims
		"sub": userID,
		"iss": cfg.Issuer,
		"aud": cfg.Audience,
		"iat": iat.Unix(),
		"nbf": nbf.Unix(),
		"exp": exp.Unix(),
		// Legacy/custom fields for compatibility
		"user_id":    userID,
		"roleIds":    userRoleId,
		"username":   username,
		"isvalidate": isvalidate,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GenerateRefreshToken generates a long-lived JWT refresh token
func GenerateRefreshToken(userID string, userRoleId []models.UserRole, username string, isvalidate bool) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	now := time.Now()
	iat := now.Add(-cfg.Leeway / 2)
	nbf := now.Add(-cfg.Leeway / 2)
	exp := now.Add(7 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"sub":        userID,
		"iss":        cfg.Issuer,
		"aud":        cfg.Audience,
		"iat":        iat.Unix(),
		"nbf":        nbf.Unix(),
		"exp":        exp.Unix(),
		"user_id":    userID,
		"roleIds":    userRoleId,
		"username":   username,
		"isvalidate": isvalidate,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken validates the JWT token and returns the user ID (sub preferred)
func ValidateToken(tokenString string) (string, error) {
	cfg := config.LoadJWTConfigFromEnv()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}
	if uid, ok := claims["user_id"].(string); ok && uid != "" {
		return uid, nil
	}
	return "", err
}
