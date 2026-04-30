package model

type Container struct {
	ID           uint         `json:"id"            gorm:"primarykey;autoIncrement"`
	DeploymentID uint         `json:"deployment_id" gorm:"not null;index"`
	ServiceID    uint         `json:"service_id"    gorm:"not null"`
	DockerID     string       `json:"docker_id"`
	Name         string       `json:"name"`
	Status       string       `json:"status"        gorm:"-"` // récupéré via Docker SDK
	Ports        []PortMapping `json:"ports"        gorm:"serializer:json"`
}
