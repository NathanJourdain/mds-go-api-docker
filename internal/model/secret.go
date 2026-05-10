package model

type Secret struct {
	Base
	ProjectID string `json:"project_id" gorm:"not null;index;type:text"`
	Name      string `json:"name"       gorm:"not null"`
}

type CreateSecretRequest struct {
	Name string `json:"name"`
}
