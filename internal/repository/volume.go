package repository

import (
	"errors"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

type VolumeRepository struct {
	db *gorm.DB
}

func NewVolumeRepository(db *gorm.DB) *VolumeRepository {
	return &VolumeRepository{db: db}
}

func (r *VolumeRepository) Create(projectID string, req model.CreateVolumeRequest) (*model.Volume, error) {
	var project model.Project
	if err := r.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	driver := req.Driver
	if driver == "" {
		driver = "local"
	}

	volume := model.Volume{
		ProjectID: project.ID,
		Name:      req.Name,
		Driver:    driver,
	}
	if err := r.db.Create(&volume).Error; err != nil {
		return nil, err
	}
	return &volume, nil
}

func (r *VolumeRepository) Delete(projectID, volumeID string) error {
	result := r.db.Where("id = ? AND project_id = ?", volumeID, projectID).Delete(&model.Volume{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
