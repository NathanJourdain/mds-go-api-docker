package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	applogger "mds-go-api-docker/internal/logger"
	"mds-go-api-docker/internal/model"
	"mds-go-api-docker/internal/repository"
)

type DeploymentService struct {
	serverRepo  *repository.ServerRepository
	deployRepo  *repository.DeploymentRepository
	projectRepo *repository.ProjectRepository
}

func NewDeploymentService(
	serverRepo *repository.ServerRepository,
	deployRepo *repository.DeploymentRepository,
	projectRepo *repository.ProjectRepository,
) *DeploymentService {
	return &DeploymentService{
		serverRepo:  serverRepo,
		deployRepo:  deployRepo,
		projectRepo: projectRepo,
	}
}

func (s *DeploymentService) dockerForDeployment(deployment *model.Deployment) (*DockerService, error) {
	if deployment.ServerID == nil || deployment.Server == nil {
		return NewDockerService()
	}
	return NewDockerServiceForServer(*deployment.Server)
}

func (s *DeploymentService) applyStatus(ctx context.Context, deployment *model.Deployment, docker *DockerService) (*model.Deployment, error) {
	running := 0
	for i, c := range deployment.Containers {
		status, err := docker.ContainerStatus(ctx, c.DockerID)
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

func (s *DeploymentService) Deploy(ctx context.Context, projectID string, req model.CreateDeploymentRequest) (*model.Deployment, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}

	deployment, err := s.deployRepo.Create(project.ID, req)
	if err != nil {
		return nil, err
	}

	log := applogger.Docker.With(
		"request_id", applogger.RequestIDFromContext(ctx),
		"deployment_id", deployment.ID,
		"project_id", projectID,
	)
	log.Info("deployment started", "name", req.Name)

	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		log.Error("docker connect failed", "error", err)
		return nil, err
	}
	defer docker.Close()

	// Create Docker networks and build a name→dockerID map
	dockerNetworkIDs := make(map[string]string, len(project.Networks))
	for _, n := range project.Networks {
		dockerName := fmt.Sprintf("%s_%s", deployment.ID[:8], n.Name)
		dockerNetworkID, err := docker.CreateNetwork(ctx, dockerName, n.Driver)
		if err != nil {
			return nil, fmt.Errorf("create network %s: %w", n.Name, err)
		}
		dockerNetworkIDs[n.Name] = dockerNetworkID
		dn := model.DeploymentNetwork{
			DeploymentID:    deployment.ID,
			Name:            n.Name,
			DockerNetworkID: dockerNetworkID,
		}
		if err := s.deployRepo.SaveNetwork(&dn); err != nil {
			return nil, err
		}
		log.Info("network provisioned", "network_name", n.Name, "docker_network_id", dockerNetworkID)
	}

	// Build env overrides and secrets maps
	overrides := make(map[string]string, len(deployment.EnvOverride))
	for _, ov := range deployment.EnvOverride {
		overrides[ov.Key] = ov.Value
	}
	secretsByName := make(map[string]string, len(deployment.SecretOverride))
	for _, sec := range deployment.SecretOverride {
		secretsByName[sec.Name] = sec.Value
	}

	sorted, err := topoSort(project.Services)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, svc := range sorted {
		log.Info("pulling image", "image", svc.Image, "service", svc.Name)
		if err := docker.PullImage(ctx, svc.Image); err != nil {
			return nil, fmt.Errorf("pull %s: %w", svc.Image, err)
		}

		// Resolve service network names to Docker network IDs
		networkIDs := make([]string, 0, len(svc.Networks))
		for _, name := range svc.Networks {
			id, ok := dockerNetworkIDs[name]
			if !ok {
				return nil, fmt.Errorf("service %q references unknown network %q", svc.Name, name)
			}
			networkIDs = append(networkIDs, id)
		}

		labels := make(map[string]string, len(svc.Labels))
		for _, l := range svc.Labels {
			labels[l.Key] = l.Value
		}

		env := mergeEnv(svc.EnvVars, overrides, secretsByName, svc.Secrets)

		replicas := 1
		if n, ok := req.Scale[svc.Name]; ok && n > 1 {
			replicas = n
		}

		for i := range replicas {
			replicaIdx := i + 1
			ports := scalePorts(svc.Ports, i)
			name := fmt.Sprintf("%s_%s_%d", deployment.Name, svc.Name, replicaIdx)

			dockerID, err := docker.CreateAndStartContainer(ctx, ContainerConfig{
				Name:     name,
				Image:    svc.Image,
				Env:      env,
				Ports:    ports,
				Volumes:  svc.VolumeMounts,
				Labels:   labels,
				Networks: networkIDs,
			})
			if err != nil {
				log.Error("container start failed", "service", svc.Name, "replica", replicaIdx, "error", err)
				return nil, fmt.Errorf("start %s replica %d: %w", svc.Name, replicaIdx, err)
			}

			log.Info("container started", "service", svc.Name, "replica", replicaIdx, "name", name, "docker_id", dockerID)

			c := model.Container{
				DeploymentID: deployment.ID,
				ServiceID:    svc.ID,
				DockerID:     dockerID,
				Name:         name,
				ReplicaIndex: replicaIdx,
				Ports:        ports,
			}
			if err := s.deployRepo.SaveContainer(&c); err != nil {
				return nil, err
			}
		}
	}

	s.deployRepo.UpdateStartedAt(deployment.ID, now) //nolint:errcheck

	deployment, err = s.deployRepo.FindByID(deployment.ID)
	if err != nil {
		return nil, err
	}

	log.Info("deployment complete", "container_count", len(deployment.Containers))
	return s.applyStatus(ctx, deployment, docker)
}

func (s *DeploymentService) ListByProject(projectID string) ([]model.Deployment, error) {
	return s.deployRepo.FindByProjectID(projectID)
}

func (s *DeploymentService) GetWithStatus(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		return nil, err
	}
	defer docker.Close()
	return s.applyStatus(ctx, deployment, docker)
}

func (s *DeploymentService) StartContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		return nil, err
	}
	defer docker.Close()

	applogger.Docker.Info("starting containers",
		"request_id", applogger.RequestIDFromContext(ctx),
		"deployment_id", id,
		"count", len(deployment.Containers),
	)
	for _, c := range deployment.Containers {
		docker.StartContainer(ctx, c.DockerID) //nolint:errcheck
	}
	applogger.Docker.Info("containers started", "deployment_id", id)
	return s.applyStatus(ctx, deployment, docker)
}

func (s *DeploymentService) StopContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		return nil, err
	}
	defer docker.Close()

	applogger.Docker.Info("stopping containers",
		"request_id", applogger.RequestIDFromContext(ctx),
		"deployment_id", id,
		"count", len(deployment.Containers),
	)
	for _, c := range deployment.Containers {
		docker.StopContainer(ctx, c.DockerID) //nolint:errcheck
	}
	applogger.Docker.Info("containers stopped", "deployment_id", id)
	return s.applyStatus(ctx, deployment, docker)
}

func (s *DeploymentService) RestartContainers(ctx context.Context, id string) (*model.Deployment, error) {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		return nil, err
	}
	defer docker.Close()

	applogger.Docker.Info("restarting containers",
		"request_id", applogger.RequestIDFromContext(ctx),
		"deployment_id", id,
		"count", len(deployment.Containers),
	)
	for _, c := range deployment.Containers {
		docker.RestartContainer(ctx, c.DockerID) //nolint:errcheck
	}
	applogger.Docker.Info("containers restarted", "deployment_id", id)
	return s.applyStatus(ctx, deployment, docker)
}

func (s *DeploymentService) Stop(ctx context.Context, id string) error {
	deployment, err := s.deployRepo.FindByID(id)
	if err != nil {
		return err
	}
	docker, err := s.dockerForDeployment(deployment)
	if err != nil {
		return err
	}
	defer docker.Close()

	applogger.Docker.Info("tearing down deployment",
		"request_id", applogger.RequestIDFromContext(ctx),
		"deployment_id", id,
		"container_count", len(deployment.Containers),
		"network_count", len(deployment.Networks),
	)
	for _, c := range deployment.Containers {
		docker.StopAndRemoveContainer(ctx, c.DockerID) //nolint:errcheck
	}
	for _, n := range deployment.Networks {
		docker.RemoveNetwork(ctx, n.DockerNetworkID) //nolint:errcheck
	}
	applogger.Docker.Info("deployment removed", "deployment_id", id)
	return s.deployRepo.Delete(id)
}

func scalePorts(ports []model.PortMapping, offset int) []model.PortMapping {
	if offset == 0 {
		return ports
	}
	result := make([]model.PortMapping, len(ports))
	for i, p := range ports {
		host := p.Host
		if n, err := strconv.Atoi(p.Host); err == nil {
			host = strconv.Itoa(n + offset)
		}
		result[i] = model.PortMapping{Host: host, Container: p.Container, Protocol: p.Protocol}
	}
	return result
}

func mergeEnv(envVars []model.EnvVar, overrides map[string]string, secrets map[string]string, serviceSecrets []string) []string {
	result := make([]string, 0, len(envVars)+len(serviceSecrets))
	for _, ev := range envVars {
		value := ev.Value
		if ev.IsVariable {
			if ov, ok := overrides[ev.Key]; ok {
				value = ov
			}
		}
		result = append(result, ev.Key+"="+value)
	}
	for _, name := range serviceSecrets {
		if value, ok := secrets[name]; ok {
			result = append(result, name+"="+value)
		}
	}
	return result
}

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
