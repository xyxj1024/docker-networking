## Simple DNS Service for Docker Swarm

### Docker Swarm Mode

Docker Swarm mode enables multiple containers across multiple host machines to run together in a cluster (a "*swarm*"). A *node* is an instance of the Docker engine participating in the swarm.

Docker swarm uses overlay network. The `overlay` network driver creates a distributed network among multiple Docker daemon hosts. This network sits on top of (overlays) the host-specific networks, allowing containers connected to it (including swarm service containers) to communicate securely when encryption is enabled.

Services are deployed on a swarm using Compose files. Swarm extended the Compose format by adding a `deploy` key to each service that specifies how many instances of the service should be running and which nodes they should run on.

### Docker Swarm API

Information about nodes can be retrieved by calling the [`NodeList`](https://pkg.go.dev/github.com/docker/docker/client#Client.NodeList) method, which returns an array of `swarm.Node`s.

Labels can be stored outside a Docker swarm service's image or running containers using [`Config`](https://pkg.go.dev/github.com/docker/docker@latest/api/types/swarm#Config)s. We can access these labels as follows:

```go
func (c *swarm.Config) Labels() string {
    mapLabels := c.Spec.Annotations.Labels
    if mapLabels == nil {
        return ""
    }
    var joinLabels []string
    for k, v := range mapLabels {
        joinLabels = append(joinLabels, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(joinLabels, ",")
}

func (c *swarm.Config) Label(name string) string {
    if c.Spec.Annotations.Labels == nil {
        return ""
    }
    return c.Spec.Annotations.Labels[name]
}
```

Labels associated with a Docker swarm node can be accessed similarly, by changing `c *swarm.Config` in the above methods to `n *swarm.Node` (and `s *swarm.Service` for service-level labels). Node labels provide a flexible method of node organization which can be used to limit critical tasks to nodes that meet certain requirements. Run:

```bash
docker node update --label-add $KEY=$VAL $NODE
```

on a manager node to add label metadata to a node.

### Run Code

Full credit to [Martin Honermeyer](https://github.com/djmaze/swarmdns). A DNS service for Docker Swarm that returns the IP addresses of all active swarm nodes. The service works on manager nodes only.

```bash
# Generate go.mod and go.sum
go mod init docker-demo6 && go mod tidy

# Initialize a swarm
docker swarm init

# Build an image named "myswarmdns"
docker-compose build

# Run DNS service in the background
docker-compose up -d

# Send request
dig swm.example.com @localhost
```