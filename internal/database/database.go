package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mds-go-api-docker/internal/model"
)

func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.Project{},
		&model.Service{},
		&model.EnvVar{},
		&model.Volume{},
		&model.Deployment{},
		&model.DeploymentEnvOverride{},
		&model.Container{},
		&model.Server{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
