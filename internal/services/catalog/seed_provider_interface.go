package catalog

import (
	"context"
	"sync"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// SeedDataProvider defines the interface for providing seed data for roles and permissions
// This allows different services (ERP, Farmers, etc.) to register their own custom roles
type SeedDataProvider interface {
	// GetServiceID returns the unique identifier for the service providing this seed data
	GetServiceID() string

	// GetServiceName returns the human-readable name of the service
	GetServiceName() string

	// GetResources returns all resources to be seeded for this service
	GetResources() []ResourceDefinition

	// GetActions returns all actions to be seeded for this service
	GetActions() []ActionDefinition

	// GetRoles returns all roles to be seeded with their permissions for this service
	GetRoles() []RoleDefinition

	// Validate validates the seed data before execution
	Validate(ctx context.Context) error
}

// SeedProviderRegistry manages registration of multiple seed data providers
// Thread-safe for concurrent access
type SeedProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]SeedDataProvider
}

// NewSeedProviderRegistry creates a new seed provider registry
func NewSeedProviderRegistry() *SeedProviderRegistry {
	return &SeedProviderRegistry{
		providers: make(map[string]SeedDataProvider),
	}
}

// Register registers a new seed data provider
// Thread-safe: uses write lock
func (r *SeedProviderRegistry) Register(provider SeedDataProvider) error {
	serviceID := provider.GetServiceID()
	if serviceID == "" {
		return ErrInvalidServiceID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[serviceID]; exists {
		return ErrProviderAlreadyRegistered
	}

	r.providers[serviceID] = provider
	return nil
}

// Get retrieves a provider by service ID
// Thread-safe: uses read lock
func (r *SeedProviderRegistry) Get(serviceID string) (SeedDataProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[serviceID]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// GetAll returns all registered providers
// Thread-safe: uses read lock
func (r *SeedProviderRegistry) GetAll() []SeedDataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]SeedDataProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// Unregister removes a provider from the registry
// Thread-safe: uses write lock
func (r *SeedProviderRegistry) Unregister(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[serviceID]; !exists {
		return ErrProviderNotFound
	}
	delete(r.providers, serviceID)
	return nil
}

// Has checks if a provider is registered
// Thread-safe: uses read lock
func (r *SeedProviderRegistry) Has(serviceID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[serviceID]
	return exists
}

// Count returns the number of registered providers
// Thread-safe: uses read lock
func (r *SeedProviderRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

// Clear removes all providers from the registry
// Thread-safe: uses write lock
func (r *SeedProviderRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = make(map[string]SeedDataProvider)
}

// BaseSeedProvider provides a base implementation for common provider operations
type BaseSeedProvider struct {
	serviceID   string
	serviceName string
}

// NewBaseSeedProvider creates a new base seed provider
func NewBaseSeedProvider(serviceID, serviceName string) *BaseSeedProvider {
	return &BaseSeedProvider{
		serviceID:   serviceID,
		serviceName: serviceName,
	}
}

// GetServiceID returns the service ID
func (b *BaseSeedProvider) GetServiceID() string {
	return b.serviceID
}

// GetServiceName returns the service name
func (b *BaseSeedProvider) GetServiceName() string {
	return b.serviceName
}

// Validate performs basic validation (can be overridden)
func (b *BaseSeedProvider) Validate(ctx context.Context) error {
	if b.serviceID == "" {
		return ErrInvalidServiceID
	}
	if b.serviceName == "" {
		return ErrInvalidServiceName
	}
	return nil
}

// ValidateRole validates a role definition
func ValidateRole(role RoleDefinition) error {
	if role.Name == "" {
		return ErrInvalidRoleName
	}
	if role.Scope != models.RoleScopeGlobal && role.Scope != models.RoleScopeOrg {
		return ErrInvalidRoleScope
	}
	if len(role.Permissions) == 0 {
		return ErrEmptyPermissions
	}
	return nil
}

// ValidateResource validates a resource definition
func ValidateResource(resource ResourceDefinition) error {
	if resource.Name == "" {
		return ErrInvalidResourceName
	}
	if resource.Type == "" {
		return ErrInvalidResourceType
	}
	return nil
}

// ValidateAction validates an action definition
func ValidateAction(action ActionDefinition) error {
	if action.Name == "" {
		return ErrInvalidActionName
	}
	if action.Category == "" {
		return ErrInvalidActionCategory
	}
	return nil
}
