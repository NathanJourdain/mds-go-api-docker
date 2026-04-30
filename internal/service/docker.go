package service

import (
	"context"
	"io"

	"mds-go-api-docker/internal/model"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerService struct {
	cli *client.Client
}

type ContainerConfig struct {
	Name    string
	Image   string
	Env     []string
	Ports   []model.PortMapping
	Volumes []model.VolumeMount
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerService{cli: cli}, nil
}

func (s *DockerService) ListImages(ctx context.Context) ([]image.Summary, error) {
	return s.cli.ImageList(ctx, image.ListOptions{All: true})
}

func (s *DockerService) GetImage(ctx context.Context, imageID string) (image.InspectResponse, error) {
	return s.cli.ImageInspect(ctx, imageID)
}

func (s *DockerService) PullImage(ctx context.Context, imageName string) error {
	out, err := s.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(io.Discard, out)
	return err
}

func (s *DockerService) CreateAndStartContainer(ctx context.Context, cfg ContainerConfig) (string, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, p := range cfg.Ports {
		proto := "tcp"
		if p.Protocol != "" {
			proto = p.Protocol
		}
		containerPort := nat.Port(p.Container + "/" + proto)
		portBindings[containerPort] = []nat.PortBinding{{HostPort: p.Host}}
		exposedPorts[containerPort] = struct{}{}
	}

	mounts := make([]mount.Mount, 0, len(cfg.Volumes))
	for _, v := range cfg.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: v.Source,
			Target: v.Target,
		})
	}

	resp, err := s.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        cfg.Image,
			Env:          cfg.Env,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			PortBindings: portBindings,
			Mounts:       mounts,
		},
		nil, nil,
		cfg.Name,
	)
	if err != nil {
		return "", err
	}

	if err := s.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (s *DockerService) ContainerStatus(ctx context.Context, dockerID string) (container.ContainerState, error) {
	info, err := s.cli.ContainerInspect(ctx, dockerID)
	if err != nil {
		return "", err
	}
	return info.State.Status, nil
}

func (s *DockerService) StartContainer(ctx context.Context, dockerID string) error {
	return s.cli.ContainerStart(ctx, dockerID, container.StartOptions{})
}

func (s *DockerService) StopContainer(ctx context.Context, dockerID string) error {
	timeout := 10
	return s.cli.ContainerStop(ctx, dockerID, container.StopOptions{Timeout: &timeout})
}

func (s *DockerService) RestartContainer(ctx context.Context, dockerID string) error {
	timeout := 10
	return s.cli.ContainerRestart(ctx, dockerID, container.StopOptions{Timeout: &timeout})
}

func (s *DockerService) StopAndRemoveContainer(ctx context.Context, dockerID string) error {
	timeout := 10
	if err := s.cli.ContainerStop(ctx, dockerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return err
	}
	return s.cli.ContainerRemove(ctx, dockerID, container.RemoveOptions{Force: true})
}
