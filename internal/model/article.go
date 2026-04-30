package model

import "time"

type Article struct {
	ID        uint      `json:"id"         gorm:"primarykey;autoIncrement"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"      gorm:"not null"`
	Content   string    `json:"content"`
	Author    string    `json:"author"     gorm:"not null"`
}

type CreateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type UpdateArticleRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Author  *string `json:"author"`
}
