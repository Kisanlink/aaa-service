package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMiddleware provides authentication and authorization middleware
type AuthMiddleware struct {
	authService       *services.AuthService
	authzService      *services.AuthorizationService
	auditService      *services.AuditService
	serviceRepository ServiceRepository
	logger            *zap.Logger
	jwtVerifier       JWTVerifier
	jwtCfg            *config.JWTConfig
}

// ServiceRepository defines methods for service authentication (imported from interfaces package)
type ServiceRepository interface {
	GetByAPIKey(ctx context.Context, apiKeyHash string) (*models.Service, error)
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	authService *services.AuthService,
	authzService *services.AuthorizationService,
	auditService *services.AuditService,
	serviceRepository ServiceRepository,
	logger *zap.Logger,
	jwtVerifier JWTVerifier,
	jwtCfg *config.JWTConfig,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService:       authService,
		authzService:      authzService,
		auditService:      auditService,
		serviceRepository: serviceRepository,
		logger:            logger,
		jwtVerifier:       jwtVerifier,
		jwtCfg:            jwtCfg,
	}
}

// HTTPAuthMiddleware provides HTTP authentication middleware
func (m *AuthMiddleware) HTTPAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			m.logger.Debug("Skipping authentication for public endpoint", zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}

		// Extract token from header or cookie
		authz := c.GetHeader("Authorization")
		token := ""
		trimmed := strings.TrimSpace(authz)
		if strings.HasPrefix(trimmed, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(trimmed, "Bearer "))
		} else if ck, err := c.Request.Cookie("auth_token"); err == nil {
			token = strings.TrimSpace(ck.Value)
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication required",
				"message": "login required",
			})
			return
		}

		// Verify token with centralized config
		claims, err := m.jwtVerifier.Verify(token, m.jwtCfg)
		if err != nil {
			m.logger.Warn("JWT verify failed",
				zap.String("path", c.Request.URL.Path),
				zap.String("reason", err.Error()),
			)
			m.logger.Warn("JWT cfg",
				zap.String("issuer", m.jwtCfg.Issuer),
				zap.String("aud", m.jwtCfg.Audience),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Set user context used by downstream handlers
		c.Set("user_id", claims.Sub)
		m.logger.Debug("Authenticated user context set", zap.String("user_id", claims.Sub), zap.String("path", c.Request.URL.Path))
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.Sub)
		ctx = context.WithValue(ctx, "ip_address", c.ClientIP())
		ctx = context.WithValue(ctx, "user_agent", c.GetHeader("User-Agent"))
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// JWTClaims captures the verified JWT data
type JWTClaims struct {
	Sub string
	Iss string
	Aud string
	Exp time.Time
	Nbf time.Time
	Iat time.Time
	Raw map[string]any
}

// JWTVerifier defines Verify behavior
type JWTVerifier interface {
	Verify(token string, cfg *config.JWTConfig) (*JWTClaims, error)
}

// HTTPAuthzMiddleware provides HTTP authorization middleware
func (m *AuthMiddleware) HTTPAuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authorization for public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Allow authenticated users to call logout without additional authorization
		if c.Request.URL.Path == "/api/v1/auth/logout" {
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

		if !allowed {
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

		// Check for API key authentication (for service-to-service calls)
		if apiKeys, ok := md["x-api-key"]; ok && len(apiKeys) > 0 {
			return m.authenticateService(ctx, apiKeys[0], info.FullMethod, handler, req)
		}

		// Fall back to JWT token authentication (for user calls)
		authHeaders, ok := md["authorization"]
		if !ok || len(authHeaders) == 0 {
			m.logger.Warn("Missing authorization token or API key in gRPC request", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "authorization token or API key is required")
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

// authenticateService validates API key and sets service context
func (m *AuthMiddleware) authenticateService(ctx context.Context, apiKey, method string, handler grpc.UnaryHandler, req interface{}) (interface{}, error) {
	// Hash the API key to compare with stored hash
	hashedAPIKey := m.hashAPIKey(apiKey)

	// Look up service by API key hash
	service, err := m.serviceRepository.GetByAPIKey(ctx, hashedAPIKey)
	if err != nil {
		m.logger.Warn("Error looking up service by API key",
			zap.String("method", method),
			zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid API key")
	}

	if service == nil {
		m.logger.Warn("Invalid API key in gRPC request",
			zap.String("method", method))
		return nil, status.Errorf(codes.Unauthenticated, "invalid API key")
	}

	// Check if service is active
	if !service.IsActive {
		m.logger.Warn("Inactive service attempted authentication",
			zap.String("service_id", service.ID),
			zap.String("service_name", service.Name),
			zap.String("method", method))
		return nil, status.Errorf(codes.Unauthenticated, "service is inactive")
	}

	// Add service information to context
	ctx = context.WithValue(ctx, "service_id", service.ID)
	ctx = context.WithValue(ctx, "service_name", service.Name)
	ctx = context.WithValue(ctx, "principal_type", "service")
	ctx = context.WithValue(ctx, "user_id", service.ID) // Set user_id to service_id for compatibility

	m.logger.Debug("gRPC service authenticated",
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name),
		zap.String("method", method))

	return handler(ctx, req)
}

// hashAPIKey hashes an API key using SHA-256 (same as principal service)
func (m *AuthMiddleware) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
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
		userIDVal, exists := c.Get("user_id")
		userIDStr := ""
		if exists {
			if s, ok := userIDVal.(string); ok {
				userIDStr = s
			}
		}
		if userIDStr == "" {
			// Not authenticated → 401
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication required",
				"message": "login required",
			})
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
			// Treat backend failure as server error
			if err := c.Error(err); err != nil {
				m.logger.Warn("Failed to add error to context", zap.Error(err))
			}
			c.Abort()
			return
		}

		if !result.Allowed {
			m.logger.Warn("Access denied - insufficient permissions",
				zap.String("user_id", userIDStr),
				zap.String("resource", resource),
				zap.String("action", action))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": fmt.Sprintf("Insufficient permissions for %s:%s", resource, action),
			})
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
	// Exact-match endpoints
	switch path {
	case "/":
		return true
	case "/api/v1/auth/login", "/api/v1/auth/register", "/api/v1/auth/refresh", "/api/v1/health":
		return true
	}

	// Prefix-based endpoints (documentation assets)
	if strings.HasPrefix(path, "/docs") {
		return true
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
		// TokenService methods - these validate tokens, so they can't require tokens
		"/pb.TokenService/ValidateToken",
		"/pb.TokenService/IntrospectToken",
		"/pb.TokenService/RefreshAccessToken",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}

	return false
}
