package model

type Server struct {
	ID      uint   `json:"id"      gorm:"primarykey;autoIncrement"`
	Name    string `json:"name"    gorm:"not null"`
	Host    string `json:"host"    gorm:"not null"`
	IsLocal bool   `json:"is_local"`
}
