package model

type Service struct {
	Base
	ProjectID    string        `json:"project_id"         gorm:"not null;index;type:text"`
	Name         string        `json:"name"               gorm:"not null"`
	Image        string        `json:"image"              gorm:"not null"`
	Ports        []PortMapping `json:"ports"              gorm:"serializer:json"`
	EnvVars      []EnvVar      `json:"env_vars,omitempty" gorm:"foreignKey:ServiceID"`
	VolumeMounts []VolumeMount `json:"volume_mounts"      gorm:"serializer:json"`
	DependsOn    []string      `json:"depends_on"         gorm:"serializer:json"`
}

type EnvVar struct {
	IDModel
	ServiceID  string `json:"service_id"  gorm:"not null;index;type:text"`
	Key        string `json:"key"         gorm:"not null"`
	Value      string `json:"value"`
	IsVariable bool   `json:"is_variable"`
}

type CreateServiceRequest struct {
	Name         string        `json:"name"`
	Image        string        `json:"image"`
	Ports        []PortMapping `json:"ports"`
	EnvVars      []EnvVar      `json:"env_vars"`
	VolumeMounts []VolumeMount `json:"volume_mounts"`
	DependsOn    []string      `json:"depends_on"`
}

type UpdateServiceRequest struct {
	Name         *string        `json:"name"`
	Image        *string        `json:"image"`
	Ports        []PortMapping  `json:"ports"`
	EnvVars      *[]EnvVar      `json:"env_vars"`
	VolumeMounts []VolumeMount  `json:"volume_mounts"`
	DependsOn    []string       `json:"depends_on"`
}
