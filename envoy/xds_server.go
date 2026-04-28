// xds_server.go — Minimal Envoy xDS (v3) control plane
//
// This program starts a gRPC server that implements Envoy's Aggregated Discovery
// Service (ADS) / xDS v3 API.  It serves static snapshots for:
//   - CDS  (Cluster Discovery Service)
//   - EDS  (Endpoint Discovery Service)
//   - LDS  (Listener Discovery Service)
//   - RDS  (Route Discovery Service)
//
// Envoy connects to this server and receives configuration dynamically, so you
// can change routes, endpoints and clusters at runtime without restarting Envoy.
//
// ─────────────────────────────────────────────────────────────────────────────
// Module setup (run once):
//   mkdir xds_server && cd xds_server
//   go mod init github.com/kleptic/xds_server
//   go get github.com/envoyproxy/go-control-plane@v0.12.0
//   go get google.golang.org/grpc@v1.62.0
//   go get google.golang.org/protobuf
//   go get github.com/golang/protobuf
// ─────────────────────────────────────────────────────────────────────────────
// Run:
//   go run xds_server.go --port 18000 --cluster my_service --endpoint 127.0.0.1:8080
//
// Envoy dynamic bootstrap (point to this server):
//   dynamic_resources:
//     ads_config:
//       api_type: GRPC
//       transport_api_version: V3
//       grpc_services:
//         - envoy_grpc:
//             cluster_name: xds_cluster
//     cds_config:
//       resource_api_version: V3
//       ads: {}
//     lds_config:
//       resource_api_version: V3
//       ads: {}
//
//   static_resources:
//     clusters:
//       - name: xds_cluster
//         type: STATIC
//         connect_timeout: 1s
//         http2_protocol_options: {}
//         load_assignment:
//           cluster_name: xds_cluster
//           endpoints:
//             - lb_endpoints:
//                 - endpoint:
//                     address:
//                       socket_address:
//                         address: 127.0.0.1
//                         port_value: 18000
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Logger
// ─────────────────────────────────────────────────────────────────────────────

type logger struct{}

func (l logger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (l logger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO]  "+format+"\n", args...)
}
func (l logger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN]  "+format+"\n", args...)
}
func (l logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// ─────────────────────────────────────────────────────────────────────────────
// Snapshot builder helpers
// ─────────────────────────────────────────────────────────────────────────────

// makeCluster builds a simple CDS cluster entry backed by EDS.
func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(500 * time.Millisecond),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ResourceApiVersion: core.ApiVersion_V3,
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
		LbPolicy: cluster.Cluster_ROUND_ROBIN,
	}
}

// makeEndpoint builds an EDS ClusterLoadAssignment for a single address:port.
func makeEndpoint(clusterName, address string, port uint32) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  address,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// makeRoute builds a simple RDS route configuration.
func makeRoute(routeName, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "local",
				Domains: []string{"*"},
				Routes: []*route.Route{
					{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: clusterName,
								},
								Timeout: durationpb.New(15 * time.Second),
							},
						},
					},
				},
			},
		},
	}
}

// makeHTTPListener builds an LDS listener with an HCM filter using RDS.
func makeHTTPListener(listenerName string, port uint32, routeName string) (*listener.Listener, error) {
	routerAny, err := anypb.New(&router.Router{})
	if err != nil {
		return nil, fmt.Errorf("marshal router filter: %w", err)
	}

	hcmConfig := &hcm.HttpConnectionManager{
		StatPrefix: "xds_hcm",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeName,
				ConfigSource: &core.ConfigSource{
					ResourceApiVersion: core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.Router,
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: routerAny,
				},
			},
		},
		GenerateRequestId: nil,
	}

	hcmAny, err := anypb.New(hcmConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal HCM: %w", err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{
			{
				Filters: []*listener.Filter{
					{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: hcmAny,
						},
					},
				},
			},
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Snapshot builder
// ─────────────────────────────────────────────────────────────────────────────

type EndpointSpec struct {
	Address string
	Port    uint32
}

func buildSnapshot(version string, clusterName string, endpoints []EndpointSpec, listenerPort uint32) (cachev3.Snapshot, error) {
	const routeName = "xds_route"
	const listenerName = "xds_listener"

	cls := makeCluster(clusterName)
	eds := makeEndpoint(clusterName, endpoints[0].Address, endpoints[0].Port)
	rds := makeRoute(routeName, clusterName)
	lds, err := makeHTTPListener(listenerName, listenerPort, routeName)
	if err != nil {
		return cachev3.Snapshot{}, err
	}

	snap, err := cachev3.NewSnapshot(
		version,
		map[types.ResponseType][]types.Resource{
			types.ClusterType:  {cls},
			types.EndpointType: {eds},
			types.RouteType:    {rds},
			types.ListenerType: {lds},
		},
	)
	if err != nil {
		return cachev3.Snapshot{}, fmt.Errorf("create snapshot: %w", err)
	}
	return snap, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// gRPC server registration
// ─────────────────────────────────────────────────────────────────────────────

func registerServices(grpcServer *grpc.Server, srv serverv3.Server) {
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, srv)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, srv)
}

// ─────────────────────────────────────────────────────────────────────────────
// main
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	port := flag.Int("port", 18000, "xDS gRPC server port")
	clusterName := flag.String("cluster", "backend_cluster", "Cluster name exposed to Envoy")
	endpointFlag := flag.String("endpoint", "127.0.0.1:8080", "Comma-separated host:port list for the cluster")
	listenerPort := flag.Uint("listener-port", 10000, "Listener port pushed to Envoy via LDS")
	nodeID := flag.String("node-id", "envoy-node", "Envoy node ID this snapshot targets (* for all)")
	flag.Parse()

	log := logger{}
	log.Infof("xDS server starting on port %d", *port)
	log.Infof("Serving cluster=%s, nodeID=%s", *clusterName, *nodeID)

	// Parse endpoints
	var eps []EndpointSpec
	for _, raw := range strings.Split(*endpointFlag, ",") {
		raw = strings.TrimSpace(raw)
		var host string
		var portVal uint32
		if _, err := fmt.Sscanf(raw, "%s", &raw); err != nil {
			continue
		}
		parts := strings.LastIndex(raw, ":")
		if parts < 0 {
			log.Errorf("invalid endpoint %q, expected host:port", raw)
			os.Exit(1)
		}
		host = raw[:parts]
		if _, err := fmt.Sscanf(raw[parts+1:], "%d", &portVal); err != nil {
			log.Errorf("invalid port in endpoint %q: %v", raw, err)
			os.Exit(1)
		}
		eps = append(eps, EndpointSpec{Address: host, Port: portVal})
	}
	if len(eps) == 0 {
		log.Errorf("no valid endpoints provided")
		os.Exit(1)
	}

	// Create snapshot cache
	cache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, log)

	// Build and set initial snapshot
	snap, err := buildSnapshot("1", *clusterName, eps, uint32(*listenerPort))
	if err != nil {
		log.Errorf("build snapshot: %v", err)
		os.Exit(1)
	}
	ctx := context.Background()
	if err := cache.SetSnapshot(ctx, *nodeID, &snap); err != nil {
		log.Errorf("set snapshot: %v", err)
		os.Exit(1)
	}
	log.Infof("Snapshot v1 set for node %q", *nodeID)

	// Create xDS server
	xdsSrv := serverv3.NewServer(ctx, cache, nil)

	// Create gRPC server with keepalive
	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(1000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
	)
	registerServices(grpcServer, xdsSrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Errorf("listen: %v", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT / SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Infof("Shutting down xDS server...")
		grpcServer.GracefulStop()
	}()

	log.Infof("xDS server listening on :%d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("gRPC serve: %v", err)
		os.Exit(1)
	}
}
