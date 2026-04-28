package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// RetryPolicy defines per-route retry behaviour
type RetryPolicy struct {
	RetryOn                  string // e.g. "5xx,gateway-error,reset,connect-failure,retriable-4xx"
	NumRetries               int
	PerTryTimeout            string // e.g. "1s"
	RetryHostPredicate       []string
	HostSelectionMaxAttempts int
}

// TimeoutPolicy defines per-route and per-cluster timeouts
type TimeoutPolicy struct {
	RouteTimeout      string // total request timeout, e.g. "15s"
	IdleTimeout       string // idle stream timeout, e.g. "60s"
	ConnectTimeout    string // upstream connect timeout, e.g. "0.5s"
	PerTryTimeout     string // per-retry attempt timeout, e.g. "3s"
	MaxStreamDuration string // e.g. "300s"
}

// RetriesTimeoutsConfig is the top-level config
type RetriesTimeoutsConfig struct {
	ClusterName  string
	ListenerPort int
	AdminPort    int
	Retry        RetryPolicy
	Timeout      TimeoutPolicy
}

// GenerateRetriesTimeoutsConfig produces an Envoy static config with retry + timeout policies
func GenerateRetriesTimeoutsConfig(cfg RetriesTimeoutsConfig) map[string]interface{} {
	retryPolicy := map[string]interface{}{
		"retry_on":                          cfg.Retry.RetryOn,
		"num_retries":                       cfg.Retry.NumRetries,
		"per_try_timeout":                   cfg.Retry.PerTryTimeout,
		"host_selection_retry_max_attempts": fmt.Sprintf("%d", cfg.Retry.HostSelectionMaxAttempts),
		"retry_back_off": map[string]interface{}{
			"base_interval": "0.25s",
			"max_interval":  "10s",
		},
	}

	if len(cfg.Retry.RetryHostPredicate) > 0 {
		predicates := []map[string]string{}
		for _, p := range cfg.Retry.RetryHostPredicate {
			predicates = append(predicates, map[string]string{"name": p})
		}
		retryPolicy["retry_host_predicate"] = predicates
	}

	route := map[string]interface{}{
		"cluster":      cfg.ClusterName,
		"timeout":      cfg.Timeout.RouteTimeout,
		"retry_policy": retryPolicy,
	}

	if cfg.Timeout.IdleTimeout != "" {
		route["idle_timeout"] = cfg.Timeout.IdleTimeout
	}
	if cfg.Timeout.MaxStreamDuration != "" {
		route["max_stream_duration"] = map[string]string{
			"max_stream_duration": cfg.Timeout.MaxStreamDuration,
		}
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
										"stream_idle_timeout": cfg.Timeout.IdleTimeout,
										"request_timeout":     cfg.Timeout.RouteTimeout,
										"route_config": map[string]interface{}{
											"name": "local_route",
											"virtual_hosts": []map[string]interface{}{
												{
													"name":    "backend",
													"domains": []string{"*"},
													"retry_policy": map[string]interface{}{
														"retry_on":    cfg.Retry.RetryOn,
														"num_retries": cfg.Retry.NumRetries,
													},
													"routes": []map[string]interface{}{
														{
															"match": map[string]string{"prefix": "/"},
															"route": route,
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
			"clusters": []map[string]interface{}{
				{
					"name":            cfg.ClusterName,
					"connect_timeout": cfg.Timeout.ConnectTimeout,
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

func printRetriesTimeoutsGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY RETRIES & TIMEOUTS GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[RETRY CONDITIONS (retry_on)]")
	fmt.Println(strings.Repeat("-", 40))
	conditions := [][]string{
		{"5xx", "Any 5xx response from upstream"},
		{"gateway-error", "502, 503, 504 responses"},
		{"reset", "Upstream closed connection"},
		{"connect-failure", "TCP connection to upstream failed"},
		{"retriable-4xx", "Retriable 4xx (e.g. 409 Conflict)"},
		{"refused-stream", "HTTP/2 REFUSED_STREAM reset"},
		{"retriable-status-codes", "Codes listed in x-envoy-retriable-status-codes"},
	}
	for _, c := range conditions {
		fmt.Printf("  %-30s %s\n", c[0], c[1])
	}

	fmt.Println("\n[TIMEOUT TYPES]")
	fmt.Println(strings.Repeat("-", 40))
	timeouts := [][]string{
		{"route timeout", "Total time for the full request-response cycle"},
		{"per_try_timeout", "Max time for each individual retry attempt"},
		{"idle_timeout", "Time a stream can be idle (no data in either direction)"},
		{"connect_timeout", "Time to establish TCP connection to upstream"},
		{"max_stream_duration", "Absolute max lifetime of a streaming request"},
	}
	for _, t := range timeouts {
		fmt.Printf("  %-25s %s\n", t[0], t[1])
	}

	fmt.Println("\n[RECOMMENDED DEFAULTS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  route timeout   : 15s")
	fmt.Println("  per_try_timeout : 5s")
	fmt.Println("  connect_timeout : 0.5s")
	fmt.Println("  num_retries     : 3")
	fmt.Println("  retry_on        : 5xx,gateway-error,connect-failure,reset")

	fmt.Println("\n[MONITORING ADMIN QUERIES]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  curl http://localhost:9901/stats | grep upstream_rq_retry")
	fmt.Println("  curl http://localhost:9901/stats | grep upstream_rq_timeout")
	fmt.Println("  curl http://localhost:9901/stats | grep upstream_cx_connect_timeout")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	retryOn := flag.String("retry-on", "5xx,gateway-error,connect-failure,reset", "Retry conditions")
	numRetries := flag.Int("retries", 3, "Number of retries")
	perTryTimeout := flag.String("per-try-timeout", "5s", "Per-retry attempt timeout")
	routeTimeout := flag.String("timeout", "15s", "Total route timeout")
	idleTimeout := flag.String("idle-timeout", "60s", "Idle stream timeout")
	connectTimeout := flag.String("connect-timeout", "0.5s", "Upstream connect timeout")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print retries & timeouts guide")
	flag.Parse()

	if *guide {
		printRetriesTimeoutsGuide()
		return
	}

	cfg := RetriesTimeoutsConfig{
		ClusterName:  *cluster,
		ListenerPort: *listenerPort,
		AdminPort:    *adminPort,
		Retry: RetryPolicy{
			RetryOn:                  *retryOn,
			NumRetries:               *numRetries,
			PerTryTimeout:            *perTryTimeout,
			HostSelectionMaxAttempts: 3,
		},
		Timeout: TimeoutPolicy{
			RouteTimeout:   *routeTimeout,
			IdleTimeout:    *idleTimeout,
			ConnectTimeout: *connectTimeout,
			PerTryTimeout:  *perTryTimeout,
		},
	}

	config := GenerateRetriesTimeoutsConfig(cfg)

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
