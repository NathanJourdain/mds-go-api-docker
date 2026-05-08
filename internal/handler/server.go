package handler

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"
)

type ServerHandler struct {
	repo *repository.ServerRepository
}

func NewServerHandler(repo *repository.ServerRepository) *ServerHandler {
	return &ServerHandler{repo: repo}
}

func (h *ServerHandler) GetAll(c *fiber.Ctx) error {
	servers, err := h.repo.FindAll()
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(servers)
}

func (h *ServerHandler) GetByID(c *fiber.Ctx) error {
	server, err := h.repo.FindByID(c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(server)
}

func (h *ServerHandler) Create(c *fiber.Ctx) error {
	var req model.CreateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	if req.Name == "" {
		return badRequest(c, "name is required")
	}
	if !req.IsLocal && req.Host == "" {
		return badRequest(c, "host is required for remote servers")
	}
	if !req.IsLocal && req.User == "" {
		return badRequest(c, "user is required for remote servers")
	}
	server, err := h.repo.Create(req)
	if err != nil {
		return internalError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(server)
}

func (h *ServerHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	server, err := h.repo.Update(c.Params("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.JSON(server)
}

func (h *ServerHandler) Delete(c *fiber.Ctx) error {
	if err := h.repo.Delete(c.Params("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ServerHandler) TestConnection(c *fiber.Ctx) error {
	server, err := h.repo.FindByID(c.Params("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return notFound(c)
		}
		return internalError(c, err)
	}

	dockerSvc, err := service.NewDockerServiceForServer(*server)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	defer dockerSvc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dockerSvc.Ping(ctx); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
