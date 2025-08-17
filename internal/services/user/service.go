package user

import (
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"go.uber.org/zap"
)

// Service implements the UserService interface with dependency injection
type Service struct {
	userRepo     interfaces.UserRepository
	roleRepo     interfaces.RoleRepository
	userRoleRepo interfaces.UserRoleRepository
	cacheService interfaces.CacheService
	logger       *zap.Logger
	validator    interfaces.Validator
}

// NewService creates a new UserService instance with proper dependency injection
func NewService(
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	userRoleRepo interfaces.UserRoleRepository,
	cacheService interfaces.CacheService,
	logger *zap.Logger,
	validator interfaces.Validator,
) interfaces.UserService {
	return &Service{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		cacheService: cacheService,
		logger:       logger,
		validator:    validator,
	}
}
