# Role Inheritance Code Locations

## Absolute File Paths

### 1. Role Models & Repositories

#### GroupRole Model
- **Path:** `/Users/kaushik/aaa-service/internal/entities/models/group_role.go`
- **Key:** GroupRole struct definition
- **Lines:** 32-48 (Model), 56-165 (Methods)

#### GroupRole Repository
- **Path:** `/Users/kaushik/aaa-service/internal/repositories/groups/group_role_repository.go`
- **Key Methods:**
  - `GetByGroupID()` - Line ~70
  - `Create()` - Full CRUD
  - Tests: `/Users/kaushik/aaa-service/internal/repositories/groups/group_role_repository_test.go`

### 2. Role Inheritance Engine (The Production-Ready Implementation)

#### Main Engine
- **Path:** `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_engine.go`
- **Size:** ~650 lines
- **Key Methods:**
  - `CalculateEffectiveRoles()` - Line 146
  - `calculateBottomUpRoles()` - Line 268
  - `VerifyBottomUpInheritance()` - Line 487
  - Data structures: `EffectiveRole` (Line 113), `InheritanceVerificationResult` (Line 620)

#### Engine Tests
- `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_engine_test.go`
- `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_bottom_up_test.go`
- `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_comprehensive_test.go`

### 3. Current Token Generation (Direct Roles Only)

#### Auth Handler
- **Path:** `/Users/kaushik/aaa-service/internal/handlers/auth/auth_handler.go`
- **Method:** `LoginV2()` - Lines 56-240
- **Role Fetching:** Lines 151-157
  - Gets user response with roles
  - Only includes direct assignments
- **Token Generation:** Lines 199-208
  - Calls `helper.GenerateAccessTokenWithContext(userRoles, ...)`

#### Token Refresh Handler
- **Path:** `/Users/kaushik/aaa-service/internal/handlers/auth/auth_handler.go`
- **Method:** `RefreshTokenV2()` - Lines 330-460
- **Similar pattern:** Lines 371-385 (role fetching)

### 4. User Service (Where Roles Are Fetched)

#### User Service - Constructor
- **Path:** `/Users/kaushik/aaa-service/internal/services/user/service.go`
- **Constructor:** `NewService()` - Lines 22-38
- **Missing:** RoleInheritanceEngine injection

#### User Service - Role Fetching
- **Path:** `/Users/kaushik/aaa-service/internal/services/user/additional_methods.go`
- **Method:** `VerifyUserCredentials()` - Line ~558
  - Calls `GetUserWithRoles()` at end
- **Method:** `GetUserWithRoles()` - Lines 245-316
  - Gets user
  - Calls `getCachedUserRoles()` - Line 269
  - Converts to response
  - **Returns only direct roles**
- **Method:** `getCachedUserRoles()` - Line ~552
  - Calls `userRoleRepo.GetActiveRolesByUserID()`
  - **Only direct roles**

### 5. User Role Repository (Direct Roles Only)

#### User Role Repository
- **Path:** `/Users/kaushik/aaa-service/internal/repositories/roles/user_role_repository.go`
- **Method:** `GetActiveRolesByUserID()` - Lines 170-196
  - Returns roles with preloaded role details
  - **Only direct user-role assignments**
- **Method:** `GetByUserID()` - Lines 82-89
  - Direct assignments only
- **Method:** `GetByUserAndRole()` - Lines 102-119
  - Specific assignment lookup

### 6. Interfaces & Service Definitions

#### UserService Interface
- **Path:** `/Users/kaushik/aaa-service/internal/interfaces/interfaces.go`
- **Lines:** 59-86
- **Method:** `GetUserWithRoles()` - Line 75
  - Exists in interface
  - Currently returns only direct roles
- **Missing:** `GetUserEffectiveRoles()` method

#### GroupService Interface
- **Path:** `/Users/kaushik/aaa-service/internal/interfaces/interfaces.go`
- **Lines:** 112-128
- **Method:** `GetUserEffectiveRoles()` - Line 127
  - **Exists in GroupService interface!**
  - But it returns `interface{}`
  - Consider moving to UserService or making more specific

### 7. Group Service Implementation

#### Group Service
- **Path:** `/Users/kaushik/aaa-service/internal/services/groups/group_service.go`
- **Key Methods:**
  - `AssignRoleToGroup()` - Line ~400+
  - `RemoveRoleFromGroup()` - Line ~500+
  - `GetGroupRoles()` - Line ~600+
  - `GetUserEffectiveRoles()` - May exist (check implementation)

## Direct vs Inherited Roles Comparison

### Direct Roles Location Chain
```
auth_handler.go:LoginV2()
  ↓
user_service.VerifyUserCredentials()
  ↓
user_service.GetUserWithRoles()
  ↓
getCachedUserRoles()
  ↓
user_role_repo.GetActiveRolesByUserID()
  ↓
Models with DIRECT ROLES ONLY
```

### Inheritance Engine Location (Unused)
```
role_inheritance_engine.go
  ↓
CalculateEffectiveRoles(orgID, userID)
  ↓
getCachedUserRoles()
  ↓
roleInheritanceEngine.calculateBottomUpRoles()
  ↓
Recursive traversal of group hierarchy
  ↓
Distance-based conflict resolution
  ↓
Returns EFFECTIVE ROLES (direct + inherited)
```

## Test Coverage Map

### GroupRole Tests
- `/Users/kaushik/aaa-service/internal/entities/models/group_role_test.go`
- `/Users/kaushik/aaa-service/internal/entities/models/group_role_integration_test.go`
- `/Users/kaushik/aaa-service/internal/repositories/groups/group_role_repository_test.go`

### RoleInheritanceEngine Tests
- **Test 1:** `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_engine_test.go`
  - Core engine tests
- **Test 2:** `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_bottom_up_test.go`
  - Bottom-up inheritance verification
- **Test 3:** `/Users/kaushik/aaa-service/internal/services/groups/role_inheritance_comprehensive_test.go`
  - Comprehensive scenario testing

### Group Service Tests
- `/Users/kaushik/aaa-service/internal/services/groups/group_service_role_test.go`
- `/Users/kaushik/aaa-service/internal/handlers/organizations/role_group_assignment_test.go`

## Database Tables

### group_roles table
- **Generated by:** GroupRole model
- **Schema:** Supports GroupID, RoleID, OrganizationID, AssignedBy, StartsAt, EndsAt, IsActive
- **Relationships:** Groups, Roles, Organizations, Users

### users table
- **Has:** Direct roles through UserRole join

### group_memberships table
- **Has:** User to Group relationships
- **Used by:** RoleInheritanceEngine to find user's groups

## Key Constants & Endpoints

### API Endpoints (from routes)
```
POST   /organizations/{orgID}/groups/{groupID}/roles
DELETE /organizations/{orgID}/groups/{groupID}/roles/{roleID}
GET    /organizations/{orgID}/groups/{groupID}/roles
```

### Cache Keys
- Role calculation: `org:{orgID}:user:{userID}:effective_roles`
- Direct groups: `org:{orgID}:user:{userID}:groups`
- TTL: 5 minutes (300 seconds)

## Key Decision Points

### 1. Where to Integrate Engine
Options:
a) In UserService.GetUserWithRoles() (recommended)
b) In auth handler's LoginV2() (alternative)
c) In both (redundant)

### 2. How to Merge Roles
- Use EffectiveRole.Distance for precedence
- Direct (0) > Child (1) > Grandchild (2)
- Deduplicate by role ID

### 3. Interface Changes
- Add method to UserService interface
- Or use existing GroupService.GetUserEffectiveRoles()

## Summary of Changes Needed

**Files to modify:** 4 main files
1. `/Users/kaushik/aaa-service/internal/services/user/service.go` - Add engine injection
2. `/Users/kaushik/aaa-service/internal/services/user/additional_methods.go` - Use engine
3. `/Users/kaushik/aaa-service/internal/handlers/auth/auth_handler.go` - Optional
4. `/Users/kaushik/aaa-service/internal/interfaces/interfaces.go` - Optional

**New code required:** ~100-200 lines (mostly in additional_methods.go)
