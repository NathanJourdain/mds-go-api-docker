package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type NetworkHandler struct {
	repo *repository.NetworkRepository
}

func NewNetworkHandler(repo *repository.NetworkRepository) *NetworkHandler {
	return &NetworkHandler{repo: repo}
}

func (h *NetworkHandler) Create(c *fiber.Ctx) error {
	var req model.CreateNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}

	network, err := h.repo.Create(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(network)
}

func (h *NetworkHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id"), c.Params("nid")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
