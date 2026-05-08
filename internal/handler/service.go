package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type ServiceHandler struct {
	repo *repository.ServiceRepository
}

func NewServiceHandler(repo *repository.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{repo: repo}
}

func (h *ServiceHandler) Create(c *fiber.Ctx) error {
	var req model.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" || req.Image == "" {
		return badRequest(c, "name and image are required")
	}

	service, err := h.repo.Create(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(service)
}

func (h *ServiceHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}

	service, err := h.repo.Update(c.Params("id"), c.Params("sid"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(service)
}

func (h *ServiceHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id"), c.Params("sid")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
