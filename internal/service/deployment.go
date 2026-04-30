package service

import (
	"context"
	"fmt"
	"time"

	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type DeploymentService struct {
	docker      *DockerService
	deployRepo  *repository.DeploymentRepository
	projectRepo *repository.ProjectRepository
}

func NewDeploymentService(
	docker *DockerService,
	deployRepo *repository.DeploymentRepository,
	projectRepo *repository.ProjectRepository,
) *DeploymentService {
	return &DeploymentService{
		docker:      docker,
		deployRepo:  deployRepo,
		projectRepo: projectRepo,
	}
}

func (s *DeploymentService) Deploy(ctx context.Context, projectID string, req model.CreateDeploymentRequest) (*model.Deployment, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}

	deployment, err := s.deployRepo.Create(project.ID, req)
	if err != nil {
		return nil, err
	}

	overrides := make(map[string]string, len(deployment.EnvOverride))
	for _, ov := range deployment.EnvOverride {
		overrides[ov.Key] = ov.Value
	}

	sorted, err := topoSort(project.Services)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, svc := range sorted {
		if err := s.docker.PullImage(ctx, svc.Image); err != nil {
			return nil, fmt.Errorf("pull %s: %w", svc.Image, err)
		}

		dockerID, err := s.docker.CreateAndStartContainer(ctx, ContainerConfig{
			Name:    fmt.Sprintf("%s_%s", deployment.Name, svc.Name),
			Image:   svc.Image,
			Env:     mergeEnv(svc.EnvVars, overrides),
			Ports:   svc.Ports,
			Volumes: svc.VolumeMounts,
		})
		if err != nil {
			return nil, fmt.Errorf("start %s: %w", svc.Name, err)
		}

		c := model.Container{
			DeploymentID: deployment.ID,
			ServiceID:    svc.ID,
			DockerID:     dockerID,
			Name:         fmt.Sprintf("%s_%s", deployment.Name, svc.Name),
			Ports:        svc.Ports,
		}
		if err := s.deployRepo.SaveContainer(&c); err != nil {
			return nil, err
		}
	}

	s.deployRepo.UpdateStartedAt(deployment.ID, now) //nolint:errcheck

	return s.GetWithStatus(ctx, deployment.ID)
}

func (s *DeploymentService) ListByProject(projectID string) ([]model.Deployment, error) {
	return s.deployRepo.FindByProjectID(projectID)
}

func (s *DeploymentService) GetWithStatus(ctx context.Context, id uint) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	running := 0
	for i, c := range deployment.Containers {
		status, err := s.docker.ContainerStatus(ctx, c.DockerID)
		if err != nil {
			deployment.Containers[i].Status = "unknown"
		} else {
			deployment.Containers[i].Status = status
			if status == "running" {
				running++
			}
		}
	}

	total := len(deployment.Containers)
	switch {
	case total == 0 || running == 0:
		deployment.Status = "stopped"
	case running == total:
		deployment.Status = "running"
	default:
		deployment.Status = "partially-running"
	}

	return deployment, nil
}

func (s *DeploymentService) GetWithStatusByStr(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByIDStr(id)
	if err != nil {
		return nil, err
	}

	running := 0
	for i, c := range deployment.Containers {
		status, err := s.docker.ContainerStatus(ctx, c.DockerID)
		if err != nil {
			deployment.Containers[i].Status = "unknown"
		} else {
			deployment.Containers[i].Status = status
			if status == "running" {
				running++
			}
		}
	}

	total := len(deployment.Containers)
	switch {
	case total == 0 || running == 0:
		deployment.Status = "stopped"
	case running == total:
		deployment.Status = "running"
	default:
		deployment.Status = "partially-running"
	}

	return deployment, nil
}

func (s *DeploymentService) StartContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByIDStr(id)
	if err != nil {
		return nil, err
	}
	for _, c := range deployment.Containers {
		s.docker.StartContainer(ctx, c.DockerID) //nolint:errcheck
	}
	return s.GetWithStatusByStr(ctx, id)
}

func (s *DeploymentService) StopContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByIDStr(id)
	if err != nil {
		return nil, err
	}
	for _, c := range deployment.Containers {
		s.docker.StopContainer(ctx, c.DockerID) //nolint:errcheck
	}
	return s.GetWithStatusByStr(ctx, id)
}

func (s *DeploymentService) RestartContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByIDStr(id)
	if err != nil {
		return nil, err
	}
	for _, c := range deployment.Containers {
		s.docker.RestartContainer(ctx, c.DockerID) //nolint:errcheck
	}
	return s.GetWithStatusByStr(ctx, id)
}

func (s *DeploymentService) Stop(ctx context.Context, id string) error {
	deployment, err := s.deployRepo.FindByIDStr(id)
	if err != nil {
		return err
	}

	for _, c := range deployment.Containers {
		s.docker.StopAndRemoveContainer(ctx, c.DockerID) //nolint:errcheck
	}

	return s.deployRepo.Delete(id)
}

// mergeEnv fusionne les EnvVar du service avec les overrides du déploiement.
func mergeEnv(envVars []model.EnvVar, overrides map[string]string) []string {
	result := make([]string, 0, len(envVars))
	for _, ev := range envVars {
		value := ev.Value
		if ev.IsVariable {
			if ov, ok := overrides[ev.Key]; ok {
				value = ov
			}
		}
		result = append(result, ev.Key+"="+value)
	}
	return result
}

// topoSort trie les services selon leurs DependsOn (tri topologique).
func topoSort(services []model.Service) ([]model.Service, error) {
	byName := make(map[string]model.Service, len(services))
	for _, s := range services {
		byName[s.Name] = s
	}

	visited := map[string]bool{}
	inProgress := map[string]bool{}
	result := make([]model.Service, 0, len(services))

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if inProgress[name] {
			return fmt.Errorf("dependency cycle detected at service %q", name)
		}
		svc, ok := byName[name]
		if !ok {
			return fmt.Errorf("dependency %q not found", name)
		}
		inProgress[name] = true
		for _, dep := range svc.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		delete(inProgress, name)
		visited[name] = true
		result = append(result, svc)
		return nil
	}

	for _, svc := range services {
		if err := visit(svc.Name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
