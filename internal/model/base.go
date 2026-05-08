package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IDModel fournit un UUID comme clé primaire. À embarquer dans tous les modèles.
type IDModel struct {
	ID string `json:"id" gorm:"primarykey;type:text"`
}

func (m *IDModel) BeforeCreate(_ *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// Base embarque IDModel + timestamps gérés par GORM.
type Base struct {
	IDModel
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
