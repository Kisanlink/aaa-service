package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type ResourceRepositoryInterface interface {
	CheckIfResourceExists(resourceName string) error
	CreateResource(newResource *model.Resource) error
	FindResourceByID(id string) (*model.Resource, error)
	FindResources(filter map[string]interface{}, page, limit int) ([]model.Resource, error)
	UpdateResource(id string, updatedResource model.Resource) error
	DeleteResource(id string) error
}

type ResourceRepository struct {
	DB *gorm.DB
}

func NewResourceRepository(db *gorm.DB) ResourceRepositoryInterface {
	return &ResourceRepository{
		DB: db,
	}
}

func (repo *ResourceRepository) CheckIfResourceExists(resourceName string) error {
	existingResource := model.Resource{}
	err := repo.DB.Table("resources").Where("name = ?", resourceName).First(&existingResource).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, fmt.Errorf("resource with name '%s' already exists", resourceName))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *ResourceRepository) CreateResource(newResource *model.Resource) error {
	if err := repo.DB.Table("resources").Create(newResource).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create resource: %w", err))
	}
	return nil
}

func (repo *ResourceRepository) FindResourceByID(id string) (*model.Resource, error) {
	var resource model.Resource
	err := repo.DB.Table("resources").Where("id = ?", id).First(&resource).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("resource with ID %s not found", id))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query resource: %w", err))
	}
	return &resource, nil
}

func (repo *ResourceRepository) FindResources(filter map[string]interface{}, page, limit int) ([]model.Resource, error) {
	var resources []model.Resource
	query := repo.DB.Table("resources")

	// Apply filters if provided
	if id, ok := filter["id"]; ok {
		query = query.Where("id = ?", id)
	}
	if name, ok := filter["name"]; ok {
		if nameStr, ok := name.(string); ok {
			query = query.Where("name ILIKE ?", "%"+nameStr+"%") // Case-insensitive partial match
		}
	}

	// Apply pagination if both page and limit are provided and valid
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Find(&resources).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve resources: %w", err))
	}
	return resources, nil
}
func (repo *ResourceRepository) UpdateResource(id string, updatedResource model.Resource) error {
	result := repo.DB.Table("resources").Where("id = ?", id).Updates(updatedResource)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update resource: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("resource with ID %s not found", id))
	}
	return nil
}

func (repo *ResourceRepository) DeleteResource(id string) error {
	result := repo.DB.Table("resources").Where("id = ?", id).Delete(&model.Resource{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete resource: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("resource with ID %s not found", id))
	}
	return nil
}
