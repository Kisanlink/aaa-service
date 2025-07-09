# Using kisanlink-db Transactions Directly

## Simple Approach

The AAA service directly imports and uses transaction functions from `kisanlink-db` when needed, without creating custom transaction wrappers.

## Usage Examples

### Example 1: Direct Transaction Usage in Service

```go
import (
    "github.com/Kisanlink/kisanlink-db/pkg/db"
)

func (s *UserService) CreateUserWithRole(ctx context.Context, userReq *CreateUserRequest, roleID string) error {
    // Get the database manager (passed from HTTP handler or injected)
    dbManager := s.getDBManager() // or however you access it
    
    // Use kisanlink-db transaction directly
    if pgManager, ok := dbManager.(*db.PostgresManager); ok {
        return pgManager.WithTransaction(ctx, func(tx *gorm.DB) error {
            // Create user within transaction
            user := &models.User{...}
            if err := s.userRepo.Create(ctx, user); err != nil {
                return err
            }
            
            // Assign role within same transaction
            userRole := &models.UserRole{...}
            if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
                return err
            }
            
            return nil
        })
    }
    
    // Fallback for non-PostgreSQL databases
    return s.createUserWithRoleNonTransactional(ctx, userReq, roleID)
}
```

### Example 2: Direct Read-Only Transaction

```go
func (s *UserService) GetUserWithCompleteData(ctx context.Context, userID string) (*CompleteUserData, error) {
    dbManager := s.getDBManager()
    
    if pgManager, ok := dbManager.(*db.PostgresManager); ok {
        var result *CompleteUserData
        
        err := pgManager.WithReadOnly(ctx, func(tx *gorm.DB) error {
            // Multiple reads within read-only transaction
            user, err := s.userRepo.GetByID(ctx, userID)
            if err != nil {
                return err
            }
            
            roles, err := s.userRoleRepo.GetByUserID(ctx, userID)
            if err != nil {
                return err
            }
            
            result = &CompleteUserData{User: user, Roles: roles}
            return nil
        })
        
        return result, err
    }
    
    // Fallback for non-PostgreSQL databases
    return s.getUserDataNonTransactional(ctx, userID)
}
```

## Key Points

1. **Direct Import**: Import `github.com/Kisanlink/kisanlink-db/pkg/db` and use transaction methods directly
2. **No Custom Wrappers**: Don't create service-level transaction wrappers
3. **Type Assertion**: Check if database manager supports transactions with type assertion
4. **Fallback**: Provide fallback for non-PostgreSQL database managers
5. **Context**: Always pass context through transaction operations

This keeps the AAA service simple while leveraging kisanlink-db transaction capabilities when needed. 