## An Experimental Envoy Control Plane for ONL

### Configuration Update Mechanisms

This control plane design supports two mechanisms for Envoy configuration update: "push" and "pull".

For the "push" mechanism, a control plane instance will periodically read config data from a database (not necessarily hosted on the same machine as the control plane instance) and feed config data to the corresponding Envoy nodes.

For the "pull" mechanism, a user can directly send update requests to a control plane instance (on behalf of some Envoy nodes). Upon receiving these requests, the control plane will serve certain config data for the corresponding Envoy nodes.

### Code Structure

```bash
go mod init envoy-control-onl && go mod tidy
```

The `callback` module is required for the xDS server. Do not modify unless necessary.

The `configresource` module contains all the functions we need to build Envoy resources (i.e., what we specify in Envoy config YAML files).

The `RunManagementServer` function defined in `./pkg/exec/xdsserver.go` runs in a goroutine to accept `DiscoveryRequest`s from connected Envoy nodes.