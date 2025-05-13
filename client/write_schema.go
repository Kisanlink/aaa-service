package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func WriteSchema() (*pb.WriteSchemaResponse, error) {
    spicedb, err := database.SpiceDB()
    if err != nil {
        log.Printf("Unable to connect to SpiceDB: %s", err)
        return nil, err
    }

    schema := `
definition user {}

definition role {
    relation has_member: user
    relation has_permission: assign_permission
}

definition assign_permission {
    relation assigned_role: role
    relation allows_action: action
}
definition action {}
`
    request := &pb.WriteSchemaRequest{Schema: schema}
    res, err := spicedb.WriteSchema(context.Background(), request)
    if err != nil {
        log.Printf("Failed to write schema: %s", err)
        return nil, err
    }

    log.Printf("Schema written successfully: %s", res)
    return res, nil
}
func ReadSchema() (*pb.ReadSchemaResponse, error) {
	// Initialize SpiceDB client
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %v", err)
		return nil, err
	}

	// Create a request object
	request := &pb.ReadSchemaRequest{}

	// Read the schema from SpiceDB
	res, err := spicedb.ReadSchema(context.Background(), request)
	if err != nil {
		log.Printf("Failed to read schema: %v", err)
		return nil, err
	}

	return res, nil
}

func UpdateSchema(roles []string, permissions []string, actions []string) (*pb.WriteSchemaResponse, error) {
    // Attempt to read the existing schema
    schemaResponse, err := ReadSchema()
    if err != nil {
        // If the error is because no schema exists, call WriteSchema to create an initial schema
        if strings.Contains(err.Error(), "No schema has been defined") {
            log.Println("No schema exists. Creating an initial schema...")
            _, err := WriteSchema()
            if err != nil {
                log.Printf("Failed to create initial schema: %v", err)
                return nil, err
            }
            // Read the schema again after creating it
            schemaResponse, err = ReadSchema()
            if err != nil {
                log.Printf("Error reading schema after creation: %v", err)
                return nil, err
            }
        } else {
            // If the error is something else, return it
            log.Printf("Error reading schema: %v", err)
            return nil, err
        }
    }

    // Extract the schema text and clean it
    existingSchema := schemaResponse.SchemaText
    existingSchema = strings.TrimSpace(existingSchema)
    // Dynamically construct updated role and permission definitions
    updatedRole := constructRoleDefinition(roles)
    updatedPermission := constructPermissionDefinition(permissions)
    updatedAction := constructActionDefinition(actions)

    // Replace the existing role and permission definitions with the updated ones
    if strings.Contains(existingSchema, "definition role {") {
        existingSchema = replaceDefinition(existingSchema, "role", updatedRole)
    } else if len(roles) > 0 {
        existingSchema += "\n" + updatedRole
    }

    if strings.Contains(existingSchema, "definition assign_permission {") {
        existingSchema = replaceDefinition(existingSchema, "assign_permission", updatedPermission)
    } else if len(permissions) > 0 {
        existingSchema += "\n" + updatedPermission
    }
    if strings.Contains(existingSchema, "definition action {") {
        existingSchema = replaceDefinition(existingSchema, "action", updatedAction)
    } else if len(actions) > 0 {
        existingSchema += "\n" + updatedAction
    }
    // Connect to SpiceDB
    spicedb, err := database.SpiceDB()
    if err != nil {
        log.Printf("Unable to connect to SpiceDB: %s", err)
        return nil, err
    }
    // Write updated schema back to SpiceDB
    request := &pb.WriteSchemaRequest{Schema: existingSchema}
    res, err := spicedb.WriteSchema(context.Background(), request)
    if err != nil {
        log.Printf("Failed to write schema: %s", err)
        return nil, err
    }
    return res, nil
}
// Helper function to construct the role definition dynamically
func constructRoleDefinition(roles []string) string {
    if len(roles) == 0 {
        return ""
    }

    roleDefinition := "definition role {\n"
    for _, role := range roles {
        // Convert role name to lowercase and replace spaces with underscores
        role = strings.ToLower(strings.ReplaceAll(role, " ", "_"))
        roleDefinition += fmt.Sprintf("    relation %s: user\n", role)
    }

    // Add the has_permission relation to connect to assign_permission
    roleDefinition += "    relation has_permission: assign_permission\n"
    roleDefinition += "}"

    return roleDefinition
}

// Helper function to construct the permission definition dynamically
func constructPermissionDefinition(permissions []string) string {
    if len(permissions) == 0 {
        return ""
    }

    permissionDefinition := "definition assign_permission {\n"
    for _, permission := range permissions {
        // Convert permission name to lowercase and replace spaces with underscores
        permission = strings.ToLower(strings.ReplaceAll(permission, " ", "_"))
        permissionDefinition += fmt.Sprintf("    relation %s: role\n", permission)
    }
    permissionDefinition += "        relation allows_action:role | action\n"
    permissionDefinition += "}"
    return permissionDefinition
}
func constructActionDefinition(actions []string) string {
    if len(actions) == 0 {
        return ""
    }

    actionDefinition := "definition action {\n"
    for _, action := range actions {
        action = strings.ToLower(strings.ReplaceAll(action, " ", "_"))
        actionDefinition += fmt.Sprintf("    relation %s: assign_permission\n", action)
    }
    actionDefinition += "}"

    return actionDefinition
}
// Helper function to replace a definition in the schema
func replaceDefinition(schema, definitionName, newDefinition string) string {
	// Find the start and end of the existing definition
	start := strings.Index(schema, "definition "+definitionName+" {")
	if start == -1 {
		return schema // Definition not found, return original schema
	}

	// Find the end of the definition (assuming it ends with a closing brace)
	end := strings.Index(schema[start:], "}") + start + 1

	// Replace the existing definition with the new one
	return schema[:start] + newDefinition + schema[end:]
}

