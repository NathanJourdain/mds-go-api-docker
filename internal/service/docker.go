package service

import (
	"context"
	"fmt"
	"io"
	"net"

	applogger "mds-go-api-docker/internal/logger"
	"mds-go-api-docker/internal/model"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	gossh "golang.org/x/crypto/ssh"
)

type DockerService struct {
	cli       *client.Client
	sshClient *gossh.Client
}

type ContainerConfig struct {
	Name     string
	Image    string
	Env      []string
	Ports    []model.PortMapping
	Volumes  []model.VolumeMount
	Labels   map[string]string
	Networks []string
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerService{cli: cli}, nil
}

func NewDockerServiceForServer(server model.Server) (*DockerService, error) {
	if server.IsLocal {
		return NewDockerService()
	}

	applogger.Docker.Debug("establishing SSH tunnel to Docker",
		"host", server.Host,
		"port", server.Port,
		"user", server.User,
	)

	signer, err := gossh.ParsePrivateKey([]byte(server.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	sshConfig := &gossh.ClientConfig{
		User:            server.User,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	sshCli, err := gossh.Dial("tcp", addr, sshConfig)
	if err != nil {
		applogger.Docker.Error("SSH dial failed", "addr", addr, "error", err)
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	cli, err := client.NewClientWithOpts(
		client.WithDialContext(func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			return sshCli.Dial("unix", "/var/run/docker.sock")
		}),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		_ = sshCli.Close()
		return nil, err
	}

	applogger.Docker.Info("SSH+Docker connection established", "host", server.Host)
	return &DockerService{cli: cli, sshClient: sshCli}, nil
}

func (s *DockerService) Close() {
	_ = s.cli.Close()
	if s.sshClient != nil {
		_ = s.sshClient.Close()
	}
}

func (s *DockerService) Ping(ctx context.Context) error {
	_, err := s.cli.Ping(ctx)
	return err
}

func (s *DockerService) ListImages(ctx context.Context) ([]image.Summary, error) {
	return s.cli.ImageList(ctx, image.ListOptions{All: true})
}

func (s *DockerService) GetImage(ctx context.Context, imageID string) (image.InspectResponse, error) {
	return s.cli.ImageInspect(ctx, imageID)
}

func (s *DockerService) PullImage(ctx context.Context, imageName string) error {
	applogger.Docker.Info("pulling image", "image", imageName)
	out, err := s.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		applogger.Docker.Error("image pull failed", "image", imageName, "error", err)
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err = io.Copy(io.Discard, out); err != nil {
		applogger.Docker.Error("image pull stream error", "image", imageName, "error", err)
		return err
	}
	applogger.Docker.Info("image ready", "image", imageName)
	return nil
}

func (s *DockerService) CreateNetwork(ctx context.Context, name, driver string) (string, error) {
	if driver == "" {
		driver = "bridge"
	}
	applogger.Docker.Debug("creating Docker network", "name", name, "driver", driver)
	resp, err := s.cli.NetworkCreate(ctx, name, dockernetwork.CreateOptions{Driver: driver})
	if err != nil {
		applogger.Docker.Error("network create failed", "name", name, "error", err)
		return "", err
	}
	applogger.Docker.Info("network created", "name", name, "docker_network_id", resp.ID)
	return resp.ID, nil
}

func (s *DockerService) RemoveNetwork(ctx context.Context, dockerNetworkID string) error {
	applogger.Docker.Info("removing network", "docker_network_id", dockerNetworkID)
	err := s.cli.NetworkRemove(ctx, dockerNetworkID)
	if err != nil {
		applogger.Docker.Warn("network remove failed", "docker_network_id", dockerNetworkID, "error", err)
	}
	return err
}

func (s *DockerService) CreateAndStartContainer(ctx context.Context, cfg ContainerConfig) (string, error) {
	applogger.Docker.Debug("creating container", "name", cfg.Name, "image", cfg.Image)

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
			Labels:       cfg.Labels,
		},
		&container.HostConfig{
			PortBindings: portBindings,
			Mounts:       mounts,
		},
		nil, nil,
		cfg.Name,
	)
	if err != nil {
		applogger.Docker.Error("container create failed", "name", cfg.Name, "image", cfg.Image, "error", err)
		return "", err
	}

	for _, networkID := range cfg.Networks {
		if err := s.cli.NetworkConnect(ctx, networkID, resp.ID, &dockernetwork.EndpointSettings{}); err != nil {
			return "", fmt.Errorf("connect to network %s: %w", networkID, err)
		}
	}

	if err := s.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		applogger.Docker.Error("container start failed", "name", cfg.Name, "docker_id", resp.ID, "error", err)
		return "", err
	}

	applogger.Docker.Info("container running", "name", cfg.Name, "image", cfg.Image, "docker_id", resp.ID)
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
	applogger.Docker.Info("starting container", "docker_id", dockerID)
	return s.cli.ContainerStart(ctx, dockerID, container.StartOptions{})
}

func (s *DockerService) StopContainer(ctx context.Context, dockerID string) error {
	applogger.Docker.Info("stopping container", "docker_id", dockerID)
	timeout := 10
	return s.cli.ContainerStop(ctx, dockerID, container.StopOptions{Timeout: &timeout})
}

func (s *DockerService) RestartContainer(ctx context.Context, dockerID string) error {
	applogger.Docker.Info("restarting container", "docker_id", dockerID)
	timeout := 10
	return s.cli.ContainerRestart(ctx, dockerID, container.StopOptions{Timeout: &timeout})
}

func (s *DockerService) StopAndRemoveContainer(ctx context.Context, dockerID string) error {
	applogger.Docker.Info("stopping and removing container", "docker_id", dockerID)
	timeout := 10
	if err := s.cli.ContainerStop(ctx, dockerID, container.StopOptions{Timeout: &timeout}); err != nil {
		applogger.Docker.Warn("container stop error (proceeding with remove)", "docker_id", dockerID, "error", err)
		return err
	}
	return s.cli.ContainerRemove(ctx, dockerID, container.RemoveOptions{Force: true})
}
