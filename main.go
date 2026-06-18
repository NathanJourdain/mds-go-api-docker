package main

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"mds-go-api-docker/internal/database"
	"mds-go-api-docker/internal/middleware"
	"mds-go-api-docker/internal/router"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "app.db"
	}

	db, err := database.New(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "mds-go-api",
	})

	app.Use(cors.New(cors.Config{AllowOrigins: "*"}))
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
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

	slog.Info("server starting", "port", port)
	log.Fatal(app.Listen(":" + port))
}
