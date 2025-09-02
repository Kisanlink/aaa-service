# Organization-Scoped Response DTOs

This directory contains response DTOs for organization-scoped endpoints that provide comprehensive organization hierarchy and role management data.

## Response DTOs

### Group Responses

- **OrganizationGroupResponse**: Individual group data within organization context
- **OrganizationGroupListResponse**: Paginated list of groups within an organization
- **OrganizationGroupMemberResponse**: Group membership details with user information
- **OrganizationGroupMembersResponse**: Paginated list of group members
- **OrganizationGroupRoleResponse**: Group role assignment details
- **OrganizationGroupRolesResponse**: Paginated list of group roles

### Hierarchy Responses

- **OrganizationHierarchyGroupNode**: Hierarchical group node with roles, members, and children
- **OrganizationCompleteHierarchyResponse**: Complete organization hierarchy with statistics
- **UserGroupMembershipResponse**: User's group membership details with inheritance info
- **UserOrganizationGroupsResponse**: User's groups within an organization

### Effective Roles

- **EffectiveRoleResponse**: Role with source information (direct, group, inherited)
- **UserEffectiveRolesResponse**: User's effective roles within an organization

## Features

### Pagination Support

All list responses include pagination metadata:

- `TotalCount`: Total number of items
- `Page`: Current page number
- `PageSize`: Items per page

### Hierarchy Information

- Parent-child relationships for groups
- Role inheritance tracking
- Source attribution for roles (direct, group_direct, group_inherited)

### Rich Metadata

- Timestamps for all entities
- User information for assignments
- Active status tracking
- Time-bounded assignments

## Testing

Each response DTO includes comprehensive unit tests covering:

- Structure validation
- Field mapping verification
- Nested object relationships
- Pagination metadata
- Hierarchy construction

## Requirements Mapping

These DTOs support the following requirements:

- **7.1**: Comprehensive organization management APIs
- **7.2**: Efficient endpoints with proper pagination
- **7.4**: Consistent, meaningful responses
