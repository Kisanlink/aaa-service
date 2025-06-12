package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

func ReadSchema() (*pb.ReadSchemaResponse, error) {
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %v", err)
		return nil, err
	}

	request := &pb.ReadSchemaRequest{}
	res, err := spicedb.ReadSchema(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ReadSchema reads the SpiceDB schema and returns a map of definition names to their relations
func ReadDefinitionSchema(client *authzed.Client) (map[string][]string, error) {
	resp, err := ReadSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to read schema: %w", err)
	}

	// Parse the schema to extract definitions and their relations
	definitions := make(map[string][]string)

	// Simple parsing of the schema text to extract definitions and relations
	// In a production environment, you should use a proper parser
	schemaText := resp.SchemaText
	lines := strings.Split(schemaText, "\n")

	var currentDef string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if this line starts a new definition
		if strings.HasPrefix(line, "definition ") {
			defParts := strings.Split(line, " ")
			if len(defParts) >= 2 {
				currentDef = defParts[1]
				if !strings.HasSuffix(currentDef, "{") {
					currentDef = strings.TrimSuffix(currentDef, "{")
				}
				currentDef = strings.TrimSpace(currentDef)

				// Skip the "user" definition as specified
				if currentDef != "user" {
					definitions[currentDef] = []string{}
				}
			}
		} else if currentDef != "" && currentDef != "user" && strings.HasPrefix(line, "relation ") {
			// Extract relation name
			relParts := strings.Split(line, ":")
			if len(relParts) >= 1 {
				relName := strings.TrimPrefix(relParts[0], "relation ")
				relName = strings.TrimSpace(relName)
				definitions[currentDef] = append(definitions[currentDef], relName)
			}
		}
	}
	return definitions, nil
}
