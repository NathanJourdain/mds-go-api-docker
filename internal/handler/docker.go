package handler

import (
	"mds-go-api-docker/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DockerHandler struct {
	docker *service.DockerService
}

func NewDockerHandler(docker *service.DockerService) *DockerHandler {
	return &DockerHandler{docker: docker}
}

func (h *DockerHandler) ListImages(c *fiber.Ctx) error {
	images, err := h.docker.ListImages(c.Context())
	if err != nil {
		return internalError(c)
	}
	return c.JSON(images)
}

func (h *DockerHandler) GetImage(c *fiber.Ctx) error {
	imageID := c.Params("id")

	image, err := h.docker.GetImage(c.Context(), imageID)
	if err != nil {
		return internalError(c)
	}

	return c.JSON(image)
}
