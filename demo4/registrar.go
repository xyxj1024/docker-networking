package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type Registrar struct {
	dockerClient    *Client
	serviceRegistry *Registry
}

const (
	HelloServiceImageName = "hello"
	ContainerRunningState = "running"
	ContainerKillState    = "kill"
	ContainerStartState   = "start"
)

func (r *Registrar) Init() error {
	containers, err := r.dockerClient.Client.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("ancestor", HelloServiceImageName),
			filters.Arg("status", ContainerRunningState),
		),
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		r.serviceRegistry.Add(c.ID, printContainerAddress(c.Ports[0].PublicPort))
	}

	return nil
}

func (r *Registrar) Observe() {
	msgCh, errCh := r.dockerClient.Client.Events(context.Background(), types.EventsOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("image", HelloServiceImageName),
			filters.Arg("event", "start"),
			filters.Arg("event", "kill"),
		),
	})

	for {
		select {
		case m := <-msgCh:
			fmt.Printf("State of the container %s is %sed\n", m.ID, m.Status)
			if m.Status == ContainerKillState {
				r.serviceRegistry.RemoveBackendByContainerId(m.ID)
			} else if m.Status == ContainerStartState {
				port, err := r.dockerClient.GetContainerPort(context.Background(), m.ID)
				if err != nil {
					fmt.Printf("Error getting newly started container port: %s\n", err.Error())
					continue
				}
				r.serviceRegistry.Add(m.ID, printContainerAddress(port))
			}
		case err := <-errCh:
			fmt.Println("Error Docker event channel", err.Error())
		}
	}
}

func printContainerAddress(port uint16) string {
	return fmt.Sprintf("http://localhost:%d", port)
}
