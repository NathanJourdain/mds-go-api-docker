package repository

import (
	"errors"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

type SecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) Create(projectID string, req model.CreateSecretRequest) (*model.Secret, error) {
	var project model.Project
	if err := r.db.First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	secret := model.Secret{
		ProjectID: project.ID,
		Name:      req.Name,
	}
	if err := r.db.Create(&secret).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *SecretRepository) Delete(projectID, secretID string) error {
	result := r.db.Where("id = ? AND project_id = ?", secretID, projectID).Delete(&model.Secret{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
