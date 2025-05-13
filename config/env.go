package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found (%v), using system environment variables", err)
	} else {
		log.Println(".env file loaded successfully")
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Warning: Environment variable %s is empty or not set", key)
	}
	return value
}
