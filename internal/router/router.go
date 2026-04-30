package router

import (
	"mds-go-api-docker/internal/handler"
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, dockerSvc *service.DockerService) {
	repoArticle := repository.NewArticleRepository(db)
	hArticle := handler.NewArticleHandler(repoArticle)

	hDocker := handler.NewDockerHandler(dockerSvc)

	api := app.Group("/api/v1")

	articles := api.Group("/articles")
	articles.Get("/", hArticle.GetAll)
	articles.Get("/:id", hArticle.GetByID)
	articles.Post("/", hArticle.Create)
	articles.Patch("/:id", hArticle.Update)
	articles.Delete("/:id", hArticle.Delete)

	docker := api.Group("/docker")
	docker.Get("/images", hDocker.ListImages)
	docker.Get("/images/:id", hDocker.GetImage)
}
