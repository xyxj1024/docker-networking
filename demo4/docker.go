package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Client struct {
	*client.Client
}

func NewDockerClient() (*Client, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(context.Background())
	return &Client{cli}, nil
}

func (c *Client) GetContainerPort(ctx context.Context, id string) (uint16, error) {
	containers, err := c.Client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("id", id),
		),
	})
	if len(containers) == 1 {
		return containers[0].Ports[0].PublicPort, nil
	}
	return 0, err
}
