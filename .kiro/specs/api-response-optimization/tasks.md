# Implementation Plan

- [ ] 1. Create response transformation infrastructure

  - Implement ResponseTransformer interface and core transformation logic
  - Create standardized response models with consistent snake_case field naming
  - Add query parameter parsing utilities for include/exclude controls
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.3, 4.4_

- [ ] 2. Implement standardized response models
- [ ] 2.1 Create base response structures with consistent field naming

  - Write UserResponse, RoleResponse, and UserRoleResponse structs with proper JSON tags
  - Implement BaseResponse struct with standardized timestamp fields
  - Create response validation utilities to ensure field naming consistency
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.3, 4.4_

- [ ] 2.2 Implement sensitive field exclusion in response models

  - Add JSON exclusion tags for password and MPIN fields in all user-related responses
  - Create transformation logic to filter sensitive data from nested objects
  - Write unit tests to verify sensitive fields are never included in responses
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 2.3 Create consistent state representation logic

  - Implement logic to ensure is_active and deleted_at fields have consistent relationship
  - Add transformation rules to normalize invalid state combinations
  - Write validation functions for timestamp consistency (UTC format)
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ] 3. Implement query parameter control system
- [ ] 3.1 Create query parameter handler for response control

  - Write QueryParameterHandler struct with parsing logic for include_user, include_role, include_profile flags
  - Implement TransformOptions struct to control nested object inclusion
  - Add validation for allowed query parameters and default behaviors
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 9.4_

- [ ] 3.2 Implement conditional response transformation

  - Modify ResponseTransformer to use TransformOptions for conditional field inclusion
  - Add logic to return null/omit nested objects when include flags are false
  - Create helper functions to determine when to include nested relationship data
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 4. Migrate core user management endpoints
- [ ] 4.1 Update user endpoints to use new response transformation

  - Modify GET /users/{id} endpoint to use UserResponse with transformation options
  - Update GET /users endpoint to use consistent response structure
  - Implement query parameter support for user detail inclusion
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4_

- [ ] 4.2 Update user role assignment endpoints

  - Modify GET /users/{id}/roles endpoint to eliminate redundant user_id fields
  - Implement lightweight role assignment responses with optional nested objects
  - Add query parameter controls for including full user/role details
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 5.1, 5.2, 5.3, 5.4_

- [ ] 5. Implement authentication validation improvements
- [ ] 5.1 Create token invalidation tracking system

  - Implement token blacklist or invalidation tracking in cache/database
  - Add middleware to check token validity against invalidation list
  - Create utilities to invalidate tokens on logout and password changes
  - _Requirements: 10.1, 10.2, 10.3, 10.4_

- [ ] 5.2 Update authentication error responses

  - Create standardized AuthenticationError response structure
  - Modify authentication middleware to return consistent 401 responses
  - Add proper error messaging for invalidated tokens
  - _Requirements: 10.1, 10.2, 10.3, 10.4_

- [ ] 6. Implement password management endpoints
- [ ] 6.1 Create password change endpoint

  - Implement POST /users/{id}/password/change endpoint with current password validation
  - Add password strength validation using PasswordValidationResponse
  - Implement token invalidation after successful password change
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [ ] 6.2 Create password reset request endpoint

  - Implement POST /auth/password/reset/request endpoint for OTP generation
  - Add contact verification and OTP sending logic
  - Create PasswordResetOTPResponse with transaction_id and expiration details
  - _Requirements: 13.1, 13.2, 13.6_

- [ ] 6.3 Create password reset verification endpoint

  - Implement POST /auth/password/reset/verify endpoint for OTP verification and password reset
  - Add OTP validation and expiration checking
  - Implement token invalidation after successful password reset
  - _Requirements: 13.3, 13.4, 13.5, 13.6_

- [ ] 7. Implement Aadhaar verification integration
- [ ] 7.1 Create Aadhaar verification service interface

  - Implement AadhaarVerificationService interface with OTP generation and verification methods
  - Create AadhaarOTPResponse and AadhaarVerificationResponse structures
  - Add integration with external aadhaar-verification service
  - _Requirements: 11.1, 11.2, 11.6_

- [ ] 7.2 Create Aadhaar verification endpoints

  - Implement POST /users/{id}/aadhaar/otp endpoint for OTP generation
  - Implement POST /users/{id}/aadhaar/verify endpoint for OTP verification
  - Add GET /users/{id}/verification-status endpoint for verification status
  - _Requirements: 11.1, 11.2, 11.3, 11.6_

- [ ] 7.3 Enhance user responses with verification status

  - Add VerificationStatusResponse to UserResponse structure
  - Implement query parameter support for include_verification flag
  - Add proper masking of sensitive Aadhaar information in responses
  - _Requirements: 11.3, 11.4, 11.5, 11.6_

- [ ] 8. Create comprehensive response validation system
- [ ] 8.1 Implement response structure validation

  - Create ResponseValidator interface with validation methods for all response types
  - Add automated tests to verify consistent field naming across endpoints
  - Implement validation for response structure consistency
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 9.1, 9.2, 9.3_

- [ ] 8.2 Create integration tests for response consistency

  - Write tests to compare user object structures across different endpoints
  - Add tests to verify query parameter functionality works correctly
  - Create tests to validate sensitive field exclusion across all endpoints
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [ ] 9. Implement error response standardization
- [ ] 9.1 Create standardized error response structures

  - Implement ErrorResponse struct with consistent field naming
  - Add error response transformation for all endpoint error cases
  - Create error response validation to ensure consistency
  - _Requirements: 10.2, 10.3, 12.4, 13.5_

- [ ] 9.2 Update all endpoints to use standardized error responses

  - Modify all handlers to return consistent error response structures
  - Add proper HTTP status codes for different error scenarios
  - Implement error response logging with appropriate context
  - _Requirements: 10.2, 10.3, 10.4_

- [ ] 10. Create performance optimization and monitoring
- [ ] 10.1 Implement response payload optimization

  - Add logic to eliminate redundant fields in nested objects
  - Implement response size monitoring and logging
  - Create performance benchmarks for response transformation
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 10.2 Add response caching for frequently accessed data

  - Implement caching layer for transformed responses
  - Add cache invalidation logic for data updates
  - Create cache performance monitoring and metrics
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 11. Create comprehensive test suite
- [ ] 11.1 Write unit tests for response transformation

  - Create tests for all ResponseTransformer methods
  - Add tests for query parameter parsing and validation
  - Write tests for sensitive field exclusion and state consistency
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 6.1, 6.2, 6.3, 6.4_

- [ ] 11.2 Write integration tests for authentication and password management

  - Create tests for token invalidation after logout and password changes
  - Add tests for password change and reset flows
  - Write tests for Aadhaar verification integration
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 12.1, 12.2, 12.3, 12.4, 12.5, 13.1, 13.2, 13.3, 13.4, 13.5_

- [ ] 11.3 Create end-to-end API response validation tests

  - Write tests to validate complete API response structures
  - Add tests for query parameter combinations and edge cases
  - Create tests to verify backward compatibility where applicable
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 12. Update documentation and finalize implementation
- [ ] 12.1 Update API documentation with new response structures

  - Update Swagger/OpenAPI specifications with new response models
  - Add documentation for query parameters and their effects
  - Create examples showing optimized response structures
  - _Requirements: 9.1, 9.2, 9.3, 9.4_

- [ ] 12.2 Create migration guide for API consumers
  - Document changes in response structures and field naming
  - Provide migration examples for common use cases
  - Add troubleshooting guide for response validation issues
  - _Requirements: 9.1, 9.2, 9.3_
