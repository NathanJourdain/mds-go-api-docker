package handler

import (
	"context"
	"errors"

	"mds-go-api-docker/internal/logger"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DeploymentHandler struct {
	svc *service.DeploymentService
}

func NewDeploymentHandler(svc *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

// contextWithRequestID enriches the Fiber request context with the request ID
// so the service layer can include it in structured log fields.
func contextWithRequestID(c *fiber.Ctx) context.Context {
	id, _ := c.Locals("requestid").(string)
	return logger.ContextWithRequestID(c.Context(), id)
}

func (h *DeploymentHandler) Deploy(c *fiber.Ctx) error {
	var req model.CreateDeploymentRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}

	deployment, err := h.svc.Deploy(contextWithRequestID(c), c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(deployment)
}

func (h *DeploymentHandler) ListByProject(c *fiber.Ctx) error {
	deployments, err := h.svc.ListByProject(c.Params("id"))
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(deployments)
}

func (h *DeploymentHandler) GetByID(c *fiber.Ctx) error {
	deployment, err := h.svc.GetWithStatus(contextWithRequestID(c), c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(deployment)
}

func (h *DeploymentHandler) Start(c *fiber.Ctx) error {
	deployment, err := h.svc.StartContainers(contextWithRequestID(c), c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(deployment)
}

func (h *DeploymentHandler) Stop(c *fiber.Ctx) error {
	deployment, err := h.svc.StopContainers(contextWithRequestID(c), c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(deployment)
}

func (h *DeploymentHandler) Restart(c *fiber.Ctx) error {
	deployment, err := h.svc.RestartContainers(contextWithRequestID(c), c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(deployment)
}

func (h *DeploymentHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Stop(contextWithRequestID(c), c.Params("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
