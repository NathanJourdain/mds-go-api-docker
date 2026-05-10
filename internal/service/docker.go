package service

import (
	"context"
	"fmt"
	"io"
	"net"

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
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	cli, err := client.NewClientWithOpts(
		client.WithDialContext(func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			return sshCli.Dial("unix", "/var/run/docker.sock")
		}),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		sshCli.Close()
		return nil, err
	}

	return &DockerService{cli: cli, sshClient: sshCli}, nil
}

func (s *DockerService) Close() {
	s.cli.Close()
	if s.sshClient != nil {
		s.sshClient.Close()
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
	out, err := s.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(io.Discard, out)
	return err
}

func (s *DockerService) CreateNetwork(ctx context.Context, name, driver string) (string, error) {
	if driver == "" {
		driver = "bridge"
	}
	resp, err := s.cli.NetworkCreate(ctx, name, dockernetwork.CreateOptions{Driver: driver})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (s *DockerService) RemoveNetwork(ctx context.Context, dockerNetworkID string) error {
	return s.cli.NetworkRemove(ctx, dockerNetworkID)
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
		return "", err
	}

	for _, networkID := range cfg.Networks {
		if err := s.cli.NetworkConnect(ctx, networkID, resp.ID, &dockernetwork.EndpointSettings{}); err != nil {
			return "", fmt.Errorf("connect to network %s: %w", networkID, err)
		}
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
