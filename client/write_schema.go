package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/aaa-service/database"
	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func WriteInitialSchema() (*pb.WriteSchemaResponse, error) {
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	// Basic initial schema with just the user definition
	schema := `definition user {}`
	request := &pb.WriteSchemaRequest{Schema: schema}
	res, err := spicedb.WriteSchema(context.Background(), request)
	if err != nil {
		log.Printf("Failed to write initial schema: %s", err)
		return nil, err
	}

	log.Printf("Initial schema written successfully")
	return res, nil
}

func UpdateSchema(schemas []model.CreateSchema) (*pb.WriteSchemaResponse, error) {
	// Build the complete schema from scratch
	var fullSchemaParts []string

	// Always include the base user definition
	fullSchemaParts = append(fullSchemaParts, "definition user {}")

	for _, schema := range schemas {
		newDefinition := constructResourceDefinition(schema)
		fullSchemaParts = append(fullSchemaParts, newDefinition)
	}

	// Combine all definitions into the full schema
	fullSchema := strings.Join(fullSchemaParts, "\n\n")

	// Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	// Write the new schema to SpiceDB
	request := &pb.WriteSchemaRequest{Schema: fullSchema}
	res, err := spicedb.WriteSchema(context.Background(), request)
	if err != nil {
		log.Printf("Failed to write schema: %s", err)
		return nil, err
	}

	log.Printf("New schema written successfully (overwriting any existing schema)")
	return res, nil
}

func constructResourceDefinition(schema model.CreateSchema) string {
	var builder strings.Builder

	// Start the definition
	builder.WriteString(fmt.Sprintf("definition %s {\n", schema.Resource))

	// Add all relations
	for _, relation := range schema.Relations {
		builder.WriteString(fmt.Sprintf("    relation %s: user\n", relation))
	}

	// Add a newline between relations and permissions if both exist
	if len(schema.Relations) > 0 && len(schema.Data) > 0 {
		builder.WriteString("\n")
	}

	// Add all permissions
	for _, permission := range schema.Data {
		// Combine roles with "+" syntax
		roles := strings.Join(permission.Roles, " + ")
		builder.WriteString(fmt.Sprintf("    permission %s = %s\n", permission.Action, roles))
	}

	// Close the definition
	builder.WriteString("}")

	return builder.String()
}
