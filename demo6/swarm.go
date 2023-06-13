package main

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
)

type SwarmNode struct {
	Ip        string
	Hostname  string
	IsManager bool
}

type Client interface {
	ListActiveNodes() ([]SwarmNode, error)
}

type SwarmClient struct {
	api *dockerclient.Client
}

func NewClient() (Client, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return SwarmClient{api: cli}, nil
}

func (cli SwarmClient) ListActiveNodes() ([]SwarmNode, error) {
	var listOptions types.NodeListOptions
	apiNodes, err := cli.api.NodeList(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	var (
		nodes []SwarmNode
		ip    string
	)
	for _, n := range apiNodes {
		if n.Status.State == swarm.NodeStateReady {
			if publicIp, ok := n.Spec.Annotations.Labels["public-ip"]; ok {
				ip = publicIp
			} else if n.Status.Addr == "0.0.0.0" {
				ip = getIpFromAddr(n.ManagerStatus.Addr)
				if err != nil {
					return nil, err
				}
			} else {
				ip = n.Status.Addr
			}

			nodes = append(nodes, SwarmNode{
				Ip:        ip,
				Hostname:  getHostname(n),
				IsManager: n.ManagerStatus != nil,
			})
		}
	}

	return nodes, nil
}

func getHostname(node swarm.Node) string {
	hostname := node.Spec.Annotations.Labels["hostname"]
	if hostname == "" {
		hostname = node.Description.Hostname
	}
	return hostname
}

func getIpFromAddr(addr string) string {
	ipPort := strings.Split(addr, ":")
	return ipPort[0]
}
