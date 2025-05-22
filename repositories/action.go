package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type ActionRepositoryInterface interface {
	CheckIfActionExists(actionName string) error
	CreateAction(newAction *model.Action) error
	FindActionByID(id string) (*model.Action, error)
	FindActions(filter map[string]interface{}) ([]model.Action, error)
	UpdateAction(id string, updatedAction model.Action) error
	DeleteAction(id string) error
}

type ActionRepository struct {
	DB *gorm.DB
}

func NewActionRepository(db *gorm.DB) ActionRepositoryInterface {
	return &ActionRepository{
		DB: db,
	}
}

func (repo *ActionRepository) CheckIfActionExists(actionName string) error {
	existingAction := model.Action{}
	err := repo.DB.Table("actions").Where("name = ?", actionName).First(&existingAction).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, fmt.Errorf("action with name '%s' already exists", actionName))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *ActionRepository) CreateAction(newAction *model.Action) error {
	if err := repo.DB.Table("actions").Create(newAction).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create action: %w", err))
	}
	return nil
}

func (repo *ActionRepository) FindActionByID(id string) (*model.Action, error) {
	var action model.Action
	err := repo.DB.Table("actions").Where("id = ?", id).First(&action).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("action with ID %s not found", id))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query action: %w", err))
	}
	return &action, nil
}

func (repo *ActionRepository) FindActions(filter map[string]interface{}) ([]model.Action, error) {
	var actions []model.Action
	query := repo.DB.Table("actions")

	// Apply filters if provided
	if id, ok := filter["id"]; ok {
		query = query.Where("id = ?", id)
	}
	if name, ok := filter["name"]; ok {
		query = query.Where("name ILIKE ?", name)
	}

	err := query.Find(&actions).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve actions: %w", err))
	}
	return actions, nil
}

func (repo *ActionRepository) UpdateAction(id string, updatedAction model.Action) error {
	result := repo.DB.Table("actions").Where("id = ?", id).Updates(updatedAction)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update action: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("action with ID %s not found", id))
	}
	return nil
}

func (repo *ActionRepository) DeleteAction(id string) error {
	result := repo.DB.Table("actions").Where("id = ?", id).Delete(&model.Action{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete action: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("action with ID %s not found", id))
	}
	return nil
}
