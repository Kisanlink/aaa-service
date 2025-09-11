# Implementation Plan

- [x] 1. Create GroupRole model and database migration

  - Add GroupRole model to internal/entities/models/ following existing patterns
  - Create database migration using automigrate for group_roles table with proper indexes
  - Add GORM hooks and helper methods consistent with existing models
  - Write unit tests for GroupRole model validation and relationships
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 2. Extend organization routes with group management endpoints

  - Update internal/routes/organization_routes.go to add organization-scoped group routes
  - Add routes for /:orgId/groups, /:orgId/groups/:groupId/users, /:orgId/groups/:groupId/roles
  - Ensure proper middleware integration with existing auth middleware
  - Test route registration and parameter extraction
  - _Requirements: 1.1, 1.2, 1.3, 7.1_

- [x] 3. Add organization-scoped group handler methods

  - Extend internal/handlers/organizations/organization_handler.go with group management methods
  - Implement GetOrganizationGroups, CreateGroupInOrganization, GetGroupInOrganization handlers
  - Add proper error handling and response formatting using existing responder patterns
  - Write unit tests for new handler methods with mocked services
  - _Requirements: 1.1, 1.2, 7.2, 7.4_

- [x] 4. Implement user-group assignment handlers within organization context

  - Add AddUserToGroupInOrganization and RemoveUserFromGroupInOrganization handlers
  - Implement GetGroupUsersInOrganization and GetUserGroupsInOrganization handlers
  - Ensure organization-level isolation and validation
  - Write unit tests for user-group assignment handlers
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 6.1, 6.2, 6.3_

- [x] 5. Implement role-group assignment handlers

  - Add AssignRoleToGroupInOrganization and RemoveRoleFromGroupInOrganization handlers
  - Implement GetGroupRolesInOrganization handler
  - Add validation for role-group assignments within organization boundaries
  - Write unit tests for role-group assignment handlers
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 6.1, 6.2, 6.3_

- [x] 6. Extend OrganizationService interface and implementation

  - Add new methods to internal/interfaces/interfaces.go OrganizationService interface
  - Implement GetOrganizationGroups and related methods in organization service
  - Add proper error handling and validation logic
  - Write unit tests for new organization service methods
  - _Requirements: 7.1, 7.2, 7.3, 6.4_

- [x] 7. Extend GroupService interface with role assignment methods

  - Add AssignRoleToGroup, RemoveRoleFromGroup, GetGroupRoles methods to GroupService interface
  - Implement organization-scoped group operations in existing group service
  - Add validation for cross-organization access prevention
  - Write unit tests for extended group service methods
  - _Requirements: 4.1, 4.2, 4.3, 6.1, 6.2, 6.3_

- [x] 8. Create role inheritance engine for bottom up inheritance

  - Implement RoleInheritanceEngine with CalculateEffectiveRoles method
  - Add logic for upward inheritance (parent groups)
  - Implement conflict resolution with most-specific-wins precedence
  - Write comprehensive unit tests for various inheritance scenarios
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 9. Implement GetUserEffectiveRoles handler and service method

  - Add GetUserEffectiveRolesInOrganization handler method
  - Integrate role inheritance engine into service layer
  - Add caching for effective roles calculation to improve performance
  - Write unit tests and integration tests for effective roles calculation
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 10. Add organization hierarchy endpoints with group integration

  - Enhance GetOrganizationHierarchy to include group structures and role assignments
  - Implement efficient hierarchy queries using existing Group parent-child relationships
  - Add proper response formatting for complex hierarchy data
  - Write unit tests for hierarchy retrieval with groups and roles
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 7.3_

- [ ] 11. Implement caching layer for performance optimization

  - Add Redis caching for organization hierarchies, user groups, and effective roles
  - Implement cache invalidation strategies for hierarchy and role changes
  - Add cache warming for frequently accessed organization data
  - Write integration tests for caching behavior and invalidation
  - _Requirements: 7.3, 7.5_

- [x] 12. Add request/response DTOs for new organization-scoped endpoints

  - Create request DTOs for group creation, user assignment, and role assignment within organizations
  - Add response DTOs for organization hierarchy, effective roles, and group listings
  - Ensure DTOs follow existing patterns in internal/entities/requests/ and internal/entities/responses/
  - Write validation tests for new DTOs
  - _Requirements: 7.1, 7.2, 7.4_

- [x] 13. Implement GroupRole repository with CRUD operations

  - Create GroupRoleRepository in internal/repositories/groups/ with kisanlink-db integration
  - Implement GetByGroupID, GetByRoleID, Create, Update, Delete methods
  - Add proper error handling and transaction support
  - Write unit tests for repository operations
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 14. Complete role-group assignment handlers implementation

  - Add AssignRoleToGroupInOrganization and RemoveRoleFromGroupInOrganization handlers
  - Implement GetGroupRolesInOrganization handler
  - Add validation for role-group assignments within organization boundaries
  - Write unit tests for role-group assignment handlers
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 6.1, 6.2, 6.3_

- [x] 15. Fix role inheritance engine group membership integration

  - Implement getUserDirectGroups method to query actual group memberships
  - Create GroupMembershipRepository if not exists or integrate with existing repository
  - Update role inheritance engine to work with real group membership data
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [ ] 16. Add comprehensive audit logging for organization operations

  - Extend existing audit logging to capture organization structure changes
  - Log group membership changes, role assignments, and hierarchy modifications
  - Ensure audit logs are organization-scoped and tamper-proof
  - Write unit tests for audit logging functionality
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 17. Create integration tests for multi-tenant isolation

  - Write integration tests to verify organization-level data isolation
  - Test cross-organization access prevention for groups, users, and roles
  - Verify that organization administrators can only access their own organization data
  - Test edge cases for multi-tenant security
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 18. Update API documentation and Swagger specs

  - Add Swagger annotations for all new organization-scoped endpoints
  - Update docs/swagger.yaml and docs/swagger.json with new API definitions
  - Ensure proper parameter documentation and response examples
  - Test API documentation generation and accuracy
  - _Requirements: 7.1, 7.4_

- [ ] 19. Enhance bottom-up role inheritance implementation

  - Verify and test the existing upward inheritance (parent groups inherit roles from child groups)
  - Ensure role inheritance engine correctly implements bottom-up inheritance only
  - Add comprehensive tests for upward inheritance scenarios
  - Document the inheritance flow clearly in code comments
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_
