package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	applogger "mds-go-api-docker/internal/logger"
)

func RequestID() fiber.Handler {
	return requestid.New()
}

func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     `{"time":"${time}","level":"INFO","request_id":"${locals:requestid}","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ip":"${ip}"}` + "\n",
		TimeFormat: "2006-01-02T15:04:05Z07:00",
		TimeZone:   "UTC",
		Output:     applogger.RequestOutput(),
	})
}
