package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"mds-go-api-docker/internal/database"
	applogger "mds-go-api-docker/internal/logger"
	"mds-go-api-docker/internal/middleware"
	"mds-go-api-docker/internal/router"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		level = slog.LevelDebug
	}
	applogger.Init(level)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "app.db"
	}

	db, err := database.New(dsn)
	if err != nil {
		applogger.App.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	applogger.App.Info("database connected", "dsn", dsn)

	app := fiber.New(fiber.Config{
		AppName: "mds-go-api",
	})

	app.Use(cors.New(cors.Config{AllowOrigins: "*"}))
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		applogger.App.Error("API_KEY environment variable is required")
		os.Exit(1)
	}
	app.Use(middleware.APIKey(apiKey))

	router.Setup(app, db)

	app.Get("/docs/openapi.yaml", func(c *fiber.Ctx) error {
		spec, err := os.ReadFile("./public/docs/openapi.yaml")
		if err != nil {
			return err
		}
		apiURL := os.Getenv("API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:3000"
		}
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.SendString(strings.ReplaceAll(string(spec), "__API_URL__", apiURL))
	})

	app.Static("/", "./public")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	const port = "3000"
	applogger.App.Info("server starting", "port", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		applogger.App.Info("shutdown signal received", "signal", sig.String())
		_ = app.Shutdown()
	}()

	if err := app.Listen(":" + port); err != nil {
		applogger.App.Error("server stopped", "error", err)
	}
}
