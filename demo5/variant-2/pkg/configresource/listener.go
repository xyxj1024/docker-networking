package configresource

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

const (
	ListenerIPAddress                          = "0.0.0.0"
	ControlPlaneCluster                        = "control_plane"
	HTTPIdleTimeout                            = 1 * time.Hour
	RequestTimeout                             = 5 * time.Minute
	MaxConcurrentHTTP2Streams                  = 100
	InitialDownstreamHTTP2StreamWindowSize     = 65536   // 64 KiB
	InitialDownstreamHTTP2ConnectionWindowSize = 1048576 // 1 MiB
)

/* Function MakeHTTPListener:
 * returns listener resource configuration given listener name, route configuration name and listener port.
 */
func MakeHTTPListener(listenerName string, listenerPort uint32, routeConfigName string) *listener.Listener {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating listener with listenerName " + listenerName)

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeConfigName,
				ConfigSource:    makeConfigSource(),
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: messageToAny(&router.Router{}),
			},
		}},
		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			IdleTimeout:                  durationpb.New(HTTPIdleTimeout),
			HeadersWithUnderscoresAction: core.HttpProtocolOptions_REJECT_REQUEST,
		},
		Http2ProtocolOptions: &core.Http2ProtocolOptions{
			MaxConcurrentStreams:        &wrappers.UInt32Value{Value: uint32(MaxConcurrentHTTP2Streams)},
			InitialStreamWindowSize:     &wrappers.UInt32Value{Value: uint32(InitialDownstreamHTTP2StreamWindowSize)},
			InitialConnectionWindowSize: &wrappers.UInt32Value{Value: uint32(InitialDownstreamHTTP2ConnectionWindowSize)},
		},
		StreamIdleTimeout: durationpb.New(RequestTimeout),
		RequestTimeout:    durationpb.New(RequestTimeout),
	}

	pbst, err := anypb.New(manager)
	if err != nil {
		logrus.Fatal(err)
	}

	return &listener.Listener{
		Name: listenerName, // e.g., "listener_0"
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  ListenerIPAddress, // listen on all IP addresses
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: listenerPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: ControlPlaneCluster}, // xDS cluster
				},
			}},
		},
	}
	return source
}

func messageToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		TypeUrl: "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

func messageToAny(msg proto.Message) *anypb.Any {
	out, err := messageToAnyWithError(msg)
	if err != nil {
		logrus.Error(fmt.Sprintf("Error marshaling Any %s: %v", prototext.Format(msg), err))
		return nil
	}
	return out
}
