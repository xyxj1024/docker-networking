package main

import (
	"context"
	"flag"
	"time"

	"envoy-control-onl/pkg/callback"
	"envoy-control-onl/pkg/exec"

	"github.com/sirupsen/logrus"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	debug        bool
	port         uint
	pushInterval time.Duration
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.DurationVar(&pushInterval, "push-interval", 60*time.Second, "Time interval between Envoy config push")
}

func main() {
	flag.Parse()
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	ctx := context.Background()
	logrus.Printf("Starting control plane")

	signal := make(chan struct{})
	cb := &callback.Callbacks{
		Signal:         signal,
		Fetches:        0,
		Requests:       0,
		DeltaRequests:  0,
		DeltaResponses: 0,
	}

	config := cache.NewSnapshotCache(
		true, // enable ADS mode
		cache.IDHash{},
		nil,
	)
	srv := server.NewServer(ctx, config, cb)
	go exec.RunManagementServer(ctx, srv, port)
	<-signal
	for {
		time.Sleep(pushInterval)
		exec.GenerateSnapshot(ctx, config)
	}
}
