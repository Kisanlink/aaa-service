package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMiddleware provides authentication and authorization middleware
type AuthMiddleware struct {
	authService  *services.AuthService
	authzService *services.AuthorizationService
	auditService *services.AuditService
	logger       *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	authService *services.AuthService,
	authzService *services.AuthorizationService,
	auditService *services.AuditService,
	logger *zap.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService:  authService,
		authzService: authzService,
		auditService: auditService,
		logger:       logger,
	}
}

// HTTPAuthMiddleware provides HTTP authentication middleware
func (m *AuthMiddleware) HTTPAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for health checks and public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			m.logger.Debug("Skipping authentication for public endpoint", zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.logger.Warn("Invalid authorization header format",
				zap.String("path", c.Request.URL.Path),
				zap.String("header", authHeader))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		m.logger.Debug("Validating token", zap.String("token_prefix", token[:10]+"..."))

		// Validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Warn("Invalid token",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))

			// Audit failed authentication
			if m.auditService != nil {
				m.auditService.LogAPIAccess(c.Request.Context(), "unknown", c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.GetHeader("User-Agent"), false, err)
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user context in Gin context (this is what RequirePermission checks)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_validated", claims.IsValidated)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("token_type", claims.TokenType)

		// Add to request context for downstream services
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "ip_address", c.ClientIP())
		ctx = context.WithValue(ctx, "user_agent", c.GetHeader("User-Agent"))
		c.Request = c.Request.WithContext(ctx)

		m.logger.Debug("User authenticated and context set",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.Request.URL.Path))

		c.Next()

		// Audit successful access
		if m.auditService != nil {
			statusCode := c.Writer.Status()
			success := statusCode >= 200 && statusCode < 400

			if !success {
				// Log failed request
				m.auditService.LogAPIAccess(c.Request.Context(), claims.UserID, c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.GetHeader("User-Agent"), false, fmt.Errorf("HTTP %d", statusCode))
			} else {
				// Log successful request
				m.auditService.LogAPIAccess(c.Request.Context(), claims.UserID, c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.GetHeader("User-Agent"), true, nil)
			}
		}
	}
}

// HTTPAuthzMiddleware provides HTTP authorization middleware
func (m *AuthMiddleware) HTTPAuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authorization for public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Error("User ID not found in context for authorization check")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Authentication context not found",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Invalid user ID type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid authentication context",
			})
			c.Abort()
			return
		}

		// Check API endpoint access
		allowed, err := m.authzService.ValidateAPIEndpointAccess(c.Request.Context(), userIDStr, c.Request.Method, c.Request.URL.Path)
		if err != nil {
			m.logger.Error("Authorization check failed",
				zap.String("user_id", userIDStr),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Authorization check failed",
			})
			c.Abort()
			return
		}

		if !allowed.Allowed {
			m.logger.Warn("Access denied",
				zap.String("user_id", userIDStr),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path))

			// Audit access denied
			if m.auditService != nil {
				m.auditService.LogAccessDenied(c.Request.Context(), userIDStr, c.Request.Method, "api", c.Request.URL.Path, "insufficient permissions")
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions to access this resource",
			})
			c.Abort()
			return
		}

		m.logger.Debug("Access authorized",
			zap.String("user_id", userIDStr),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// GRPCAuthInterceptor provides gRPC authentication interceptor
func (m *AuthMiddleware) GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for public methods
		if m.isPublicGRPCMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			m.logger.Warn("Missing metadata in gRPC request", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		// Extract authorization token
		authHeaders, ok := md["authorization"]
		if !ok || len(authHeaders) == 0 {
			m.logger.Warn("Missing authorization token in gRPC request", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		token := authHeaders[0]

		// Remove Bearer prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Warn("Invalid token in gRPC request",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add user information to context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "is_validated", claims.IsValidated)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		m.logger.Debug("gRPC user authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("method", info.FullMethod))

		return handler(ctx, req)
	}
}

// RequireRole creates a middleware that requires specific roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context
		rolesInterface, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Role information not available",
			})
			c.Abort()
			return
		}

		userRoles, ok := rolesInterface.([]*services.TokenClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid role information",
			})
			c.Abort()
			return
		}

		// Extract role names from user roles
		userRoleNames := make(map[string]bool)
		for _, role := range userRoles {
			userRoleNames[role.UserID] = true // This would need proper role name extraction
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRoleNames[requiredRole] {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			userID, _ := c.Get("user_id")
			m.logger.Warn("Access denied - insufficient roles",
				zap.String("user_id", fmt.Sprintf("%v", userID)),
				zap.Strings("required_roles", requiredRoles))

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient role permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates a middleware that requires specific permissions
func (m *AuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Error("User ID not found in context for permission check",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "User context not found - authentication may have failed",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Invalid user ID type in context",
				zap.String("user_id_type", fmt.Sprintf("%T", userID)),
				zap.String("user_id_value", fmt.Sprintf("%v", userID)))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Invalid user context type",
			})
			c.Abort()
			return
		}

		m.logger.Debug("Checking permission",
			zap.String("user_id", userIDStr),
			zap.String("resource", resource),
			zap.String("action", action))

		// Check specific permission
		permission := &services.Permission{
			UserID:     userIDStr,
			Resource:   resource,
			ResourceID: resource, // Use resource type as resource ID for general permissions
			Action:     action,
		}

		result, err := m.authzService.CheckPermission(c.Request.Context(), permission)
		if err != nil {
			m.logger.Error("Permission check failed",
				zap.String("user_id", userIDStr),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.Error(err))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Permission check failed",
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			m.logger.Warn("Access denied - insufficient permissions",
				zap.String("user_id", userIDStr),
				zap.String("resource", resource),
				zap.String("action", action))

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": fmt.Sprintf("Insufficient permissions for %s:%s", resource, action),
			})
			c.Abort()
			return
		}

		m.logger.Debug("Permission granted",
			zap.String("user_id", userIDStr),
			zap.String("resource", resource),
			zap.String("action", action))

		c.Next()
	}
}

// isPublicEndpoint checks if an endpoint is public (doesn't require authentication)
func (m *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/docs",
		"/api/v2/auth/login",
		"/api/v2/auth/register",
		"/api/v2/auth/refresh",
		"/",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

// isPublicGRPCMethod checks if a gRPC method is public (doesn't require authentication)
func (m *AuthMiddleware) isPublicGRPCMethod(method string) bool {
	publicMethods := []string{
		"/pb.UserServiceV2/Login",
		"/pb.UserServiceV2/Register",
		"/pb.UserServiceV2/RefreshToken",
		"/grpc.health.v1.Health/Check",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}

	return false
}
