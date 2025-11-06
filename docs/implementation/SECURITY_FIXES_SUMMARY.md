# AAA Service Security Fixes Summary

## Issues Fixed

### 1. Deleted Users Appearing in User Lists

**Problem**: The `ListUsers` endpoint was returning deleted users in the response.

**Root Cause**: The user service and repository were not filtering out soft-deleted users (where `deleted_at` is not null).

**Fix Applied**:

- Updated `ListUsers` method in `internal/services/user/read.go` to filter out deleted users at service level
- Updated `List` method in `internal/repositories/users/user_repository.go` to add `WhereNull("deleted_at")` filter
- Updated `ListAll` method to also filter out deleted users
- Added additional service-level filtering as defense in depth

### 2. Deleted Users Able to Login

**Problem**: Deleted users could still authenticate and receive access tokens.

**Root Cause**: The authentication service was not checking the `deleted_at` field during login.

**Fix Applied**:

- Added security checks in `Login` method in `internal/services/auth_service.go`
- Added security checks in `LoginWithUsername` method
- Added security checks in `RefreshToken` method
- Updated repository methods `GetByUsername` and `GetByPhoneNumber` to filter deleted users
- All authentication flows now reject deleted users with "invalid credentials" error

### 3. Set MPIN Working Without Password Validation

**Problem**: Users could set MPIN without providing their current password for verification.

**Root Cause**: The `SetMPin` method was not requiring password validation.

**Fix Applied**:

- Updated `SetMPin` method signature in `internal/interfaces/interfaces.go` to require `currentPassword` parameter
- Updated `SetMPin` implementation in `internal/services/auth_service.go` to verify current password
- Updated `SetMPin` implementation in `internal/services/user/additional_methods.go` to verify current password
- Updated auth handler in `internal/handlers/auth/auth_handler.go` to pass password from request
- The `SetMPinRequest` already included password field, so no request structure changes needed

### 4. Users Can Set MPIN Multiple Times

**Problem**: Users could set MPIN repeatedly without proper workflow control.

**Root Cause**: No check was performed to see if MPIN was already set.

**Fix Applied**:

- Added check in `SetMPin` method to return conflict error if MPIN is already set
- Created separate `UpdateMPin` method for updating existing MPIN
- `UpdateMPin` requires current MPIN verification instead of password
- Added proper error handling for conflict scenarios in auth handler

## Additional Security Enhancements

### Defense in Depth

- Added multiple layers of deleted user filtering (repository, service, and auth levels)
- All user lookup methods now filter out deleted users by default
- Added security logging for suspicious activities (deleted user access attempts)

### Proper Error Handling

- Consistent error responses that don't leak information about deleted users
- Proper HTTP status codes (401 for authentication failures, 409 for conflicts)
- Detailed audit logging for security events

### Method Separation

- Clear separation between `SetMPin` (first-time setup with password) and `UpdateMPin` (change existing with current MPIN)
- Proper validation and error messages for each scenario

## Testing Recommendations

1. **Test Deleted User Filtering**:

   - Create a user, delete them, verify they don't appear in user lists
   - Attempt login with deleted user credentials, verify it fails
   - Attempt token refresh for deleted user, verify it fails

2. **Test MPIN Security**:

   - Attempt to set MPIN without password, verify it fails
   - Set MPIN with correct password, verify success
   - Attempt to set MPIN again, verify conflict error
   - Use update-mpin endpoint to change existing MPIN

3. **Test Authentication Flows**:
   - Verify all login methods reject deleted users
   - Verify token refresh rejects deleted users
   - Verify proper error messages don't leak user existence

## Files Modified

1. `internal/services/user/read.go` - Updated ListUsers and GetUserByID
2. `internal/services/auth_service.go` - Added deleted user checks in auth methods, updated SetMPin
3. `internal/repositories/users/user_repository.go` - Added deleted_at filtering to all query methods
4. `internal/handlers/auth/auth_handler.go` - Updated SetMPin handler to pass password
5. `internal/services/user/additional_methods.go` - Updated SetMPin and added UpdateMPin methods
6. `internal/interfaces/interfaces.go` - Updated UserService interface

## Security Impact

These fixes address critical security vulnerabilities that could have allowed:

- Information disclosure through deleted user data
- Unauthorized access by deleted users
- Weak authentication controls for MPIN setup

All fixes maintain backward compatibility while significantly improving security posture.
