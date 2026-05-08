package model

type Container struct {
	IDModel
	DeploymentID string        `json:"deployment_id" gorm:"not null;index;type:text"`
	ServiceID    string        `json:"service_id"    gorm:"not null;type:text"`
	DockerID     string        `json:"docker_id"`
	Name         string        `json:"name"`
	Status       string        `json:"status"        gorm:"-"`
	Ports        []PortMapping `json:"ports"         gorm:"serializer:json"`
}
