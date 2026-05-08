package repository

import (
	"errors"

	"mds-go-api-docker/internal/model"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) FindAll() ([]model.Server, error) {
	var servers []model.Server
	if err := r.db.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *ServerRepository) FindByID(id string) (*model.Server, error) {
	var server model.Server
	err := r.db.First(&server, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) Create(req model.CreateServerRequest) (*model.Server, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}
	server := model.Server{
		Name:       req.Name,
		Host:       req.Host,
		User:       req.User,
		Port:       port,
		PrivateKey: req.PrivateKey,
		IsLocal:    req.IsLocal,
	}
	if err := r.db.Create(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) Update(id string, req model.UpdateServerRequest) (*model.Server, error) {
	server, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.User != nil {
		updates["user"] = *req.User
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.PrivateKey != nil {
		updates["private_key"] = *req.PrivateKey
	}
	if req.IsLocal != nil {
		updates["is_local"] = *req.IsLocal
	}

	if err := r.db.Model(server).Updates(updates).Error; err != nil {
		return nil, err
	}
	return server, nil
}

func (r *ServerRepository) Delete(id string) error {
	result := r.db.Delete(&model.Server{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
