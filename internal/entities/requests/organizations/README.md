# Organization-Scoped Request DTOs

This directory contains request DTOs for organization-scoped endpoints that enable hierarchical group management within organizations.

## Request DTOs

### Group Management

- **CreateOrganizationGroupRequest**: Create groups within an organization context
- **UpdateOrganizationGroupRequest**: Update group properties within an organization
- **RemoveUserFromGroupRequest**: Remove users from organization groups (URL-based)
- **RemoveRoleFromGroupRequest**: Remove roles from organization groups (URL-based)

### User-Group Assignment

- **AssignUserToGroupRequest**: Assign users to groups within organizations with time bounds

### Role-Group Assignment

- **AssignRoleToGroupRequest**: Assign roles to groups within organizations with time bounds

## Validation

All request DTOs include comprehensive validation using struct tags:

- Required fields validation
- UUID format validation for IDs
- String length constraints
- Enum validation for principal types
- Time bounds validation

## Testing

Each DTO includes comprehensive unit tests covering:

- Valid request scenarios
- Invalid field validation
- Edge cases and boundary conditions
- Structure verification

## Requirements Mapping

These DTOs support the following requirements:

- **7.1**: CRUD operations for organizations, groups, and relationships
- **7.2**: Validation of all relationships and constraints
- **7.4**: Consistent, meaningful error responses
