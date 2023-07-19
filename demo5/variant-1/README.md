## Preparation

- (Required) Initialize Go module:

    ```bash
    go mod init envoy-control && go mod tidy
    ```

- (Optional) To run without containers, `cd` into the `xds` folder and:
    
    ```bash
    go build    # which will generate an executable "envoy-control"
    ```

Keep the `go.mod` and `go.sum` files untouched.

## Run without Docker

Make sure that the listener's address of the bootstrap configuration files for Envoy proxies is set to `127.0.0.1`.

```bash
# Terminal 1: Start Envoy-1
cd path/to/demo/envoy
envoy -c envoy-1/envoy_google.yaml -l debug

# Terminal 2: Start Envoy-2
cd path/to/demo/envoy
envoy -c envoy-2/envoy_wustl.yaml -l debug

# Terminal 3: Start Envoy-3
cd path/to/demo/envoy
envoy -c envoy-3/envoy_bbc.yaml -l debug

# Terminal 4: Start control plane
cd path/to/demo/xds
go run envoy-control

# Terminal 5: Access proxy. Replace port number 10001 with the listener's port we give
curl localhost:10001
```

## Run with Docker Compose

Make sure that the listener's address of the bootstrap configuration files for Envoy proxies is set to `xds` (the corresponding container name), the container name associated with the xDS server.

First, inside the `demo` folder, run `docker compose build`. Then, `docker compose up` all `envoy-x`s.

Run the xDS server:

```bash
docker compose run xds
```

This command should lead us to container `xds`'s shell:

```bash
go run envoy-control
```

The xDS server will make configuration update for an Envoy proxy every 30 seconds.

## Problems

If I comment out the `entrypoint` line inside `docker-compose.yaml`, and run `docker compose up xds`, the control plane seems to be able to receive the first `DiscoveryRequest` and keep spinning:

![run_xds](run_xds.png)

If I uncomment the `entrypoint` line, and run `go run envoy-control` inside the container `xds`'s shell, the address of `xds` cannot be resolved.

### Update:

The problem is solved when I change the `resource.go` program code to read from a JSON file rather than from standard input.

To run the xDS server natively (rather than inside a container), change the listener's address of the bootstrap configuration files for Envoy proxies to `host.docker.internal`.