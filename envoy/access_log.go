package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// AccessLogFormat selects the output format
type AccessLogFormat string

const (
	FormatText AccessLogFormat = "text"
	FormatJSON AccessLogFormat = "json"
	FormatGRPC AccessLogFormat = "grpc" // sends to a gRPC ALS server
)

// AccessLogConfig holds parameters for generating an Envoy access log config
type AccessLogConfig struct {
	ClusterName  string
	ListenerPort int
	AdminPort    int
	Format       AccessLogFormat
	LogPath      string // for text/json: file path, e.g. /dev/stdout
	ALSHost      string // for grpc: ALS server host
	ALSPort      int    // for grpc: ALS server port
}

// buildAccessLogFilter returns the http_filters access log block
func buildAccessLogFilter(cfg AccessLogConfig) []map[string]interface{} {
	switch cfg.Format {
	case FormatJSON:
		return []map[string]interface{}{
			{
				"name": "envoy.access_loggers.file",
				"typed_config": map[string]interface{}{
					"@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
					"path":  cfg.LogPath,
					"log_format": map[string]interface{}{
						"json_format": map[string]interface{}{
							"timestamp":             "%START_TIME%",
							"method":                "%REQ(:METHOD)%",
							"path":                  "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%",
							"protocol":              "%PROTOCOL%",
							"response_code":         "%RESPONSE_CODE%",
							"response_flags":        "%RESPONSE_FLAGS%",
							"bytes_received":        "%BYTES_RECEIVED%",
							"bytes_sent":            "%BYTES_SENT%",
							"duration_ms":           "%DURATION%",
							"upstream_host":         "%UPSTREAM_HOST%",
							"upstream_cluster":      "%UPSTREAM_CLUSTER%",
							"upstream_service_time": "%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%",
							"x_forwarded_for":       "%REQ(X-FORWARDED-FOR)%",
							"user_agent":            "%REQ(USER-AGENT)%",
							"request_id":            "%REQ(X-REQUEST-ID)%",
							"authority":             "%REQ(:AUTHORITY)%",
							"trace_id":              "%REQ(X-B3-TRACEID)%",
						},
					},
				},
			},
		}
	case FormatGRPC:
		return []map[string]interface{}{
			{
				"name": "envoy.access_loggers.http_grpc",
				"typed_config": map[string]interface{}{
					"@type": "type.googleapis.com/envoy.extensions.access_loggers.grpc.v3.HttpGrpcAccessLogConfig",
					"common_config": map[string]interface{}{
						"log_name": "envoy-access-log",
						"grpc_service": map[string]interface{}{
							"envoy_grpc": map[string]interface{}{
								"cluster_name": "als_cluster",
							},
						},
						"transport_api_version": "V3",
					},
					"additional_request_headers_to_log":  []string{"x-request-id", "x-b3-traceid"},
					"additional_response_headers_to_log": []string{"x-envoy-upstream-service-time"},
				},
			},
		}
	default: // text
		return []map[string]interface{}{
			{
				"name": "envoy.access_loggers.file",
				"typed_config": map[string]interface{}{
					"@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
					"path":  cfg.LogPath,
					"log_format": map[string]interface{}{
						"text_format_source": map[string]interface{}{
							"inline_string": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION%ms \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\"\n",
						},
					},
				},
			},
		}
	}
}

// GenerateAccessLogConfig produces an Envoy static config with access logging
func GenerateAccessLogConfig(cfg AccessLogConfig) map[string]interface{} {
	accessLog := buildAccessLogFilter(cfg)

	clusters := []map[string]interface{}{
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
	}

	// Add ALS cluster for gRPC mode
	if cfg.Format == FormatGRPC {
		clusters = append(clusters, map[string]interface{}{
			"name":                   "als_cluster",
			"connect_timeout":        "1s",
			"type":                   "STRICT_DNS",
			"lb_policy":              "ROUND_ROBIN",
			"http2_protocol_options": map[string]interface{}{},
			"load_assignment": map[string]interface{}{
				"cluster_name": "als_cluster",
				"endpoints": []map[string]interface{}{
					{
						"lb_endpoints": []map[string]interface{}{
							{
								"endpoint": map[string]interface{}{
									"address": map[string]interface{}{
										"socket_address": map[string]interface{}{
											"address":    cfg.ALSHost,
											"port_value": cfg.ALSPort,
										},
									},
								},
							},
						},
					},
				},
			},
		})
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
									"name": "envoy.filters.network.http_connection_manager",
									"typed_config": map[string]interface{}{
										"@type":               "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
										"stat_prefix":         "ingress_http",
										"generate_request_id": true,
										"access_log":          accessLog,
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
									},
								},
							},
						},
					},
				},
			},
			"clusters": clusters,
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

func printAccessLogGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY ACCESS LOG GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[FORMATS]")
	fmt.Println(strings.Repeat("-", 40))
	formats := [][]string{
		{"text", "Apache/Nginx-style text log. Human-readable, easy to grep."},
		{"json", "Structured JSON per-request. Best for log aggregation (Loki, Elasticsearch)."},
		{"grpc", "Stream to an ALS (Access Log Service) gRPC server in real time."},
	}
	for _, f := range formats {
		fmt.Printf("  %-10s %s\n", f[0], f[1])
	}

	fmt.Println("\n[KEY FORMAT VARIABLES]")
	fmt.Println(strings.Repeat("-", 40))
	vars := [][]string{
		{"%START_TIME%", "Request start time (ISO 8601)"},
		{"%REQ(:METHOD)%", "HTTP method"},
		{"%REQ(:PATH)%", "Request path"},
		{"%PROTOCOL%", "HTTP/1.1 or HTTP/2"},
		{"%RESPONSE_CODE%", "HTTP response status code"},
		{"%RESPONSE_FLAGS%", "Envoy response flags (UF=upstream failure, etc.)"},
		{"%DURATION%", "Total request duration in ms"},
		{"%BYTES_RECEIVED%", "Request body bytes"},
		{"%BYTES_SENT%", "Response body bytes"},
		{"%UPSTREAM_HOST%", "Host:port of upstream selected"},
		{"%UPSTREAM_CLUSTER%", "Cluster name"},
		{"%REQ(X-REQUEST-ID)%", "Correlation ID"},
		{"%REQ(X-B3-TRACEID)%", "Distributed trace ID"},
	}
	for _, v := range vars {
		fmt.Printf("  %-35s %s\n", v[0], v[1])
	}

	fmt.Println("\n[RESPONSE FLAGS REFERENCE]")
	fmt.Println(strings.Repeat("-", 40))
	flags := [][]string{
		{"UF", "Upstream connection failure"},
		{"UO", "Upstream overflow (circuit breaker)"},
		{"NR", "No route configured"},
		{"URX", "Upstream retry limit exceeded"},
		{"RL", "Rate limited"},
		{"UAEX", "Unauthorized - ext authz denied"},
		{"DPE", "Downstream protocol error"},
	}
	for _, f := range flags {
		fmt.Printf("  %-10s %s\n", f[0], f[1])
	}

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	format := flag.String("format", "json", "Access log format: text | json | grpc")
	logPath := flag.String("log-path", "/dev/stdout", "Log file path (text/json formats)")
	alsHost := flag.String("als-host", "als.default.svc.cluster.local", "ALS gRPC server host (grpc format)")
	alsPort := flag.Int("als-port", 9090, "ALS gRPC server port (grpc format)")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print access log guide")
	flag.Parse()

	if *guide {
		printAccessLogGuide()
		return
	}

	cfg := AccessLogConfig{
		ClusterName:  *cluster,
		ListenerPort: *listenerPort,
		AdminPort:    *adminPort,
		Format:       AccessLogFormat(*format),
		LogPath:      *logPath,
		ALSHost:      *alsHost,
		ALSPort:      *alsPort,
	}

	config := GenerateAccessLogConfig(cfg)

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
