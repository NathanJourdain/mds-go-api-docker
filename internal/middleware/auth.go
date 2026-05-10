package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
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
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid API key",
			})
		},
	})
}
