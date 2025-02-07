package database

import (
	"fmt"
	"os"

	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var err error

func ConnectDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL environment variable is not set")
	}
	fmt.Println("DATABASE_URL environment variable is set...")

	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Discard.LogMode(logger.Error),
	})
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}
	fmt.Println("Connected to DB")

	migrationErr := DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
	)
	if migrationErr != nil {
		panic(fmt.Sprintf("Error migrating database: %v", migrationErr))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Sprintf("Failed to get underlying sql.DB: %v", err))
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0)

	fmt.Println("DB Connection pool configured for 100 connections")
	fmt.Println("DB Migration complete")
}
