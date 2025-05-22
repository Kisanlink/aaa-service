package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type ResourceServiceInterface interface {
	CheckIfResourceExists(resourceName string) error
	CreateResource(newResource *model.Resource) error
	FindResourceByID(id string) (*model.Resource, error)
	FindResources(filter map[string]interface{}) ([]model.Resource, error)
	UpdateResource(id string, updatedResource model.Resource) error
	DeleteResource(id string) error
}

type ResourceService struct {
	repo repositories.ResourceRepositoryInterface
}

func NewResourceService(repo repositories.ResourceRepositoryInterface) ResourceServiceInterface {
	return &ResourceService{
		repo: repo,
	}
}

func (s *ResourceService) CheckIfResourceExists(resourceName string) error {
	err := s.repo.CheckIfResourceExists(resourceName)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check if resource exists: %w", err))
	}
	return nil
}

func (s *ResourceService) CreateResource(newResource *model.Resource) error {
	err := s.repo.CreateResource(newResource)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create resource: %w", err))
	}
	return nil
}

func (s *ResourceService) FindResourceByID(id string) (*model.Resource, error) {
	result, err := s.repo.FindResourceByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get resource by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("resource not found"))
	}
	return result, nil
}
func (s *ResourceService) FindResources(filter map[string]interface{}) ([]model.Resource, error) {
	resources, err := s.repo.FindResources(filter)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve resources: %w", err))
	}
	return resources, nil
}

func (s *ResourceService) UpdateResource(id string, updatedResource model.Resource) error {
	err := s.repo.UpdateResource(id, updatedResource)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update resource: %w", err))
	}
	return nil
}

func (s *ResourceService) DeleteResource(id string) error {
	err := s.repo.DeleteResource(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete resource: %w", err))
	}
	return nil
}
