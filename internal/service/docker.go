package service

import (
	"context"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerService struct {
	cli *client.Client
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
