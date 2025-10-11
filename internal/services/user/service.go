package user

import (
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"go.uber.org/zap"
)

// Service implements the UserService interface with dependency injection
type Service struct {
	userRepo            interfaces.UserRepository
	roleRepo            interfaces.RoleRepository
	userRoleRepo        interfaces.UserRoleRepository
	groupMembershipRepo any // Optional: for fetching user's groups
	groupRepo           any // Optional: for fetching group details
	organizationRepo    any // Optional: for fetching organization details
	cacheService        interfaces.CacheService
	logger              *zap.Logger
	validator           interfaces.Validator
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

// SetOrganizationalRepositories injects organizational repositories for JWT context enhancement
// This is optional and should be called after service initialization if organizational context is needed
// Accepts any types that implement the necessary methods (duck typing)
func (s *Service) SetOrganizationalRepositories(
	groupMembershipRepo any,
	groupRepo any,
	organizationRepo any,
) {
	s.groupMembershipRepo = groupMembershipRepo
	s.groupRepo = groupRepo
	s.organizationRepo = organizationRepo
	s.logger.Info("Organizational repositories injected for JWT context enhancement")
}
