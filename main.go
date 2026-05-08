package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"mds-go-api-docker/internal/database"
	"mds-go-api-docker/internal/router"
	"mds-go-api-docker/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "app.db"
	}

	db, err := database.New(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	dockerSvc, err := service.NewDockerService()
	if err != nil {
		log.Fatalf("docker: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "mds-go-api",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	router.Setup(app, db, dockerSvc)

	app.Static("/", "./public")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
