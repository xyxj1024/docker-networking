package configresource

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/durationpb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

/* Function MakeCluster:
 * returns cluster resource configuration given cluster name, upstream host and upstream port.
 */
func MakeCluster(clusterName string, upstreamHost string, upstreamPort uint32) *cluster.Cluster {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating cluster with clusterName %s, upstreamHost %s", clusterName, upstreamHost)

	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(2 * time.Second),
		ClusterDiscoveryType: getClusterDiscoveryType(upstreamHost),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       MakeEndpoint(clusterName, upstreamHost, upstreamPort),
	}
}

/* Function MakeEndpoint:
 * returns endpoint resource configuration given cluster name, upstream host and upstream port.
 */
func MakeEndpoint(clusterName string, upstreamHost string, upstreamPort uint32) *endpoint.ClusterLoadAssignment {
	hst := &endpoint.Endpoint{
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  upstreamHost, // e.g., www.google.com; can also be a Docker service name
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: upstreamPort,
					},
				},
			},
		},
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: hst,
				},
			}},
		}},
	}
}

/* Helper function getClusterDiscoveryType:
 * returns a strict DNS type if the given string is not an IP address;
 * returns a static type, otherwise.
 */
func getClusterDiscoveryType(s string) *cluster.Cluster_Type {
	isIP := true
	if net.ParseIP(s) == nil {
		isIP = false
	}

	if isIP {
		return &cluster.Cluster_Type{Type: cluster.Cluster_STATIC}
	}
	return &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS}
}
