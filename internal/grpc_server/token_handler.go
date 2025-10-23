package grpc_server

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/helper"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TokenHandler implements token-related gRPC services
type TokenHandler struct {
	pb.UnimplementedTokenServiceServer
	authService    *services.AuthService
	userService    interfaces.UserService
	authzService   *services.AuthorizationService
	cacheService   interfaces.CacheService
	userRepository interfaces.UserRepository
	logger         *zap.Logger
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(
	authService *services.AuthService,
	userService interfaces.UserService,
	authzService *services.AuthorizationService,
	cacheService interfaces.CacheService,
	userRepository interfaces.UserRepository,
	logger *zap.Logger,
) *TokenHandler {
	return &TokenHandler{
		authService:    authService,
		userService:    userService,
		authzService:   authzService,
		cacheService:   cacheService,
		userRepository: userRepository,
		logger:         logger,
	}
}

// ValidateToken validates a JWT token and returns claims and user context
func (h *TokenHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	h.logger.Info("gRPC ValidateToken request",
		zap.Bool("include_user_details", req.IncludeUserDetails),
		zap.Bool("include_permissions", req.IncludePermissions),
		zap.Bool("include_organization", req.IncludeOrganization),
		zap.Bool("strict_validation", req.StrictValidation))

	// Validate the token
	claims, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		h.logger.Warn("Token validation failed", zap.Error(err))
		return &pb.ValidateTokenResponse{
			StatusCode: 401,
			Message:    "Invalid token",
			Valid:      false,
		}, nil
	}

	// Parse full token context to extract organization info from user_context
	tokenContext, err := helper.ValidateTokenWithContext(req.Token)
	if err != nil {
		h.logger.Warn("Failed to parse full token context", zap.Error(err))
		// Continue with basic validation, but without organization context
	}

	validationID := uuid.New().String()
	now := time.Now()

	// Build token claims response
	tokenClaims := &pb.TokenClaims{
		UserId:    claims.UserID,
		Username:  claims.Username,
		TokenType: claims.TokenType,
		Jti:       claims.ID,
		Subject:   claims.Subject,
		Issuer:    claims.Issuer,
	}

	// Add timestamps
	if !claims.ExpiresAt.IsZero() {
		tokenClaims.ExpiresAt = timestamppb.New(claims.ExpiresAt.Time)
	}
	if !claims.IssuedAt.IsZero() {
		tokenClaims.IssuedAt = timestamppb.New(claims.IssuedAt.Time)
	}
	if !claims.NotBefore.IsZero() {
		tokenClaims.NotBefore = timestamppb.New(claims.NotBefore.Time)
	}

	// Add audience
	if len(claims.Audience) > 0 {
		tokenClaims.Audience = claims.Audience
	}

	// Extract roles from claims.Roles (which are UserRole objects with Role relationship)
	var roleNames []string
	for _, userRole := range claims.Roles {
		if userRole.Role.Name != "" {
			roleNames = append(roleNames, userRole.Role.Name)
		}
	}
	tokenClaims.Roles = roleNames

	// Add permissions from claims
	tokenClaims.Permissions = claims.Permissions

	response := &pb.ValidateTokenResponse{
		StatusCode:   200,
		Message:      "Token validated successfully",
		Valid:        true,
		Claims:       tokenClaims,
		ValidationId: validationID,
		ValidatedAt:  timestamppb.New(now),
	}

	// Include user details if requested
	if req.IncludeUserDetails {
		user, err := h.userRepository.GetByID(ctx, claims.UserID, &models.User{})
		if err != nil {
			h.logger.Warn("Failed to fetch user details", zap.String("user_id", claims.UserID), zap.Error(err))
			response.Warnings = append(response.Warnings, "User details could not be loaded")
		} else {
			response.UserContext = h.buildUserContext(user, claims, tokenContext)
		}
	}

	// Include permissions if requested
	if req.IncludePermissions {
		permissions, err := h.authzService.GetUserPermissions(ctx, claims.UserID)
		if err != nil {
			h.logger.Warn("Failed to fetch permissions", zap.String("user_id", claims.UserID), zap.Error(err))
			response.Warnings = append(response.Warnings, "Permissions could not be loaded")
		} else {
			response.Permissions = permissions
			if response.UserContext != nil {
				response.UserContext.Permissions = permissions
			}
		}
	}

	// Validate required permissions if specified
	if len(req.RequiredPermissions) > 0 {
		for _, reqPerm := range req.RequiredPermissions {
			hasPermission := false
			for _, userPerm := range response.Permissions {
				if userPerm == reqPerm {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				h.logger.Warn("User lacks required permission",
					zap.String("user_id", claims.UserID),
					zap.String("permission", reqPerm))
				return &pb.ValidateTokenResponse{
					StatusCode: 403,
					Message:    fmt.Sprintf("Missing required permission: %s", reqPerm),
					Valid:      false,
				}, nil
			}
		}
	}

	h.logger.Info("Token validated successfully",
		zap.String("user_id", claims.UserID),
		zap.String("validation_id", validationID))

	return response, nil
}

// buildUserContext creates a UserContext from user model and claims
func (h *TokenHandler) buildUserContext(user *models.User, claims *services.TokenClaims, tokenContext *helper.TokenContext) *pb.UserContext {
	userContext := &pb.UserContext{
		Id:          user.ID,
		IsValidated: user.IsValidated,
	}

	// Username
	if user.Username != nil {
		userContext.Username = *user.Username
	}

	// Status
	if user.Status != nil {
		userContext.Status = *user.Status
	}

	// Profile fields (from UserProfile relationship)
	if user.Profile.Name != nil {
		userContext.FullName = *user.Profile.Name
	}

	// Extract email from Contacts if available
	for _, contact := range user.Contacts {
		if contact.Type == "email" && contact.Value != "" {
			userContext.Email = contact.Value
			break
		}
	}

	// Add roles from claims
	var roleNames []string
	for _, userRole := range claims.Roles {
		if userRole.Role.Name != "" {
			roleNames = append(roleNames, userRole.Role.Name)
		}
	}
	userContext.Roles = roleNames

	// Populate organization context from JWT token's user_context.organizations array
	if tokenContext != nil && tokenContext.UserContext != nil && len(tokenContext.UserContext.Organizations) > 0 {
		// Use the first organization from the token's organizations array
		firstOrg := tokenContext.UserContext.Organizations[0]
		userContext.OrganizationId = firstOrg.ID
		userContext.OrganizationName = firstOrg.Name

		h.logger.Debug("Populated organization context from JWT token",
			zap.String("user_id", user.ID),
			zap.String("organization_id", firstOrg.ID),
			zap.String("organization_name", firstOrg.Name))
	}

	return userContext
}

// RefreshAccessToken refreshes an access token using a refresh token
func (h *TokenHandler) RefreshAccessToken(ctx context.Context, req *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	h.logger.Info("gRPC RefreshAccessToken request")

	// Refresh the token using AuthService
	// Note: AuthService.RefreshToken takes refreshToken and mPin as parameters
	refreshResp, err := h.authService.RefreshToken(ctx, req.RefreshToken, "") // Empty mPin for now
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		return &pb.RefreshAccessTokenResponse{
			StatusCode: 401,
			Message:    "Token refresh failed",
		}, status.Errorf(codes.Unauthenticated, "token refresh failed: %v", err)
	}

	return &pb.RefreshAccessTokenResponse{
		StatusCode:   200,
		Message:      "Token refreshed successfully",
		AccessToken:  refreshResp.AccessToken,
		RefreshToken: refreshResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(refreshResp.ExpiresIn),
	}, nil
}

// RevokeToken revokes a token
func (h *TokenHandler) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	h.logger.Info("gRPC RevokeToken request",
		zap.String("token_type_hint", req.TokenTypeHint),
		zap.Bool("revoke_all", req.RevokeAllUserTokens))

	// Parse the token to get user ID
	claims, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		h.logger.Warn("Failed to parse token for revocation", zap.Error(err))
		return &pb.RevokeTokenResponse{
			StatusCode: 400,
			Message:    "Invalid token",
			Success:    false,
		}, nil
	}

	// Add token to blacklist cache
	revocationID := uuid.New().String()
	cacheKey := fmt.Sprintf("revoked_token:%s", claims.ID)

	// Store in cache for the remaining lifetime of the token
	expiry := time.Until(claims.ExpiresAt.Time)
	if expiry > 0 {
		h.cacheService.Set(cacheKey, true, int(expiry.Seconds()))
	}

	tokensRevoked := int32(1)

	// If revoking all user tokens, add user to revocation list
	if req.RevokeAllUserTokens {
		userCacheKey := fmt.Sprintf("revoked_user:%s", claims.UserID)
		h.cacheService.Set(userCacheKey, true, int((24 * time.Hour).Seconds())) // Revoke for 24 hours
		tokensRevoked = -1                                                      // Unknown count when revoking all
	}

	h.logger.Info("Token revoked successfully",
		zap.String("user_id", claims.UserID),
		zap.String("revocation_id", revocationID))

	return &pb.RevokeTokenResponse{
		StatusCode:    200,
		Message:       "Token revoked successfully",
		Success:       true,
		TokensRevoked: tokensRevoked,
		RevocationId:  revocationID,
		RevokedAt:     timestamppb.New(time.Now()),
	}, nil
}

// IntrospectToken introspects a token (OAuth 2.0 style)
func (h *TokenHandler) IntrospectToken(ctx context.Context, req *pb.IntrospectTokenRequest) (*pb.IntrospectTokenResponse, error) {
	h.logger.Info("gRPC IntrospectToken request")

	// Validate the token
	claims, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		return &pb.IntrospectTokenResponse{
			Active: false,
		}, nil
	}

	// Check if token is revoked
	cacheKey := fmt.Sprintf("revoked_token:%s", claims.ID)
	if _, exists := h.cacheService.Get(cacheKey); exists {
		return &pb.IntrospectTokenResponse{
			Active: false,
		}, nil
	}

	return &pb.IntrospectTokenResponse{
		Active:    true,
		Username:  claims.Username,
		Sub:       claims.Subject,
		Iss:       claims.Issuer,
		Exp:       claims.ExpiresAt.Unix(),
		Iat:       claims.IssuedAt.Unix(),
		Jti:       claims.ID,
		TokenType: claims.TokenType,
	}, nil
}

// CreateToken creates a new token (for service-to-service communication)
func (h *TokenHandler) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	h.logger.Info("gRPC CreateToken request",
		zap.String("user_id", req.UserId),
		zap.String("organization_id", req.OrganizationId))

	// Get user details
	user, err := h.userRepository.GetByID(ctx, req.UserId, &models.User{})
	if err != nil {
		h.logger.Error("Failed to fetch user for token creation", zap.Error(err))
		return &pb.CreateTokenResponse{
			StatusCode: 404,
			Message:    "User not found",
		}, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// This is a simplified implementation
	// In production, you'd want to use AuthService.generateAccessToken
	h.logger.Warn("CreateToken method needs full implementation with proper role/permission loading",
		zap.String("user_id", user.ID))

	return &pb.CreateTokenResponse{
		StatusCode: 501,
		Message:    "CreateToken method not fully implemented",
	}, status.Errorf(codes.Unimplemented, "method not fully implemented")
}

// ListActiveTokens lists active tokens for a user
func (h *TokenHandler) ListActiveTokens(ctx context.Context, req *pb.ListActiveTokensRequest) (*pb.ListActiveTokensResponse, error) {
	h.logger.Info("gRPC ListActiveTokens request",
		zap.String("user_id", req.UserId))

	// This would require a token store/database
	// For now, return not implemented
	return &pb.ListActiveTokensResponse{
		StatusCode: 501,
		Message:    "ListActiveTokens method not fully implemented",
	}, status.Errorf(codes.Unimplemented, "method not fully implemented")
}

// BlacklistToken adds a token to the blacklist
func (h *TokenHandler) BlacklistToken(ctx context.Context, req *pb.BlacklistTokenRequest) (*pb.BlacklistTokenResponse, error) {
	h.logger.Info("gRPC BlacklistToken request",
		zap.String("token_id", req.TokenId))

	// Add to blacklist cache
	blacklistID := uuid.New().String()
	cacheKey := fmt.Sprintf("revoked_token:%s", req.TokenId)

	// Determine expiry
	expiry := 24 * time.Hour // Default
	if req.BlacklistUntil != nil {
		expiry = time.Until(req.BlacklistUntil.AsTime())
	}

	h.cacheService.Set(cacheKey, req.Reason, int(expiry.Seconds()))

	h.logger.Info("Token blacklisted successfully",
		zap.String("token_id", req.TokenId),
		zap.String("blacklist_id", blacklistID))

	return &pb.BlacklistTokenResponse{
		StatusCode:    200,
		Message:       "Token blacklisted successfully",
		Success:       true,
		BlacklistId:   blacklistID,
		BlacklistedAt: timestamppb.New(time.Now()),
	}, nil
}
