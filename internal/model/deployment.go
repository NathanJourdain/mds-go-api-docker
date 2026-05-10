package model

import "time"

type Deployment struct {
	Base
	StartedAt      *time.Time                  `json:"started_at"`
	ProjectID      string                      `json:"project_id"              gorm:"not null;index;type:text"`
	ServerID       *string                     `json:"server_id"               gorm:"index;type:text"`
	Server         *Server                     `json:"server,omitempty"        gorm:"foreignKey:ServerID"`
	Name           string                      `json:"name"                    gorm:"not null"`
	Status         string                      `json:"status"                  gorm:"-"`
	EnvOverride    []DeploymentEnvOverride     `json:"env_override,omitempty"  gorm:"foreignKey:DeploymentID"`
	SecretOverride []DeploymentSecretOverride  `json:"secret_override,omitempty" gorm:"foreignKey:DeploymentID"`
	Containers     []Container                 `json:"containers,omitempty"    gorm:"foreignKey:DeploymentID"`
	Networks       []DeploymentNetwork         `json:"networks,omitempty"      gorm:"foreignKey:DeploymentID"`
}

type DeploymentEnvOverride struct {
	IDModel
	DeploymentID string `json:"deployment_id" gorm:"not null;index;type:text"`
	Key          string `json:"key"           gorm:"not null"`
	Value        string `json:"value"`
}

type DeploymentSecretOverride struct {
	IDModel
	DeploymentID string `json:"deployment_id" gorm:"not null;index;type:text"`
	Name         string `json:"name"          gorm:"not null"`
	Value        string `json:"-"             gorm:"not null"`
}

type DeploymentNetwork struct {
	IDModel
	DeploymentID    string `json:"deployment_id"     gorm:"not null;index;type:text"`
	Name            string `json:"name"`
	DockerNetworkID string `json:"docker_network_id"`
}

type CreateDeploymentRequest struct {
	Name           string            `json:"name"`
	ServerID       *string           `json:"server_id"`
	EnvOverride    []DeploymentEnvOverride `json:"env_override"`
	SecretOverride map[string]string `json:"secret_override"`
	Scale          map[string]int    `json:"scale"`
}
