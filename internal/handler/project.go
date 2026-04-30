package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type ProjectHandler struct {
	repo *repository.ProjectRepository
}

func NewProjectHandler(repo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

func (h *ProjectHandler) GetAll(c *fiber.Ctx) error {
	projects, err := h.repo.FindAll()
	if err != nil {
		return internalError(c)
	}
	return c.JSON(projects)
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	project, err := h.repo.FindByID(c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.JSON(project)
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	var req model.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}
	project, err := h.repo.Create(req)
	if err != nil {
		return internalError(c)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	project, err := h.repo.Update(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.JSON(project)
}

func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
