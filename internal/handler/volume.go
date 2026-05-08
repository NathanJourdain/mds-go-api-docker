package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type VolumeHandler struct {
	repo *repository.VolumeRepository
}

func NewVolumeHandler(repo *repository.VolumeRepository) *VolumeHandler {
	return &VolumeHandler{repo: repo}
}

func (h *VolumeHandler) Create(c *fiber.Ctx) error {
	var req model.CreateVolumeRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}

	volume, err := h.repo.Create(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(volume)
}

func (h *VolumeHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id"), c.Params("vid")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
