package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// CircuitBreakerConfig holds parameters for Envoy circuit breaker / outlier detection
type CircuitBreakerConfig struct {
	ClusterName              string
	ListenerPort             int
	AdminPort                int
	MaxConnections           int
	MaxPendingRequests       int
	MaxRequests              int
	MaxRetries               int
	MaxConnectionPools       int
	OutlierDetection         bool
	ConsecutiveErrors        int
	Interval                 string // e.g. "10s"
	BaseEjectionTime         string // e.g. "30s"
	MaxEjectionPercent       int
	SuccessRateMinHosts      int
	SuccessRateRequestVolume int
	SuccessRateStdevFactor   int
}

// GenerateCircuitBreakerConfig produces an Envoy static config with circuit breakers
func GenerateCircuitBreakerConfig(cfg CircuitBreakerConfig) map[string]interface{} {
	cluster := map[string]interface{}{
		"name":            cfg.ClusterName,
		"connect_timeout": "1s",
		"type":            "LOGICAL_DNS",
		"lb_policy":       "ROUND_ROBIN",
		"circuit_breakers": map[string]interface{}{
			"thresholds": []map[string]interface{}{
				{
					"priority":             "DEFAULT",
					"max_connections":      cfg.MaxConnections,
					"max_pending_requests": cfg.MaxPendingRequests,
					"max_requests":         cfg.MaxRequests,
					"max_retries":          cfg.MaxRetries,
					"max_connection_pools": cfg.MaxConnectionPools,
					"track_remaining":      true,
				},
			},
		},
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
	}

	if cfg.OutlierDetection {
		cluster["outlier_detection"] = map[string]interface{}{
			"consecutive_5xx":             cfg.ConsecutiveErrors,
			"interval":                    cfg.Interval,
			"base_ejection_time":          cfg.BaseEjectionTime,
			"max_ejection_percent":        cfg.MaxEjectionPercent,
			"success_rate_minimum_hosts":  cfg.SuccessRateMinHosts,
			"success_rate_request_volume": cfg.SuccessRateRequestVolume,
			"success_rate_stdev_factor":   cfg.SuccessRateStdevFactor,
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
										"@type":       "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
										"stat_prefix": "ingress_http",
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
			"clusters": []map[string]interface{}{cluster},
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

func printCircuitBreakerGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY CIRCUIT BREAKER & OUTLIER DETECTION GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[CIRCUIT BREAKER THRESHOLDS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  %-30s %s\n", "max_connections", "Max TCP connections to the cluster")
	fmt.Printf("  %-30s %s\n", "max_pending_requests", "Max queued requests waiting for a connection")
	fmt.Printf("  %-30s %s\n", "max_requests", "Max active requests in parallel")
	fmt.Printf("  %-30s %s\n", "max_retries", "Max concurrent retries")
	fmt.Printf("  %-30s %s\n", "max_connection_pools", "Max connection pool instances (per host)")

	fmt.Println("\n[OUTLIER DETECTION]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  Automatically ejects unhealthy hosts from the load-balancing pool.")
	fmt.Printf("  %-35s %s\n", "consecutive_5xx", "Failures before ejection")
	fmt.Printf("  %-35s %s\n", "interval", "Analysis interval window")
	fmt.Printf("  %-35s %s\n", "base_ejection_time", "Initial ejection duration")
	fmt.Printf("  %-35s %s\n", "max_ejection_percent", "% of hosts that can be ejected at once")
	fmt.Printf("  %-35s %s\n", "success_rate_minimum_hosts", "Min hosts needed for rate analysis")

	fmt.Println("\n[MONITORING ADMIN QUERIES]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # Circuit breaker stats")
	fmt.Println("  curl http://localhost:9901/stats | grep circuit_breakers")
	fmt.Println("")
	fmt.Println("  # Outlier detection events")
	fmt.Println("  curl http://localhost:9901/stats | grep outlier_detection")
	fmt.Println("")
	fmt.Println("  # See remaining capacity")
	fmt.Println("  curl http://localhost:9901/stats | grep cx_open")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	maxConn := flag.Int("max-connections", 1024, "Max connections")
	maxPending := flag.Int("max-pending", 1024, "Max pending requests")
	maxReq := flag.Int("max-requests", 1024, "Max concurrent requests")
	maxRetries := flag.Int("max-retries", 3, "Max concurrent retries")
	maxPools := flag.Int("max-pools", 1024, "Max connection pools")
	outlier := flag.Bool("outlier", true, "Enable outlier detection")
	consecutiveErrors := flag.Int("consecutive-errors", 5, "Consecutive errors before ejection")
	interval := flag.String("interval", "10s", "Outlier detection interval")
	ejectionTime := flag.String("ejection-time", "30s", "Base ejection time")
	maxEjectionPct := flag.Int("max-ejection-pct", 50, "Max ejection percentage")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print circuit breaker guide")
	flag.Parse()

	if *guide {
		printCircuitBreakerGuide()
		return
	}

	cfg := CircuitBreakerConfig{
		ClusterName:              *cluster,
		ListenerPort:             *listenerPort,
		AdminPort:                *adminPort,
		MaxConnections:           *maxConn,
		MaxPendingRequests:       *maxPending,
		MaxRequests:              *maxReq,
		MaxRetries:               *maxRetries,
		MaxConnectionPools:       *maxPools,
		OutlierDetection:         *outlier,
		ConsecutiveErrors:        *consecutiveErrors,
		Interval:                 *interval,
		BaseEjectionTime:         *ejectionTime,
		MaxEjectionPercent:       *maxEjectionPct,
		SuccessRateMinHosts:      5,
		SuccessRateRequestVolume: 100,
		SuccessRateStdevFactor:   1900,
	}

	config := GenerateCircuitBreakerConfig(cfg)

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
