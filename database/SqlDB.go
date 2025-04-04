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

	// if err := DB.AutoMigrate(
	// 	&model.Address{},
	// 	&model.Role{},
	// 	&model.Permission{},
	// 	&model.RolePermission{},
	// 	&model.User{},
	// 	&model.UserRole{},
	// ); err != nil {
	// 	panic("Error migrating database: " + err.Error())
	// }

	// start from here and remove line later

	if err := DB.AutoMigrate(&model.Address{}); err != nil {
		panic("Error migrating Address table: " + err.Error())
	}

	if err := DB.AutoMigrate(&model.Role{}); err != nil {
		panic("Error migrating Role table: " + err.Error())
	}

	if err := DB.AutoMigrate(&model.Permission{}); err != nil {
		panic("Error migrating Permission table: " + err.Error())
	}

	if err := DB.AutoMigrate(&model.RolePermission{}); err != nil {
		panic("Error migrating RolePermission table: " + err.Error())
	}

	if err := DB.AutoMigrate(&model.User{}); err != nil {
		panic("Error migrating User table: " + err.Error())
	}

	if err := DB.AutoMigrate(&model.UserRole{}); err != nil {
		panic("Error migrating UserRole table: " + err.Error())
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Sprintf("Failed to get underlying sql.DB: %v", err))
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0)

	fmt.Println("DB Connection pool configured for 100 connections")
	// fmt.Println("DB Migration complete")
}
