package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/aaa-service/database"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

// const schema = `
// definition user {}

// definition role {
//     relation super_admin: user
//     relation admin: user
//     relation ceo: user
//     relation director: user
//     relation shareholder: user
//     relation farmer: user
//     relation kisansathi: user
//     relation collaborator: user
//     relation staff: user
//     relation system_staff: user
//     relation consumer: user
//     relation student: user
//     relation company_manager: user
//     relation company_owner: user

// 	permission all_all = super_admin
// 	permission all = admin
// 	permission fpo = ceo + director + shareholder + staff + system_staff
// 	permission kisanlink = farmer + kisansathi + collaborator
// 	permission amrti = consumer
// 	permission asa = student + company_manager + company_owner
// }
// `


func WriteSchema()(*pb.WriteSchemaResponse ,error){
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Fatalf("unable to connect spice db: %s", err)
	}
    schema := `
definition user {}

definition role {
    relation has_member: user
    relation has_permission: assign_permission
}

definition assign_permission {
    relation assigned_role: role
}
`
	request := &pb.WriteSchemaRequest{Schema: schema}
	res, err := spicedb.WriteSchema(context.Background(), request)
	if err != nil {
		log.Fatalf("failed to write schema: %s", err)
	}
	log.Fatalf("result: %s", res)
	return res ,err
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

	// // Print the schema response properly
	// log.Printf("Schema read successfully: %+v", res)

	return res, nil
}




func UpdateSchema(roles []string, permissions []string) (*pb.WriteSchemaResponse, error) {
	// Read existing schema
	schemaResponse, err := ReadSchema()
	if err != nil {
		log.Printf("Error reading schema: %v", err)
		return nil, err
	}

	//  Extract the schema text and clean it
	existingSchema := schemaResponse.SchemaText
	existingSchema = strings.TrimSpace(existingSchema)

	//  Dynamically construct updated role and permission definitions
	updatedRole := constructRoleDefinition(roles)
	updatedPermission := constructPermissionDefinition(permissions)

	// Replace the existing role and permission definitions with the updated ones
	// Check if the role definition exists and replace it
	if strings.Contains(existingSchema, "definition role {") {
		existingSchema = replaceDefinition(existingSchema, "role", updatedRole)
	} else if len(roles) > 0 {
		// If the role definition doesn't exist but roles are provided, append it
		existingSchema += "\n" + updatedRole
	}

	// Check if the permission definition exists and replace it
	if strings.Contains(existingSchema, "definition assign_permission {") {
		existingSchema = replaceDefinition(existingSchema, "assign_permission", updatedPermission)
	} else if len(permissions) > 0 {
		// If the permission definition doesn't exist but permissions are provided, append it
		existingSchema += "\n" + updatedPermission
	}

	//  Remove the `new_role` definition if it exists
	// existingSchema = removeDefinition(existingSchema, "new_role")

	// log.Printf("Updated Schema: %s", existingSchema)

	// Connect to SpiceDB
	spicedb, err := database.SpiceDB()
	if err != nil {
		log.Printf("Unable to connect to SpiceDB: %s", err)
		return nil, err
	}

	//  Write updated schema back to SpiceDB
	request := &pb.WriteSchemaRequest{Schema: existingSchema}
	res, err := spicedb.WriteSchema(context.Background(), request)
	if err != nil {
		log.Printf("Failed to write schema: %s", err)
		return nil, err
	}

	log.Printf("Schema update result: %s", res)
	return res, nil
}

// Helper function to construct the role definition dynamically
func constructRoleDefinition(roles []string) string {
	if len(roles) == 0 {
		return ""
	}

	roleDefinition := "definition role {\n"
	for _, role := range roles {
		roleDefinition += fmt.Sprintf("    relation %s: user\n", role)
	}
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
		permissionDefinition += fmt.Sprintf("    relation %s: role\n", permission)
	}
	permissionDefinition += "}"
	return permissionDefinition
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

// func removeDefinition(schema, definitionName string) string {
// 	start := strings.Index(schema, "definition "+definitionName+" {")
// 	if start == -1 {
// 		return schema 
// 	}
// 	end := strings.Index(schema[start:], "}") + start + 1
// 	return strings.TrimSpace(schema[:start] + schema[end:])
// }