package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	applogger "mds-go-api-docker/internal/logger"
)

func APIKey(key string) fiber.Handler {
	return keyauth.New(keyauth.Config{
		KeyLookup: "header:X-API-Key",
		Next: func(c *fiber.Ctx) bool {
			return !strings.HasPrefix(c.Path(), "/api/")
		},
		Validator: func(c *fiber.Ctx, k string) (bool, error) {
			return k == key, nil
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			applogger.App.Warn("auth failure",
				"ip", c.IP(),
				"method", c.Method(),
				"path", c.Path(),
				"request_id", c.Locals("requestid"),
				"error", err.Error(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid API key",
			})
		},
	})
}
