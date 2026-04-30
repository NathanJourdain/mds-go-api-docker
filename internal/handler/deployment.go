package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"
)

type DeploymentHandler struct {
	svc *service.DeploymentService
}

func NewDeploymentHandler(svc *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

func (h *DeploymentHandler) Deploy(c *fiber.Ctx) error {
	var req model.CreateDeploymentRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}

	deployment, err := h.svc.Deploy(c.Context(), c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(deployment)
}

func (h *DeploymentHandler) ListByProject(c *fiber.Ctx) error {
	deployments, err := h.svc.ListByProject(c.Params("id"))
	if err != nil {
		return internalError(c)
	}
	return c.JSON(deployments)
}

func (h *DeploymentHandler) GetByID(c *fiber.Ctx) error {
	deployment, err := h.svc.GetWithStatusByStr(c.Context(), c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.JSON(deployment)
}

func (h *DeploymentHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Stop(c.Context(), c.Params("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
