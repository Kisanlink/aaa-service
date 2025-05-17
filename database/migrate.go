package database

import (
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/model"
)

// ModelRegistry maps table names to their model structs
var ModelRegistry = map[string]interface{}{
	"user":            &model.User{},
	"role":            &model.Role{},
	"permission":      &model.Permission{},
	"address":         &model.Address{},
	"role_permission": &model.RolePermission{},
	"user_role":       &model.UserRole{},
}

// Migrate runs migrations for all or specific tables
func Migrate(tables ...string) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	var models []interface{}

	if len(tables) == 0 {
		// Migrate all tables
		fmt.Println("Running migrations for all tables...")
		for _, model := range ModelRegistry {
			models = append(models, model)
		}
	} else {
		// Migrate specific tables
		fmt.Printf("Running migrations for tables: %v\n", tables)
		for _, table := range tables {
			if model, exists := ModelRegistry[strings.ToLower(table)]; exists {
				models = append(models, model)
			} else {
				return fmt.Errorf("unknown table: %s", table)
			}
		}
	}

	err := DB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("error migrating database: %v", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}

// Reset drops tables and runs migrations
func Reset(tables ...string) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	var models []interface{}

	if len(tables) == 0 {
		// Reset all tables
		fmt.Println("Resetting all tables...")
		for table, model := range ModelRegistry {
			fmt.Printf("Dropping table: %s\n", table)
			if err := DB.Migrator().DropTable(model); err != nil {
				return fmt.Errorf("error dropping table %s: %v", table, err)
			}
			models = append(models, model)
		}
	} else {
		// Reset specific tables
		fmt.Printf("Resetting tables: %v\n", tables)
		for _, table := range tables {
			table = strings.ToLower(table)
			if model, exists := ModelRegistry[table]; exists {
				fmt.Printf("Dropping table: %s\n", table)
				if err := DB.Migrator().DropTable(model); err != nil {
					return fmt.Errorf("error dropping table %s: %v", table, err)
				}
				models = append(models, model)
			} else {
				return fmt.Errorf("unknown table: %s", table)
			}
		}
	}

	// Run migrations for the affected tables
	err := DB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("error migrating database: %v", err)
	}

	fmt.Println("Reset and migrations completed successfully")
	return nil
}
