package repository

import (
	"errors"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(projectID string, req model.CreateServiceRequest) (*model.Service, error) {
	var project model.Project
	if err := r.db.First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	service := model.Service{
		ProjectID:    project.ID,
		Name:         req.Name,
		Image:        req.Image,
		Ports:        req.Ports,
		Secrets:      req.Secrets,
		Networks:     req.Networks,
		VolumeMounts: req.VolumeMounts,
		DependsOn:    req.DependsOn,
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&service).Error; err != nil {
			return err
		}
		for i := range req.EnvVars {
			req.EnvVars[i].ID = ""
			req.EnvVars[i].ServiceID = service.ID
		}
		if len(req.EnvVars) > 0 {
			if err := tx.Create(&req.EnvVars).Error; err != nil {
				return err
			}
		}
		for i := range req.Labels {
			req.Labels[i].ID = ""
			req.Labels[i].ServiceID = service.ID
		}
		if len(req.Labels) > 0 {
			return tx.Create(&req.Labels).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.findByID(service.ID)
}

func (r *ServiceRepository) Update(projectID, serviceID string, req model.UpdateServiceRequest) (*model.Service, error) {
	service, err := r.findByIDAndProject(serviceID, projectID)
	if err != nil {
		return nil, err
	}

	updates := model.Service{}
	selected := []string{}

	if req.Name != nil {
		updates.Name = *req.Name
		selected = append(selected, "Name")
	}
	if req.Image != nil {
		updates.Image = *req.Image
		selected = append(selected, "Image")
	}
	if req.Ports != nil {
		updates.Ports = req.Ports
		selected = append(selected, "Ports")
	}
	if req.VolumeMounts != nil {
		updates.VolumeMounts = req.VolumeMounts
		selected = append(selected, "VolumeMounts")
	}
	if req.DependsOn != nil {
		updates.DependsOn = req.DependsOn
		selected = append(selected, "DependsOn")
	}
	if req.Secrets != nil {
		updates.Secrets = req.Secrets
		selected = append(selected, "Secrets")
	}
	if req.Networks != nil {
		updates.Networks = req.Networks
		selected = append(selected, "Networks")
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if len(selected) > 0 {
			if err := tx.Model(service).Select(selected).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.EnvVars != nil {
			if err := tx.Where("service_id = ?", service.ID).Delete(&model.EnvVar{}).Error; err != nil {
				return err
			}
			for i := range *req.EnvVars {
				(*req.EnvVars)[i].ID = ""
				(*req.EnvVars)[i].ServiceID = service.ID
			}
			if len(*req.EnvVars) > 0 {
				if err := tx.Create(req.EnvVars).Error; err != nil {
					return err
				}
			}
		}
		if req.Labels != nil {
			if err := tx.Where("service_id = ?", service.ID).Delete(&model.Label{}).Error; err != nil {
				return err
			}
			for i := range *req.Labels {
				(*req.Labels)[i].ID = ""
				(*req.Labels)[i].ServiceID = service.ID
			}
			if len(*req.Labels) > 0 {
				return tx.Create(req.Labels).Error
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.findByID(service.ID)
}

func (r *ServiceRepository) Delete(projectID, serviceID string) error {
	service, err := r.findByIDAndProject(serviceID, projectID)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("service_id = ?", service.ID).Delete(&model.EnvVar{})
		tx.Where("service_id = ?", service.ID).Delete(&model.Label{})
		return tx.Delete(service).Error
	})
}

func (r *ServiceRepository) findByID(id string) (*model.Service, error) {
	var service model.Service
	err := r.db.Preload("EnvVars").Preload("Labels").First(&service, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &service, nil
}

func (r *ServiceRepository) findByIDAndProject(serviceID, projectID string) (*model.Service, error) {
	var service model.Service
	err := r.db.Where("id = ? AND project_id = ?", serviceID, projectID).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &service, nil
}
