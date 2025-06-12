// database/migrate.go
package database

import (
	"fmt"
	"os"
)

// HandleMigrationCommands processes migration-related command line arguments
func HandleMigrationCommands() bool {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			handleMigrateCommand()
			return true
		case "reset":
			handleResetCommand()
			return true
		}
	}
	return false
}

// handleMigrateCommand processes the migrate command
func handleMigrateCommand() {
	if len(os.Args) > 2 {
		if err := Migrate(os.Args[2:]...); err != nil {
			fmt.Printf("Migration error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := Migrate(); err != nil {
			fmt.Printf("Migration error: %v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}

// handleResetCommand processes the reset command
func handleResetCommand() {
	if len(os.Args) > 2 {
		if err := Reset(os.Args[2:]...); err != nil {
			fmt.Printf("Reset error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := Reset(); err != nil {
			fmt.Printf("Reset error: %v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}
