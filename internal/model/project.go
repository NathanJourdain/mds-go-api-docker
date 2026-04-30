package model

import "time"

type Project struct {
	ID          uint      `json:"id"          gorm:"primarykey;autoIncrement"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
