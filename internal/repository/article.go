package repository

import (
	"errors"

	"gorm.io/gorm"
	"mds-go-api-docker/internal/model"
)

var ErrNotFound = errors.New("not found")

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) FindAll() ([]model.Article, error) {
	var articles []model.Article
	if err := r.db.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *ArticleRepository) FindByID(id string) (*model.Article, error) {
	var article model.Article
	if err := r.db.First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Create(req model.CreateArticleRequest) (*model.Article, error) {
	article := model.Article{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	}
	if err := r.db.Create(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Update(id string, req model.UpdateArticleRequest) (*model.Article, error) {
	article, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Author != nil {
		updates["author"] = *req.Author
	}

	if err := r.db.Model(article).Updates(updates).Error; err != nil {
		return nil, err
	}
	return article, nil
}

func (r *ArticleRepository) Delete(id string) error {
	result := r.db.Delete(&model.Article{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
