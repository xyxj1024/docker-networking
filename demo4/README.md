## Basic Service Discovery in Go

The original blog post by [Abdulsametileri](https://github.com/Abdulsametileri) is [here](https://itnext.io/lets-implement-basic-service-discovery-using-go-d91c513883f6).

In this example, Docker containers *running on the local machine* are found through IP addresses, each taking the form of `http://localhost:xxxx`, where `xxxx` is the public port:

```go
// See: https://github.com/moby/moby/blob/master/api/types/port.go
type Port struct {
	// Host IP address that the container's port is mapped to
	IP string `json:"IP,omitempty"`
	// Port on the container
	PrivatePort uint16 `json:"PrivatePort"`
	// Port exposed on the host
	PublicPort uint16 `json:"PublicPort,omitempty"`
	// type
	Type string `json:"Type"`
}
```

This port number can be retrieved for each container via the following command:

```bash
docker inspect --format="{{json .}}" $CONT_ID | jq '.NetworkSettings.Ports["8080/tcp"][0].HostPort'
```

- Build Docker image for hello services:

    ```bash
    docker build services/ -t hello
    ```

- Build and run Go module:

    ```bash
    go mod init hello-service-discovery && go mod tidy
    go build
    go run hello-service-discovery
    ```

- Run containers in the background:

    ```bash
    docker run -d -p 8080:8080 hello && \
    docker run -d -p 8081:8080 hello && \
    docker run -d -p 8082:8080 hello && \
    docker run -d -p 8083:8080 hello
    ```

- Remove all containers:

    ```bash
    docker rm -f $(docker ps -a -q)
    ```

- Request hello service:

    ```bash
    hey http://localhost:3000/
    ```