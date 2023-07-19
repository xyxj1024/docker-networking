package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	debug       bool
	port        uint
	gatewayPort uint
	mode        string
	version     int32
	config      cache.SnapshotCache
)

const (
	Ads                      = "ads"
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.UintVar(&gatewayPort, "gateway", 18001, "Management server port for HTTP gateway")
	flag.StringVar(&mode, "ads", Ads, "Management server type (ads only now)")
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}

func RunManagementServer(ctx context.Context, server server.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	registerServer(grpcServer, server)
	logrus.WithFields(logrus.Fields{"port": port}).Info("Management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Error(err)
		}
	}()
	<-ctx.Done()
	grpcServer.GracefulStop()
}

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	logrus.Printf("Starting control plane")

	signal := make(chan struct{})
	cb := &Callbacks{
		Signal:         signal,
		Fetches:        0,
		Requests:       0,
		DeltaRequests:  0,
		DeltaResponses: 0,
	}

	config = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
	srv := server.NewServer(ctx, config, cb)
	go RunManagementServer(ctx, srv, port)
	<-signal
	for {
		time.Sleep(30 * time.Second)
		GenerateSnapshot(ctx, config)
	}
}
