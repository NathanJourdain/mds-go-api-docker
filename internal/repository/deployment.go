package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

type DeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(projectID uint, req model.CreateDeploymentRequest) (*model.Deployment, error) {
	deployment := model.Deployment{
		ProjectID: projectID,
		Name:      req.Name,
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&deployment).Error; err != nil {
			return err
		}
		for i := range req.EnvOverride {
			req.EnvOverride[i].ID = 0
			req.EnvOverride[i].DeploymentID = deployment.ID
		}
		if len(req.EnvOverride) > 0 {
			return tx.Create(&req.EnvOverride).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.FindByID(deployment.ID)
}

func (r *DeploymentRepository) FindByProjectID(projectID string) ([]model.Deployment, error) {
	var deployments []model.Deployment
	err := r.db.Where("project_id = ?", projectID).Find(&deployments).Error
	return deployments, err
}

func (r *DeploymentRepository) FindByID(id uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Preload("Containers").Preload("EnvOverride").First(&deployment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &deployment, nil
}

func (r *DeploymentRepository) FindByIDStr(id string) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Preload("Containers").Preload("EnvOverride").First(&deployment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &deployment, nil
}

func (r *DeploymentRepository) SaveContainer(c *model.Container) error {
	return r.db.Create(c).Error
}

func (r *DeploymentRepository) UpdateStartedAt(id uint, t time.Time) error {
	return r.db.Model(&model.Deployment{}).Where("id = ?", id).Update("started_at", t).Error
}

func (r *DeploymentRepository) ReplaceEnvOverride(deploymentID uint, overrides []model.DeploymentEnvOverride) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("deployment_id = ?", deploymentID).Delete(&model.DeploymentEnvOverride{}).Error; err != nil {
			return err
		}
		for i := range overrides {
			overrides[i].ID = 0
			overrides[i].DeploymentID = deploymentID
		}
		if len(overrides) > 0 {
			return tx.Create(&overrides).Error
		}
		return nil
	})
}

func (r *DeploymentRepository) Delete(id string) error {
	var deployment model.Deployment
	if err := r.db.First(&deployment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("deployment_id = ?", deployment.ID).Delete(&model.Container{})
		tx.Where("deployment_id = ?", deployment.ID).Delete(&model.DeploymentEnvOverride{})
		return tx.Delete(&deployment).Error
	})
}
