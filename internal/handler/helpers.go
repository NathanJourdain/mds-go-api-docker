package handler

import (
	applogger "mds-go-api-docker/internal/logger"

	"github.com/gofiber/fiber/v2"
)

func notFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func internalError(c *fiber.Ctx, err error) error {
	applogger.App.Error("internal error",
		"request_id", c.Locals("requestid"),
		"method", c.Method(),
		"path", c.Path(),
		"error", err.Error(),
	)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}
