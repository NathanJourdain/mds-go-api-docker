package router

import (
	"mds-go-api-docker/internal/handler"
	"mds-go-api-docker/internal/repository"
	"mds-go-api-docker/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, dockerSvc *service.DockerService) {
	projectRepo := repository.NewProjectRepository(db)

	hDocker := handler.NewDockerHandler(dockerSvc)
	hProject := handler.NewProjectHandler(projectRepo)
	hService := handler.NewServiceHandler(repository.NewServiceRepository(db))
	hVolume := handler.NewVolumeHandler(repository.NewVolumeRepository(db))
	hDeployment := handler.NewDeploymentHandler(
		service.NewDeploymentService(dockerSvc, repository.NewDeploymentRepository(db), projectRepo),
	)

	api := app.Group("/api/v1")

	projects := api.Group("/projects")
	projects.Get("/", hProject.GetAll)
	projects.Post("/", hProject.Create)
	projects.Get("/:id", hProject.GetByID)
	projects.Put("/:id", hProject.Update)
	projects.Delete("/:id", hProject.Delete)

	projects.Post("/:id/services", hService.Create)
	projects.Put("/:id/services/:sid", hService.Update)
	projects.Delete("/:id/services/:sid", hService.Delete)

	projects.Post("/:id/volumes", hVolume.Create)
	projects.Delete("/:id/volumes/:vid", hVolume.Delete)

	projects.Post("/:id/deployments", hDeployment.Deploy)
	projects.Get("/:id/deployments", hDeployment.ListByProject)

	deployments := api.Group("/deployments")
	deployments.Get("/:id", hDeployment.GetByID)
	deployments.Delete("/:id", hDeployment.Delete)
	deployments.Post("/:id/start", hDeployment.Start)
	deployments.Post("/:id/stop", hDeployment.Stop)
	deployments.Post("/:id/restart", hDeployment.Restart)

	docker := api.Group("/docker")
	docker.Get("/images", hDocker.ListImages)
	docker.Get("/images/:id", hDocker.GetImage)
}
