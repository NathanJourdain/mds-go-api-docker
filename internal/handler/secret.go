package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type SecretHandler struct {
	repo *repository.SecretRepository
}

func NewSecretHandler(repo *repository.SecretRepository) *SecretHandler {
	return &SecretHandler{repo: repo}
}

func (h *SecretHandler) Create(c *fiber.Ctx) error {
	var req model.CreateSecretRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}

	secret, err := h.repo.Create(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(secret)
}

func (h *SecretHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id"), c.Params("sid")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
