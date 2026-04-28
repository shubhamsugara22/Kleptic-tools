package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// TracingBackend selects which tracing provider to configure
type TracingBackend string

const (
	Zipkin  TracingBackend = "zipkin"
	Jaeger  TracingBackend = "jaeger"
	OpenTel TracingBackend = "opentelemetry"
)

// TracingConfig holds parameters for generating an Envoy tracing config
type TracingConfig struct {
	ClusterName      string
	ListenerPort     int
	AdminPort        int
	Backend          TracingBackend
	CollectorHost    string
	CollectorPort    int
	ServiceName      string
	SamplingRate     float64 // 0.0 to 100.0
	MaxPathTagLength int
}

// buildTracingProvider returns the tracing provider block for a given backend
func buildTracingProvider(cfg TracingConfig) map[string]interface{} {
	switch cfg.Backend {
	case Zipkin:
		return map[string]interface{}{
			"name": "envoy.tracers.zipkin",
			"typed_config": map[string]interface{}{
				"@type":                      "type.googleapis.com/envoy.config.trace.v3.ZipkinConfig",
				"collector_cluster":          "tracing_cluster",
				"collector_endpoint":         "/api/v2/spans",
				"shared_span_context":        false,
				"collector_endpoint_version": "HTTP_JSON",
			},
		}
	case Jaeger:
		return map[string]interface{}{
			"name": "envoy.tracers.zipkin",
			"typed_config": map[string]interface{}{
				"@type":              "type.googleapis.com/envoy.config.trace.v3.ZipkinConfig",
				"collector_cluster":  "tracing_cluster",
				"collector_endpoint": "/api/v2/spans",
				// Jaeger accepts Zipkin v2 JSON over HTTP
				"collector_endpoint_version": "HTTP_JSON",
				"shared_span_context":        true,
			},
		}
	case OpenTel:
		return map[string]interface{}{
			"name": "envoy.tracers.opentelemetry",
			"typed_config": map[string]interface{}{
				"@type": "type.googleapis.com/envoy.config.trace.v3.OpenTelemetryConfig",
				"grpc_service": map[string]interface{}{
					"envoy_grpc": map[string]interface{}{
						"cluster_name": "tracing_cluster",
					},
					"timeout": "0.250s",
				},
				"service_name": cfg.ServiceName,
			},
		}
	default:
		return map[string]interface{}{}
	}
}

// GenerateTracingConfig produces an Envoy static config with distributed tracing
func GenerateTracingConfig(cfg TracingConfig) map[string]interface{} {
	tracingProvider := buildTracingProvider(cfg)

	tracingCluster := map[string]interface{}{
		"name":            "tracing_cluster",
		"connect_timeout": "1s",
		"type":            "LOGICAL_DNS",
		"lb_policy":       "ROUND_ROBIN",
		"load_assignment": map[string]interface{}{
			"cluster_name": "tracing_cluster",
			"endpoints": []map[string]interface{}{
				{
					"lb_endpoints": []map[string]interface{}{
						{
							"endpoint": map[string]interface{}{
								"address": map[string]interface{}{
									"socket_address": map[string]interface{}{
										"address":    cfg.CollectorHost,
										"port_value": cfg.CollectorPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// OpenTelemetry uses gRPC (HTTP/2)
	if cfg.Backend == OpenTel {
		tracingCluster["http2_protocol_options"] = map[string]interface{}{}
	}

	hcm := map[string]interface{}{
		"@type":               "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
		"stat_prefix":         "ingress_http",
		"generate_request_id": true,
		"tracing": map[string]interface{}{
			"provider": tracingProvider,
			"client_sampling": map[string]interface{}{
				"value": cfg.SamplingRate,
			},
			"random_sampling": map[string]interface{}{
				"value": cfg.SamplingRate,
			},
			"overall_sampling": map[string]interface{}{
				"value": cfg.SamplingRate,
			},
			"verbose":             false,
			"max_path_tag_length": cfg.MaxPathTagLength,
			"custom_tags": []map[string]interface{}{
				{
					"tag": "service.name",
					"literal": map[string]string{
						"value": cfg.ServiceName,
					},
				},
				{
					"tag": "node.id",
					"environment": map[string]string{
						"name":          "HOSTNAME",
						"default_value": "unknown",
					},
				},
			},
		},
		"route_config": map[string]interface{}{
			"name": "local_route",
			"virtual_hosts": []map[string]interface{}{
				{
					"name":    "backend",
					"domains": []string{"*"},
					"routes": []map[string]interface{}{
						{
							"match": map[string]string{"prefix": "/"},
							"route": map[string]string{"cluster": cfg.ClusterName},
							"decorator": map[string]string{
								"operation": "checkAvailability",
							},
						},
					},
				},
			},
		},
		"http_filters": []map[string]interface{}{
			{
				"name": "envoy.filters.http.router",
				"typed_config": map[string]interface{}{
					"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
				},
			},
		},
	}

	return map[string]interface{}{
		"static_resources": map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"name": "main_listener",
					"address": map[string]interface{}{
						"socket_address": map[string]interface{}{
							"address":    "0.0.0.0",
							"port_value": cfg.ListenerPort,
						},
					},
					"filter_chains": []map[string]interface{}{
						{
							"filters": []map[string]interface{}{
								{
									"name":         "envoy.filters.network.http_connection_manager",
									"typed_config": hcm,
								},
							},
						},
					},
				},
			},
			"clusters": []map[string]interface{}{
				{
					"name":            cfg.ClusterName,
					"connect_timeout": "0.5s",
					"type":            "LOGICAL_DNS",
					"lb_policy":       "ROUND_ROBIN",
					"load_assignment": map[string]interface{}{
						"cluster_name": cfg.ClusterName,
						"endpoints": []map[string]interface{}{
							{
								"lb_endpoints": []map[string]interface{}{
									{
										"endpoint": map[string]interface{}{
											"address": map[string]interface{}{
												"socket_address": map[string]interface{}{
													"address":    "service.example.com",
													"port_value": 8080,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				tracingCluster,
			},
		},
		"admin": map[string]interface{}{
			"address": map[string]interface{}{
				"socket_address": map[string]interface{}{
					"address":    "0.0.0.0",
					"port_value": cfg.AdminPort,
				},
			},
		},
	}
}

func printTracingGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY DISTRIBUTED TRACING GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[BACKENDS]")
	fmt.Println(strings.Repeat("-", 40))
	backends := [][]string{
		{"zipkin", "HTTP/JSON to Zipkin v2 API. Lightweight, good for small clusters."},
		{"jaeger", "Zipkin-compatible HTTP/JSON. Jaeger accepts the same wire format."},
		{"opentelemetry", "gRPC/OTLP. Vendor-neutral; works with Jaeger, Tempo, Honeycomb, etc."},
	}
	for _, b := range backends {
		fmt.Printf("  %-20s %s\n", b[0], b[1])
	}

	fmt.Println("\n[SAMPLING]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  client_sampling  : % of requests where client requests tracing (x-b3-sampled)")
	fmt.Println("  random_sampling  : % of NEW traces Envoy samples (when no upstream decision)")
	fmt.Println("  overall_sampling : Hard cap — never exceed this % regardless of other settings")
	fmt.Println("")
	fmt.Println("  Tip: set all three to 100.0 during dev/testing, then tune random_sampling down")
	fmt.Println("       to 1-10% in production to reduce overhead.")

	fmt.Println("\n[PROPAGATION HEADERS]")
	fmt.Println(strings.Repeat("-", 40))
	headers := [][]string{
		{"x-b3-traceid", "128-bit trace identifier"},
		{"x-b3-spanid", "64-bit span identifier"},
		{"x-b3-parentspanid", "Parent span (absent for root)"},
		{"x-b3-sampled", "1=record, 0=discard"},
		{"x-request-id", "Envoy-generated UUID (auto-added when generate_request_id=true)"},
		{"traceparent", "W3C Trace Context header (OpenTelemetry)"},
	}
	for _, h := range headers {
		fmt.Printf("  %-25s %s\n", h[0], h[1])
	}

	fmt.Println("\n[QUICK START DOCKER]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # Zipkin all-in-one")
	fmt.Println("  docker run -d -p 9411:9411 openzipkin/zipkin")
	fmt.Println("")
	fmt.Println("  # Jaeger all-in-one")
	fmt.Println("  docker run -d -p 9411:9411 jaegertracing/all-in-one")
	fmt.Println("")
	fmt.Println("  # OpenTelemetry Collector")
	fmt.Println("  docker run -d -p 4317:4317 otel/opentelemetry-collector")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	backend := flag.String("backend", "zipkin", "Tracing backend: zipkin | jaeger | opentelemetry")
	collectorHost := flag.String("collector-host", "zipkin.default.svc.cluster.local", "Tracing collector host")
	collectorPort := flag.Int("collector-port", 9411, "Tracing collector port")
	serviceName := flag.String("service-name", "envoy-proxy", "Service name tag in traces")
	samplingRate := flag.Float64("sampling-rate", 100.0, "Sampling rate 0.0-100.0")
	maxPathLen := flag.Int("max-path-len", 256, "Max URL path length captured in trace tags")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print tracing guide")
	flag.Parse()

	if *guide {
		printTracingGuide()
		return
	}

	cfg := TracingConfig{
		ClusterName:      *cluster,
		ListenerPort:     *listenerPort,
		AdminPort:        *adminPort,
		Backend:          TracingBackend(*backend),
		CollectorHost:    *collectorHost,
		CollectorPort:    *collectorPort,
		ServiceName:      *serviceName,
		SamplingRate:     *samplingRate,
		MaxPathTagLength: *maxPathLen,
	}

	config := GenerateTracingConfig(cfg)

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling config: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, b, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config written to %s\n", *output)
	} else {
		fmt.Println(string(b))
	}
}
