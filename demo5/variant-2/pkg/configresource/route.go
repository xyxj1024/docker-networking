package configresource

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

/* Function MakeRoute:
 * returns route resource configuration given route configuration name,
 * virtual host name, cluster name, upstream host, path prefix and request timeout.
 */
func MakeRoute(routeConfigName, virtualHostName, clusterName, upstreamHost, pathPrefix string, requestTimeout time.Duration) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeConfigName, // e.g., "local_route"
		VirtualHosts: []*route.VirtualHost{{
			Name:    virtualHostName, // e.g., "local_service"
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: pathPrefix,
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
						// PrefixRewrite: "/robots.txt", /* removing this line causes no harm */
						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: upstreamHost,
						},
						// https://github.com/envoyproxy/envoy/issues/8517#issuecomment-540225144
						IdleTimeout: durationpb.New(requestTimeout),
						Timeout:     durationpb.New(requestTimeout),
					},
				},
			}},
		}},
	}
}
