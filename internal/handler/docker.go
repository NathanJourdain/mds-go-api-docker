package handler

import (
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DockerHandler struct {
	serverRepo *repository.ServerRepository
}

func NewDockerHandler(serverRepo *repository.ServerRepository) *DockerHandler {
	return &DockerHandler{serverRepo: serverRepo}
}

func (h *DockerHandler) dockerService(c *fiber.Ctx) (*service.DockerService, error) {
	serverID := c.Query("server_id")
	if serverID == "" {
		return service.NewDockerService()
	}
	server, err := h.serverRepo.FindByID(serverID)
	if err != nil {
		return nil, err
	}
	return service.NewDockerServiceForServer(*server)
}

func (h *DockerHandler) ListImages(c *fiber.Ctx) error {
	docker, err := h.dockerService(c)
	if err != nil {
		return internalError(c, err)
	}
	defer docker.Close()
	images, err := docker.ListImages(c.Context())
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(images)
}

func (h *DockerHandler) GetImage(c *fiber.Ctx) error {
	docker, err := h.dockerService(c)
	if err != nil {
		return internalError(c, err)
	}
	defer docker.Close()
	image, err := docker.GetImage(c.Context(), c.Params("id"))
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(image)
}
