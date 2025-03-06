package middleware

import (
	"context"
	"log"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/controller/user"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/repositories"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func AuthInterceptor(db *gorm.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeaders, ok := md["authorization"]
	if !ok || len(authHeaders) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	token := authHeaders[0]
	userID, err := helper.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Add user ID to context for downstream handlers
	ctx = context.WithValue(ctx, "user_id", userID)
	userRepo := repositories.NewUserRepository(db) // Initialize your UserRepository
	createdUser, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to fetch user details: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user details")
	}

	// Fetch user roles and permissions
	roles, permissions, err := userRepo.FindUserRolesAndPermissions(ctx, createdUser.ID)
	if err != nil {
		log.Printf("Failed to fetch user roles and permissions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions")
	}


	// Check user permissions
	results, err := client.CheckUserPermissions(createdUser.Username,user.LowerCaseSlice(roles), 
	user.LowerCaseSlice(permissions),)
	if err != nil {
		log.Printf("Failed to check permissions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check permissions")
	}


	// Check if the user has all required permissions
	for permission, hasPermission := range results {
		if !hasPermission {
			log.Printf("User %s does not have permission: %s", userID, permission)
			return nil, status.Errorf(codes.PermissionDenied, "user does not have permission: %s", permission)
		}
	}
	return handler(ctx, req)
}
}