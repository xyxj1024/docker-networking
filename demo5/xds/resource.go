package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	listenerName    = "listener_0"
	virtualHostName = "local_service"
	routeConfigName = "local_route"
	// secretName      = "server_cert"
)

func makeCluster(clusterName string, upstreamHost string) *cluster.Cluster {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating cluster with clusterName %s, upstreamHost %s", clusterName, upstreamHost)

	hst := &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  upstreamHost,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(443),
				},
			},
		},
	}
	/*
		uctx := &tls.UpstreamTlsContext{}
		tctx, err := anypb.New(uctx)
		if err != nil {
			logrus.Fatal(err)
		}
	*/

	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(2 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{{
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: hst,
							}},
					},
				},
			}},
		},
		/*
			TransportSocket: &core.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: tctx,
				},
			},
		*/
	}
}

func makeRoute(clusterName string, upstreamHost string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeConfigName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    virtualHostName,
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
						PrefixRewrite: "/robots.txt",
						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: upstreamHost,
						},
					},
				},
			}},
		}},
	}
}

func makeHTTPListener( /* pub []byte, priv []byte, */ listenerPort uint32) *listener.Listener {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating listener with listenerName " + listenerName)

	routerConfig, _ := anypb.New(&router.Router{})
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: routeConfigName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}

	pbst, err := anypb.New(manager)
	if err != nil {
		logrus.Fatal(err)
	}
	/*
		sdsTls := &tls.DownstreamTlsContext{
			CommonTlsContext: &tls.CommonTlsContext{
				TlsCertificates: []*tls.TlsCertificate{{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(pub)},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(priv)},
					},
				}},
			},
		}

		scfg, err := anypb.New(sdsTls)
		if err != nil {
			logrus.Fatal(err)
		}
	*/

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
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
			/*
				TransportSocket: &core.TransportSocket{
					Name: "envoy.transport_sockets.tls",
					ConfigType: &core.TransportSocket_TypedConfig{
						TypedConfig: scfg,
					},
				},
			*/
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
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}

/*
func makeSecret(pub []byte, priv []byte) *tls.Secret {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating secret with secretName " + secretName)
	return &tls.Secret{
		Name: secretName,
		Type: &tls.Secret_TlsCertificate{
			TlsCertificate: &tls.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(pub)},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(priv)},
				},
			},
		},
	}
}
*/

func GenerateSnapshot(ctx context.Context, config cache.SnapshotCache) {
	num := len(config.GetStatusKeys())
	logrus.Infof("%d connected nodes\n", num)
	if num > 0 {
		for i := 0; i < num; i++ {
			nodeId := config.GetStatusKeys()[i]
			pattern := regexp.MustCompile(`^test-id-([A-z]*)$`)
			service := pattern.FindStringSubmatch(nodeId)[1]
			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot " + fmt.Sprint(version) +
				" for nodeID " + fmt.Sprint(nodeId) +
				", cluster service_" + service)

			atomic.AddInt32(&version, 1)

			/*
				pub, err := os.ReadFile("certs/envoy-proxy-server.crt")
				if err != nil {
					logrus.Fatal(err)
				}
				priv, err := os.ReadFile("certs/envoy-proxy-server.key")
				if err != nil {
					logrus.Fatal(err)
				}
			*/

			jsonFile, err := os.ReadFile("input.json")
			if err != nil {
				logrus.Fatal(err)
			}
			var input Input
			json.Unmarshal([]byte(jsonFile), &input)
			for i := 0; i < len(input.Data); i++ {
				if strings.Contains(input.Data[i].ClusterName, service) {
					resources := make(map[string][]types.Resource, 3)
					resources[resource.ClusterType] = []types.Resource{makeCluster(input.Data[i].ClusterName, input.Data[i].UpstreamHost)}
					resources[resource.RouteType] = []types.Resource{makeRoute(input.Data[i].ClusterName, input.Data[i].UpstreamHost)}
					resources[resource.ListenerType] = []types.Resource{makeHTTPListener( /* pub, priv,*/ input.Data[i].ListenerPort)}
					// resources[resource.SecretType] = []types.Resource{makeSecret(pub, priv)}
					snap, _ := cache.NewSnapshot(fmt.Sprint(version), resources)
					if err := snap.Consistent(); err != nil {
						logrus.Errorf("Snapshot inconsistency: %+v\n%+v", snap, err)
						os.Exit(1)
					}

					if err = config.SetSnapshot(ctx, nodeId, snap); err != nil {
						logrus.Fatalf("Snapshot error %q for %+v", err, snap)
					}

					logrus.Infof("Snapshot served: %+v", snap)

					break
				}
			}
			time.Sleep(30 * time.Second)
		}
	}
}
