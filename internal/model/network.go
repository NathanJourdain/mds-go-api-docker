package model

type Network struct {
	Base
	ProjectID string `json:"project_id" gorm:"not null;index;type:text"`
	Name      string `json:"name"       gorm:"not null"`
	Driver    string `json:"driver"     gorm:"default:bridge"`
}

type CreateNetworkRequest struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}
