package database

import (
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)


func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")
	err := db.AutoMigrate(
		&model.Address{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.User{}, 
		&model.UserRole{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
func ResetMigrations(db *gorm.DB) error {
	log.Println("Resetting database...")

	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		log.Printf("Dropped table: %s\n", table)
	}
	if err := Migrate(db); err != nil {
		return fmt.Errorf("failed to run migrations after reset: %w", err)
	}

	log.Println("Database reset completed successfully")
	return nil
}