package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type ArticleHandler struct {
	repo *repository.ArticleRepository
}

func NewArticleHandler(repo *repository.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{repo: repo}
}

func (h *ArticleHandler) GetAll(c *fiber.Ctx) error {
	articles, err := h.repo.FindAll()
	if err != nil {
		return internalError(c)
	}
	return c.JSON(articles)
}

func (h *ArticleHandler) GetByID(c *fiber.Ctx) error {
	article, err := h.repo.FindByID(c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.JSON(article)
}

func (h *ArticleHandler) Create(c *fiber.Ctx) error {
	var req model.CreateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Title == "" || req.Author == "" {
		return badRequest(c, "title and author are required")
	}
	article, err := h.repo.Create(req)
	if err != nil {
		return internalError(c)
	}
	return c.Status(fiber.StatusCreated).JSON(article)
}

func (h *ArticleHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	article, err := h.repo.Update(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.JSON(article)
}

func (h *ArticleHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func notFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
