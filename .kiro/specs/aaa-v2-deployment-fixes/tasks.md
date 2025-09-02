# Implementation Plan

- [x] 1. Enhance Authentication Request/Response Models

  - Create enhanced LoginRequest struct supporting both password and MPIN authentication and additional flags
  - Create comprehensive UserInfo response struct with role details and profile information
  - Add UserRoleDetail and RoleDetail response structures for complete role information controlled by flags
  - Update existing auth_requests.go to support optional password/mpin fields
  - _Requirements: 4.1, 4.2, 6.1, 6.2_

- [x] 2. Create Role Management Request/Response Models

  - Create AssignRoleRequest and AssignRoleResponse structures for role assignment API
  - Create RemoveRoleRequest structure for role removal API
  - Add validation methods for role management requests
  - Create error response structures for role-specific errors
  - _Requirements: 2.1, 2.2, 3.1, 3.2_

- [x] 3. Create MPIN Management Request Models

  - Create SetMPinRequest structure for initial MPIN setup
  - Create UpdateMPinRequest structure for MPIN updates
  - Add comprehensive validation for MPIN format and security requirements
  - Create MPIN-specific error response structures
  - _Requirements: 5.1, 5.2, 5.5, 5.6_

- [x] 4. Enhance UserRole Repository with Role Assignment Methods

  - Implement GetActiveRolesByUserID method to fetch user roles with role details
  - Implement AssignRole method with transaction support and validation
  - Implement RemoveRole method with proper constraint handling
  - Implement IsRoleAssigned method for duplicate assignment checks
  - Add unit tests for all new repository methods
  - _Requirements: 2.1, 2.4, 3.1, 8.1_

- [x] 5. Create Role Service for Role Management Operations

  - Implement RoleService interface with role assignment and removal methods
  - Add ValidateRoleAssignment method to check user and role existence
  - Implement GetUserRoles method for retrieving user role details
  - Add comprehensive error handling for role operations
  - Add unit tests for role service methods
  - _Requirements: 2.1, 2.2, 2.5, 2.6, 3.1, 3.3_

- [x] 6. Enhance User Service with MPIN and Role Support

  - Implement VerifyUserCredentials method supporting both password and MPIN authentication
  - Implement SetMPin method with secure hashing and validation
  - Implement UpdateMPin method with current MPIN verification
  - Implement VerifyMPin method for MPIN validation
  - Enhance GetUserWithRoles method to return complete role information
  - Add unit tests for all enhanced user service methods
  - _Requirements: 4.1, 4.4, 5.1, 5.3, 5.4, 1.1_

- [x] 7. Enhance User Repository with Soft Delete and Role Loading

  - Implement SoftDeleteWithCascade method for proper user deletion with role cleanup
  - Enhance GetWithActiveRoles method to efficiently load user with role details
  - Add VerifyMPin method for secure MPIN verification
  - Implement proper transaction handling for cascade operations
  - Add unit tests for enhanced repository methods
  - _Requirements: 7.1, 7.2, 7.3, 8.1, 8.4_

- [x] 8. Create Role Management Handler

  - Create RoleHandler with AssignRole endpoint (POST /users/{id}/roles)
  - Create RemoveRole endpoint (DELETE /users/{id}/roles/{role_id})
  - Add proper request validation and error handling
  - Implement authorization checks for role management operations
  - Add comprehensive logging for role operations
  - Add unit tests for role handler methods
  - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.2, 9.3_

- [x] 9. Enhance Authentication Handler with MPIN Support

  - Update LoginV2 method to support both password and MPIN authentication
  - Enhance login response to include complete user information with roles
  - Update authentication logic to handle optional password/mpin fields
  - Add proper error handling for authentication failures
  - Add unit tests for enhanced authentication handler
  - _Requirements: 4.1, 4.2, 4.3, 1.1, 1.2, 1.3_

- [x] 10. Create MPIN Management Handler

  - Create SetMPinV2 endpoint for initial MPIN setup
  - Create UpdateMPinV2 endpoint for MPIN updates
  - Add proper authentication middleware integration
  - Implement secure MPIN validation and error handling
  - Add comprehensive logging for MPIN operations
  - Add unit tests for MPIN handler methods
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 9.1, 9.2_

- [x] 11. Enhance User Handler with Improved Deletion

  - Update DeleteUser method to use soft delete with proper cascade handling
  - Enhance error handling to return proper HTTP status codes
  - Add transaction support for user deletion operations
  - Implement proper cleanup of user relationships (roles, profiles, contacts)
  - Add unit tests for enhanced user deletion
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [x] 12. Add Route Configurations for New Endpoints

  - Add routes for role management endpoints (/users/{id}/roles)
  - Add routes for MPIN management endpoints (/auth/set-mpin, /auth/update-mpin)
  - Configure proper middleware for authentication and authorization
  - Add rate limiting for sensitive operations
  - Update API documentation with new endpoints
  - _Requirements: 2.1, 3.1, 5.1, 9.4_

- [x] 13. Implement Security Enhancements

  - Add rate limiting for authentication attempts and MPIN operations
  - Implement audit logging for all security-sensitive operations
  - Add input sanitization and validation for all new endpoints
  - Implement proper error messages that don't leak sensitive information
  - Add security headers and CORS configuration
  - _Requirements: 9.1, 9.2, 9.4, 9.5, 10.5_

- [x] 14. Add Database Indexes for Performance

  - Create index on user_roles(user_id, is_active) for efficient role queries
  - Create composite index on user_roles(user_id, role_id) for assignment checks
  - Verify existing indexes on users(phone_number, country_code) for authentication
  - Add database migration scripts for new indexes
  - Test query performance with new indexes
  - _Requirements: 8.1, 8.2_

- [x] 15. Implement Caching for Role and User Data

  - Add caching for user role information with appropriate TTL
  - Implement cache invalidation for role assignment/removal operations
  - Add caching for user profile data in login responses
  - Implement cache warming strategies for frequently accessed data
  - Add unit tests for caching functionality
  - _Requirements: 8.2, 8.3_

- [ ] 16. Create Integration Tests for Authentication Flows

  - Create end-to-end tests for password-based authentication
  - Create end-to-end tests for MPIN-based authentication
  - Test role assignment and removal workflows
  - Test user deletion with proper cleanup
  - Test MPIN management workflows
  - _Requirements: 1.1, 4.1, 2.1, 7.1, 5.1_

- [x] 17. Add Comprehensive Error Handling and Logging

  - Implement consistent error response format across all endpoints
  - Add structured logging for all operations with request IDs
  - Create audit logs for security-sensitive operations
  - Add error monitoring and alerting configuration
  - Test error scenarios and response formats
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 18. Update API Documentation and Examples
  - Update Swagger documentation for enhanced login endpoint
  - Add documentation for role management endpoints
  - Add documentation for MPIN management endpoints
  - Create API usage examples for all new functionality
  - Update error response documentation
  - _Requirements: 1.1, 2.1, 5.1, 10.1_
