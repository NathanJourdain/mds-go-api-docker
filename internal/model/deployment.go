package model

import "time"

type Deployment struct {
	ID          uint                    `json:"id"           gorm:"primarykey;autoIncrement"`
	CreatedAt   time.Time               `json:"created_at"`
	StartedAt   *time.Time              `json:"started_at"`
	ProjectID   uint                    `json:"project_id"   gorm:"not null;index"`
	Name        string                  `json:"name"         gorm:"not null"`
	Status      string                  `json:"status"       gorm:"-"` // calculé via Docker SDK
	EnvOverride []DeploymentEnvOverride `json:"env_override,omitempty" gorm:"foreignKey:DeploymentID"`
	Containers  []Container             `json:"containers,omitempty"   gorm:"foreignKey:DeploymentID"`
}

// DeploymentEnvOverride surcharge les EnvVar marquées IsVariable=true au moment du déploiement.
type DeploymentEnvOverride struct {
	ID           uint   `json:"id"            gorm:"primarykey;autoIncrement"`
	DeploymentID uint   `json:"deployment_id" gorm:"not null;index"`
	Key          string `json:"key"           gorm:"not null"`
	Value        string `json:"value"`
}

type CreateDeploymentRequest struct {
	Name        string                  `json:"name"`
	EnvOverride []DeploymentEnvOverride `json:"env_override"`
}
