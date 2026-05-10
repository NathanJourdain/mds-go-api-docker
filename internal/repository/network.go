package repository

import (
	"errors"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

type NetworkRepository struct {
	db *gorm.DB
}

func NewNetworkRepository(db *gorm.DB) *NetworkRepository {
	return &NetworkRepository{db: db}
}

func (r *NetworkRepository) Create(projectID string, req model.CreateNetworkRequest) (*model.Network, error) {
	var project model.Project
	if err := r.db.First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	driver := req.Driver
	if driver == "" {
		driver = "bridge"
	}

	network := model.Network{
		ProjectID: project.ID,
		Name:      req.Name,
		Driver:    driver,
	}
	if err := r.db.Create(&network).Error; err != nil {
		return nil, err
	}
	return &network, nil
}

func (r *NetworkRepository) Delete(projectID, networkID string) error {
	result := r.db.Where("id = ? AND project_id = ?", networkID, projectID).Delete(&model.Network{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
