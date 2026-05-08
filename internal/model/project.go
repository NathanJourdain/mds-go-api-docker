package model

type Project struct {
	Base
	Name        string    `json:"name"        gorm:"not null;uniqueIndex"`
	Description string    `json:"description"`
	Services    []Service `json:"services,omitempty" gorm:"foreignKey:ProjectID"`
	Volumes     []Volume  `json:"volumes,omitempty"  gorm:"foreignKey:ProjectID"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
