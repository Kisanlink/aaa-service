# Requirements Document

## Introduction

This feature addresses critical deployment gaps in the AAA v2 service that are preventing proper user authentication, role management, and user lifecycle operations. The current system has several production-blocking issues including missing role information in login responses, lack of role assignment capabilities, incomplete MPIN authentication support, and broken user deletion functionality. These fixes are essential for a production-ready authentication and authorization service.

## Requirements

### Requirement 1: Role Information in Login Response

**User Story:** As a client application, I want to receive complete role information when a user logs in, so that I can properly authorize user actions and display appropriate UI elements.

#### Acceptance Criteria

1. WHEN a user successfully logs in THEN the system SHALL return the user's assigned roles with complete role details including role ID, name, description, and active status
2. WHEN a user has multiple roles THEN the system SHALL return all active roles in an array format
3. WHEN a user has no assigned roles THEN the system SHALL return an empty array instead of null
4. WHEN role information is retrieved by the user THEN the system SHALL include the user_role relationship details including assignment status and timestamps

### Requirement 2: Role Assignment API

**User Story:** As an administrator, I want to assign roles to existing users through a dedicated API, so that I can manage user permissions after account creation.

#### Acceptance Criteria

1. WHEN an administrator calls the role assignment API THEN the system SHALL create a user-role relationship in the database
2. WHEN assigning a role THEN the system SHALL validate that both the user and role exist and are active
3. WHEN a role is successfully assigned THEN the system SHALL return confirmation with user ID and role details
4. WHEN attempting to assign a duplicate role THEN the system SHALL return an appropriate error message
5. WHEN assigning a role to a non-existent user THEN the system SHALL return a 404 error
6. WHEN assigning a non-existent role THEN the system SHALL return a 404 error

### Requirement 3: Role Removal API

**User Story:** As an administrator, I want to remove roles from users through a dedicated API, so that I can revoke permissions when needed.

#### Acceptance Criteria

1. WHEN an administrator calls the role removal API THEN the system SHALL remove the user-role relationship from the database
2. WHEN removing a role THEN the system SHALL validate that the user-role relationship exists
3. WHEN a role is successfully removed THEN the system SHALL return confirmation of the removal
4. WHEN attempting to remove a non-existent role assignment THEN the system SHALL return a 404 error

### Requirement 4: MPIN Authentication Support

**User Story:** As a user, I want to authenticate using my MPIN instead of a password, so that I can have a more convenient and secure mobile authentication experience.

#### Acceptance Criteria

1. WHEN a user submits login credentials THEN the system SHALL accept either password or MPIN for authentication
2. WHEN authenticating with MPIN THEN the system SHALL validate the MPIN against the stored hash
3. WHEN both password and MPIN are provided THEN the system SHALL prioritize password authentication
4. WHEN MPIN authentication fails THEN the system SHALL return appropriate error messages
5. WHEN a user has no MPIN set THEN MPIN authentication SHALL fail with a clear error message

### Requirement 5: MPIN Management

**User Story:** As a user, I want to set and update my MPIN, so that I can use MPIN-based authentication for my account.

#### Acceptance Criteria

1. WHEN creating a new user account THEN the system SHALL optionally accept an MPIN during registration
2. WHEN an MPIN is provided during registration THEN the system SHALL hash and store it securely
3. WHEN a user wants to set an MPIN after registration THEN the system SHALL provide an API to set the initial MPIN
4. WHEN a user wants to update their MPIN THEN the system SHALL require the current MPIN for verification
5. WHEN updating an MPIN THEN the system SHALL validate the new MPIN meets security requirements
6. WHEN an MPIN is set or updated THEN the system SHALL update the has_mpin flag to true

### Requirement 6: Enhanced Login Response

**User Story:** As a client application, I want to receive comprehensive user information in the login response, so that I can provide a complete user experience without additional API calls.

#### Acceptance Criteria

1. WHEN a user successfully logs in THEN the system SHALL return complete user profile information including full name and email
2. WHEN returning user information THEN the system SHALL include user profile details if they exist and controlled by a flag
3. WHEN a user has contact information THEN the system SHALL include relevant contact details in the response controlled by a flag
4. WHEN user information is incomplete THEN the system SHALL return available fields and indicate missing information even when flag is true

### Requirement 7: User Deletion Functionality

**User Story:** As an administrator and Super Admin, I want to delete user accounts through the API, so that I can manage user lifecycle and comply with data retention policies.

#### Acceptance Criteria

1. WHEN an administrator deletes a user THEN the system SHALL perform a soft delete by default to maintain audit trails
2. WHEN deleting a user THEN the system SHALL remove or deactivate all associated role assignments
3. WHEN deleting a user THEN the system SHALL handle cascade operations for related entities (contacts, profiles, etc.)
4. WHEN a user is successfully deleted THEN the system SHALL return a 200 OK response with confirmation details
5. WHEN attempting to delete a non-existent user THEN the system SHALL return a 404 error
6. WHEN deletion fails due to constraints THEN the system SHALL return appropriate error messages with details

### Requirement 8: Data Consistency and Integrity

**User Story:** As a system administrator, I want all user and role operations to maintain data consistency, so that the system remains reliable and secure.

#### Acceptance Criteria

1. WHEN performing role assignments THEN the system SHALL use database transactions to ensure consistency
2. WHEN deleting users THEN the system SHALL handle foreign key constraints properly
3. WHEN updating user information THEN the system SHALL validate data integrity before committing changes
4. WHEN operations fail THEN the system SHALL rollback partial changes to maintain consistency
5. WHEN concurrent operations occur THEN the system SHALL handle race conditions appropriately

### Requirement 9: Security and Validation

**User Story:** As a security administrator, I want all authentication and authorization operations to follow security best practices, so that the system remains secure against common attacks.

#### Acceptance Criteria

1. WHEN storing MPINs THEN the system SHALL use secure hashing algorithms (bcrypt or similar)
2. WHEN validating MPINs THEN the system SHALL enforce minimum security requirements (length, complexity)
3. WHEN performing role operations THEN the system SHALL validate administrator permissions
4. WHEN handling authentication THEN the system SHALL implement rate limiting and brute force protection
5. WHEN logging operations THEN the system SHALL create audit trails for security-sensitive actions

### Requirement 10: Error Handling and Logging

**User Story:** As a developer and system administrator, I want comprehensive error handling and logging, so that I can troubleshoot issues and monitor system health.

#### Acceptance Criteria

1. WHEN errors occur THEN the system SHALL return consistent error response formats
2. WHEN operations fail THEN the system SHALL log detailed error information for debugging
3. WHEN returning errors THEN the system SHALL include appropriate HTTP status codes
4. WHEN logging events THEN the system SHALL include request IDs for traceability
5. WHEN handling sensitive operations THEN the system SHALL log security events appropriately
