package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type ActionServiceInterface interface {
	CheckIfActionExists(actionName string) error
	CreateAction(newAction *model.Action) error
	FindActionByID(id string) (*model.Action, error)
	FindActions(filter map[string]interface{}, page, limit int) ([]model.Action, error)
	UpdateAction(id string, updatedAction model.Action) error
	DeleteAction(id string) error
}

type ActionService struct {
	repo repositories.ActionRepositoryInterface
}

func NewActionService(repo repositories.ActionRepositoryInterface) ActionServiceInterface {
	return &ActionService{
		repo: repo,
	}
}

func (s *ActionService) CheckIfActionExists(actionName string) error {
	err := s.repo.CheckIfActionExists(actionName)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check if action exists: %w", err))
	}
	return nil
}

func (s *ActionService) CreateAction(newAction *model.Action) error {
	err := s.repo.CreateAction(newAction)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create action: %w", err))
	}
	return nil
}

func (s *ActionService) FindActionByID(id string) (*model.Action, error) {
	result, err := s.repo.FindActionByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get action by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("action not found"))
	}
	return result, nil
}

func (s *ActionService) FindActions(filter map[string]interface{}, page, limit int) ([]model.Action, error) {
	actions, err := s.repo.FindActions(filter, page, limit)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve actions: %w", err))
	}
	return actions, nil
}
func (s *ActionService) UpdateAction(id string, updatedAction model.Action) error {
	err := s.repo.UpdateAction(id, updatedAction)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update action: %w", err))
	}
	return nil
}

func (s *ActionService) DeleteAction(id string) error {
	err := s.repo.DeleteAction(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete action: %w", err))
	}
	return nil
}
