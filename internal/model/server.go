package model

type Server struct {
	Base
	Name       string `json:"name"       gorm:"not null"`
	Host       string `json:"host"       gorm:"not null"`
	User       string `json:"user"       gorm:"not null"`
	Port       int    `json:"port"       gorm:"not null;default:22"`
	PrivateKey string `json:"-"          gorm:"type:text"`
	IsLocal    bool   `json:"is_local"`
}

type CreateServerRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	User       string `json:"user"`
	Port       int    `json:"port"`
	PrivateKey string `json:"private_key"`
	IsLocal    bool   `json:"is_local"`
}

type UpdateServerRequest struct {
	Name       *string `json:"name"`
	Host       *string `json:"host"`
	User       *string `json:"user"`
	Port       *int    `json:"port"`
	PrivateKey *string `json:"private_key"`
	IsLocal    *bool   `json:"is_local"`
}
