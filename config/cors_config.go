package config

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
)

type ServerConfig struct {
	AllowedOrigins   []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	Port             string
}

func LoadConfig() *ServerConfig {
	allowedOrigins := strings.Split(GetEnv("CORS_ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173"}
	}

	allowMethods := strings.Split(GetEnv("CORS_ALLOW_METHODS"), ",")
	if len(allowMethods) == 0 {
		allowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	allowHeaders := strings.Split(GetEnv("CORS_ALLOW_HEADERS"), ",")
	if len(allowHeaders) == 0 {
		allowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	}

	exposeHeaders := strings.Split(GetEnv("CORS_EXPOSE_HEADERS"), ",")
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"Content-Length"}
	}

	port := GetEnv("PORT")
	if port == "" {
		port = "8080"
	}

	return &ServerConfig{
		AllowedOrigins:   allowedOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: GetEnv("CORS_ALLOW_CREDENTIALS") != "false",
		Port:             port,
	}
}

func GetCorsConfig(config *ServerConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}
}
