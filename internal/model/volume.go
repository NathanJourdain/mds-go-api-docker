package model

import "time"

type Volume struct {
	ID        uint      `json:"id"         gorm:"primarykey;autoIncrement"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ProjectID uint      `json:"project_id" gorm:"not null;index"`
	Name      string    `json:"name"       gorm:"not null"`
	Driver    string    `json:"driver"     gorm:"default:local"`
}

type CreateVolumeRequest struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}
