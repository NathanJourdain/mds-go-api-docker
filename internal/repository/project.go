package repository

import (
	"errors"

	"mds-go-api-docker/internal/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) FindAll() ([]model.Project, error) {
	var projects []model.Project
	if err := r.db.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) FindByID(id string) (*model.Project, error) {
	var project model.Project
	err := r.db.
		Preload("Services.EnvVars").
		Preload("Services.Labels").
		Preload("Volumes").
		Preload("Networks").
		Preload("Secrets").
		First(&project, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Create(req model.CreateProjectRequest) (*model.Project, error) {
	project := model.Project{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := r.db.Create(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Update(id string, req model.UpdateProjectRequest) (*model.Project, error) {
	project, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if err := r.db.Model(project).Updates(updates).Error; err != nil {
		return nil, err
	}
	return project, nil
}

func (r *ProjectRepository) Delete(id string) error {
	result := r.db.Delete(&model.Project{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
