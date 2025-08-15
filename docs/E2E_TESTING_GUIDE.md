# End-to-End Testing Guide for AAA Service

This document describes the comprehensive end-to-end testing framework for the AAA (Authentication, Authorization, and Accounting) service, designed to verify the complete workflow from user registration to multi-service module management.

## Overview

The AAA service provides a comprehensive authorization platform that supports:
- **User registration** with multiple role types (normal user, admin, super_admin, etc.)
- **Module registration** for external services (farmers-module, marketplace-module, etc.)
- **Role-based access control** with fine-grained permissions
- **Resource and action management** across multiple services
- **End-to-end authorization flows**

## Test Architecture

### Test Structure

```
test/integration/
├── e2e_aaa_core_test.go        # Core AAA functionality tests
└── e2e_farmers_module_test.go  # Module registration and multi-service tests
```

### Test Suite Components

The `TestSuite` struct provides a complete testing environment:

```go
type TestSuite struct {
    DB           *gorm.DB              // In-memory SQLite database
    Router       *gin.Engine           // HTTP router with all routes
    Logger       *zap.Logger           // Structured logging
    AuthService  *services.AuthService // Authentication service
    UserService  *services.UserService // User management service
    RoleService  *services.RoleService // Role management service
    Validator    *validator.Validator  // Request validation
    AdminToken   string               // Admin user token
    SuperToken   string               // Super admin token
    UserToken    string               // Regular user token
}
```

## Test Categories

### 1. Core AAA Functionality (`TestE2E_AAA_CoreWorkflow`)

This test verifies the fundamental AAA service capabilities:

#### **1.1 Service Initialization**
- Health check endpoint
- Static actions seeding (CRUD operations, admin actions, etc.)
- Default roles creation (super_admin, admin, user, viewer, aaa_admin, module_admin)

#### **1.2 User Registration Flow**
Tests user registration with different role types:

```go
// Super Admin User
superAdminUser := registerUser("+1234567890", "+1", "SuperSecure123!", []string{superAdminRoleID})

// Admin User
adminUser := registerUser("+1234567891", "+1", "AdminSecure123!", []string{adminRoleID})

// Regular User
regularUser := registerUser("+1234567892", "+1", "UserSecure123!", []string{userRoleID})
```

#### **1.3 Authentication and Authorization**
- Token generation and validation
- Role-based access control
- Protected endpoint access
- Invalid token handling

#### **1.4 Verification Points**
- ✅ Static actions seeded (>10 actions including create, read, update, delete, manage, admin)
- ✅ Default roles created (6 essential roles)
- ✅ Users registered with correct role assignments
- ✅ Authentication tokens generated and validated
- ✅ Role bindings created in database
- ✅ Access control enforced on protected endpoints

### 2. Module Registration (`TestE2E_FarmersModuleWorkflow`)

This test demonstrates the complete module registration and management workflow:

#### **2.1 Farmers Module Registration**
Comprehensive module with complete agricultural management capabilities:

```go
farmersModule := &ModuleRegistrationRequest{
    ServiceName: "farmers-module",
    Description: "Comprehensive farmers management service",
    Actions: []ModuleActionRequest{
        // 15 actions covering farmers, farms, crops, inventory, finance, weather
        {Name: "create_farmer", Description: "Create farmer profile", Category: "farmers"},
        {Name: "manage_farm", Description: "Manage farm operations", Category: "farms"},
        {Name: "view_crop_data", Description: "View crop analytics", Category: "crops"},
        {Name: "financial_tracking", Description: "Track farm finances", Category: "finance"},
        // ... and 11 more actions
    },
    Resources: []ModuleResourceRequest{
        // 7 resources covering databases, APIs, and file storage
        {Name: "farmers_database", ResourceType: "aaa/table"},
        {Name: "crops_database", ResourceType: "aaa/table"},
        {Name: "farmers_api", ResourceType: "aaa/api"},
        // ... and 4 more resources
    },
    Roles: []ModuleRoleRequest{
        // 5 specialized roles with different permission levels
        {Name: "farmer", Permissions: ["view_farmer", "update_farmer", "view_crop_data"]},
        {Name: "farm_manager", Permissions: ["create_farmer", "manage_farm", "financial_tracking"]},
        {Name: "crop_specialist", Permissions: ["create_crop_plan", "generate_crop_reports"]},
        {Name: "agricultural_analyst", Permissions: ["view_crop_data", "generate_crop_reports"]},
        {Name: "farmers_admin", Permissions: ["all_farmers_actions"]},
    },
}
```

#### **2.2 Multi-Role User Registration**
Tests user registration with module-specific roles:

```go
// Farmer with basic access
farmerUser := registerUser("+1234567900", "FarmerSecure123!", []string{farmerRoleID})

// Farm Manager with operational access
managerUser := registerUser("+1234567901", "ManagerSecure123!", []string{farmManagerRoleID})

// Crop Specialist with analytical access
specialistUser := registerUser("+1234567902", "SpecialistSecure123!", []string{cropSpecialistRoleID})
```

#### **2.3 Verification Points**
- ✅ Module service created in database
- ✅ 15 module-specific actions created
- ✅ 5 specialized roles created with appropriate permissions
- ✅ 7 resources created (databases, APIs, files)
- ✅ 5 explicit permissions created linking actions to resources
- ✅ Users registered with module roles
- ✅ Role-permission associations established
- ✅ Module information retrieval working

### 3. Multi-Module Integration (`TestE2E_MultipleModulesWorkflow`)

This test verifies the system's ability to handle multiple services:

#### **3.1 Multiple Module Registration**
```go
// Marketplace Module
marketplaceModule := registerModule("marketplace-module", marketplaceActions, marketplaceRoles)

// Logistics Module
logisticsModule := registerModule("logistics-module", logisticsActions, logisticsRoles)
```

#### **3.2 Cross-Module User Assignment**
```go
// User with roles from multiple modules
multiRoleUser := registerUser("+1234567903", []string{sellerRoleID, driverRoleID})
```

#### **3.3 Verification Points**
- ✅ Multiple modules registered successfully
- ✅ Module isolation maintained
- ✅ Cross-module role assignment supported
- ✅ Module listing and information retrieval

## Default Roles and Permissions

### Core AAA Roles

| Role | Scope | Description | Use Case |
|------|-------|-------------|----------|
| `super_admin` | GLOBAL | Super Administrator with global access | System administration |
| `admin` | ORG | Administrator with organization-level access | Organization management |
| `user` | ORG | Regular user with basic access | Standard application usage |
| `viewer` | ORG | Read-only access user | Reporting and monitoring |
| `aaa_admin` | GLOBAL | AAA service administrator | Service configuration |
| `module_admin` | ORG | Module administrator for service management | Module lifecycle management |

### Farmers Module Roles

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| `farmer` | Individual farmer | view_farmer, update_farmer, view_crop_data |
| `farm_manager` | Operations manager | create_farmer, manage_farm, financial_tracking |
| `crop_specialist` | Agricultural expert | create_crop_plan, generate_crop_reports |
| `agricultural_analyst` | Data analyst | view_crop_data, generate_crop_reports, financial_tracking |
| `farmers_admin` | Module administrator | All farmers module permissions |

## Static Actions

The system includes 20+ built-in static actions:

### CRUD Operations
- `create`, `read`, `view`, `update`, `edit`, `delete`, `list`

### Administrative Actions
- `manage`, `admin`, `configure`, `monitor`

### Security Actions
- `assign_roles`, `remove_roles`, `grant_permissions`, `revoke_permissions`

### Audit Actions
- `audit`, `log`, `inspect`

### Workflow Actions
- `approve`, `reject`, `submit`, `cancel`

### Special Actions
- `impersonate`, `bypass`, `override` (high-privilege operations)

## Running the Tests

### Prerequisites
```bash
# Ensure all dependencies are installed
go mod download

# Verify the service builds
go build ./...
```

### Test Commands

```bash
# Run all end-to-end tests
make test-e2e

# Run tests with verbose output
make test-e2e-verbose

# Run only farmers module tests
make test-farmers

# Run with specific timeout
go test -v ./test/integration/... -timeout 60s

# Run with race detection
go test -v -race ./test/integration/...
```

### Test Output Example

```
=== RUN   TestE2E_AAA_CoreWorkflow
=== RUN   TestE2E_AAA_CoreWorkflow/1._Health_Check
=== RUN   TestE2E_AAA_CoreWorkflow/2._Verify_Static_Actions_Seeded
=== RUN   TestE2E_AAA_CoreWorkflow/3._Verify_Default_Roles_Created
=== RUN   TestE2E_AAA_CoreWorkflow/4._Register_Super_Admin_User
=== RUN   TestE2E_AAA_CoreWorkflow/5._Register_Admin_User
=== RUN   TestE2E_AAA_CoreWorkflow/6._Register_Regular_User
=== RUN   TestE2E_AAA_CoreWorkflow/7._Login_Users_and_Get_Tokens
=== RUN   TestE2E_AAA_CoreWorkflow/8._Verify_User_Role_Assignments
=== RUN   TestE2E_AAA_CoreWorkflow/9._Test_Access_Control
=== RUN   TestE2E_AAA_CoreWorkflow/10._Test_Token_Validation
--- PASS: TestE2E_AAA_CoreWorkflow (2.34s)

=== RUN   TestE2E_FarmersModuleWorkflow
=== RUN   TestE2E_FarmersModuleWorkflow/1._Register_Farmers_Module
=== RUN   TestE2E_FarmersModuleWorkflow/2._Verify_Module_Registration_in_Database
=== RUN   TestE2E_FarmersModuleWorkflow/3._Register_Users_with_Farmers_Module_Roles
=== RUN   TestE2E_FarmersModuleWorkflow/4._Test_Module_Role_Permissions
=== RUN   TestE2E_FarmersModuleWorkflow/5._Test_Module_Information_Retrieval
=== RUN   TestE2E_FarmersModuleWorkflow/6._Test_End-to-End_Authorization_Flow
--- PASS: TestE2E_FarmersModuleWorkflow (1.87s)
```

## Test Database Schema

The tests use an in-memory SQLite database with the complete AAA schema:

```go
db.AutoMigrate(
    &models.User{}, &models.UserProfile{}, &models.Role{}, &models.Permission{},
    &models.Action{}, &models.Organization{}, &models.Group{}, &models.GroupMembership{},
    &models.Resource{}, &models.Binding{}, &models.Event{}, &models.EventCheckpoint{},
    &models.Attribute{}, &models.ColumnGroup{}, &models.ColumnGroupMember{},
    &models.Principal{}, &models.Service{},
)
```

## Integration with CI/CD

### GitHub Actions Integration
```yaml
- name: Run E2E Tests
  run: |
    make test-e2e

- name: Run Farmers Module Tests
  run: |
    make test-farmers
```

### Test Coverage
The E2E tests provide coverage for:
- ✅ **User Management**: Registration, authentication, role assignment
- ✅ **Role-Based Access Control**: Permission verification, access enforcement
- ✅ **Module Registration**: Service onboarding, action/resource creation
- ✅ **Multi-Service Integration**: Cross-module permissions, role composition
- ✅ **Database Integrity**: Proper relationships, constraints, migrations
- ✅ **API Endpoints**: HTTP request/response handling, error cases
- ✅ **Authorization Flow**: End-to-end permission checking

## Troubleshooting

### Common Issues

1. **Test Timeout**: Increase timeout for slow systems
   ```bash
   go test -v ./test/integration/... -timeout 120s
   ```

2. **Database Connection**: Ensure SQLite driver is available
   ```bash
   go get gorm.io/driver/sqlite
   ```

3. **Missing Dependencies**: Update modules
   ```bash
   go mod tidy
   ```

### Debug Mode
Enable verbose logging in tests:
```go
logger, _ := zap.NewDevelopment()
```

## Future Enhancements

1. **Performance Testing**: Add load testing for high-volume scenarios
2. **Security Testing**: Penetration testing for authorization bypass
3. **Integration Testing**: Real database and SpiceDB integration
4. **API Testing**: Comprehensive HTTP API testing with different clients
5. **Mobile Testing**: Mobile application integration scenarios

## Summary

The comprehensive end-to-end testing framework ensures that the AAA service works correctly from basic user registration through complex multi-service module management. The tests verify:

- ✅ **Core functionality**: User registration, authentication, role assignment
- ✅ **Module system**: Dynamic service registration with roles and permissions
- ✅ **Authorization**: Role-based access control across services
- ✅ **Scalability**: Multiple modules and cross-module user assignment
- ✅ **Data integrity**: Proper database relationships and constraints
- ✅ **API compatibility**: HTTP endpoints and error handling

This testing framework provides confidence that the AAA service can handle real-world scenarios involving complex agricultural systems, marketplaces, logistics, and other domain-specific modules while maintaining security and performance standards.
