package callback

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

type Callbacks struct {
	Signal         chan struct{}
	Debug          bool
	Fetches        int
	Requests       int
	DeltaRequests  int
	DeltaResponses int
	mu             sync.Mutex // only one goroutine at a time can access callback structure
}

var _ server.Callbacks = &Callbacks{}

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	logrus.WithFields(logrus.Fields{"fetches": cb.Fetches, "requests": cb.Requests}).Info("Report() callbacks")
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	logrus.Infof("OnStreamOpen %d of type %v", id, typ)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64, node *core.Node) {
	logrus.Infof("OnStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	logrus.Infof("OnDeltaStreamOpen %d of type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	logrus.Infof("OnDeltaStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnStreamRequest(id int64, req *discoverygrpc.DiscoveryRequest) error {
	logrus.Infof("OnStreamRequest %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnStreamResponse(ctx context.Context, id int64, req *discoverygrpc.DiscoveryRequest, res *discoverygrpc.DiscoveryResponse) {
	logrus.Infof("OnStreamResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.Report()
}

func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discoverygrpc.DeltaDiscoveryRequest, res *discoverygrpc.DeltaDiscoveryResponse) {
	logrus.Infof("OnStreamDeltaResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}

func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discoverygrpc.DeltaDiscoveryRequest) error {
	logrus.Infof("OnStreamDeltaRequest... %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaRequests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discoverygrpc.DiscoveryRequest) error {
	logrus.Infof("OnFetchRequest... Request [%v]", req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchResponse(req *discoverygrpc.DiscoveryRequest, res *discoverygrpc.DiscoveryResponse) {
	logrus.Infof("OnFetchResponse... Request [%v], Response [%v]", req.TypeUrl, res.TypeUrl)
}
