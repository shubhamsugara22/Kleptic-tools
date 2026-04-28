package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// RateLimitDescriptor defines a key-value pair for rate limit matching
type RateLimitDescriptor struct {
	Key   string
	Value string
}

// TokenBucketConfig defines local rate limit parameters
type TokenBucketConfig struct {
	MaxTokens     int
	TokensPerFill int
	FillInterval  string // e.g. "1s", "60s"
}

// RateLimitConfig holds all parameters for generating an Envoy rate limit config
type RateLimitConfig struct {
	ClusterName      string
	ListenerPort     int
	AdminPort        int
	TokenBucket      TokenBucketConfig
	GlobalRateLimit  bool
	RateLimitService string // address of external rate limit service (global mode)
	Descriptors      []RateLimitDescriptor
}

// GenerateLocalRateLimitConfig generates an Envoy static config with local rate limiting
func GenerateLocalRateLimitConfig(cfg RateLimitConfig) map[string]interface{} {
	httpFilters := []map[string]interface{}{
		{
			"name": "envoy.filters.http.local_ratelimit",
			"typed_config": map[string]interface{}{
				"@type":        "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit",
				"stat_prefix":  "http_local_rate_limiter",
				"token_bucket": map[string]interface{}{
					"max_tokens":      cfg.TokenBucket.MaxTokens,
					"tokens_per_fill": cfg.TokenBucket.TokensPerFill,
					"fill_interval":   cfg.TokenBucket.FillInterval,
				},
				"filter_enabled": map[string]interface{}{
					"runtime_key":  "local_rate_limit_enabled",
					"default_value": map[string]interface{}{
						"numerator":   100,
						"denominator": "HUNDRED",
					},
				},
				"filter_enforced": map[string]interface{}{
					"runtime_key":  "local_rate_limit_enforced",
					"default_value": map[string]interface{}{
						"numerator":   100,
						"denominator": "HUNDRED",
					},
				},
				"response_headers_to_add": []map[string]interface{}{
					{
						"append": false,
						"header": map[string]string{
							"key":   "x-local-rate-limit",
							"value": "true",
						},
					},
				},
			},
		},
		{
			"name": "envoy.filters.http.router",
			"typed_config": map[string]interface{}{
				"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
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
															"match":  map[string]string{"prefix": "/"},
															"route":  map[string]string{"cluster": cfg.ClusterName},
														},
													},
												},
											},
										},
										"http_filters": httpFilters,
									},
								},
							},
						},
					},
				},
			},
			"clusters": []map[string]interface{}{
				{
					"name":                   cfg.ClusterName,
					"connect_timeout":        "0.25s",
					"type":                   "LOGICAL_DNS",
					"lb_policy":              "ROUND_ROBIN",
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

// GenerateGlobalRateLimitConfig generates an Envoy config with global rate limiting via external service
func GenerateGlobalRateLimitConfig(cfg RateLimitConfig) map[string]interface{}{
	descriptors := []map[string]interface{}{}
	for _, d := range cfg.Descriptors {
		descriptors = append(descriptors, map[string]interface{}{
			"entries": []map[string]string{
				{"key": d.Key, "value": d.Value},
			},
		})
	}

	httpFilters := []map[string]interface{}{
		{
			"name": "envoy.filters.http.ratelimit",
			"typed_config": map[string]interface{}{
				"@type":            "type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit",
				"domain":           "envoy_ratelimit",
				"failure_mode_deny": false,
				"rate_limit_service": map[string]interface{}{
					"grpc_service": map[string]interface{}{
						"envoy_grpc": map[string]interface{}{
							"cluster_name": "rate_limit_cluster",
						},
					},
					"transport_api_version": "V3",
				},
			},
		},
		{
			"name": "envoy.filters.http.router",
			"typed_config": map[string]interface{}{
				"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
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
															"route": map[string]interface{}{
																"cluster": cfg.ClusterName,
																"rate_limits": []map[string]interface{}{
																	{"actions": descriptors},
																},
															},
														},
													},
												},
											},
										},
										"http_filters": httpFilters,
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
					"connect_timeout": "0.25s",
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
				{
					"name":             "rate_limit_cluster",
					"connect_timeout":  "0.25s",
					"type":             "STRICT_DNS",
					"lb_policy":        "ROUND_ROBIN",
					"http2_protocol_options": map[string]interface{}{},
					"load_assignment": map[string]interface{}{
						"cluster_name": "rate_limit_cluster",
						"endpoints": []map[string]interface{}{
							{
								"lb_endpoints": []map[string]interface{}{
									{
										"endpoint": map[string]interface{}{
											"address": map[string]interface{}{
												"socket_address": map[string]interface{}{
													"address":    cfg.RateLimitService,
													"port_value": 8081,
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

func printRateLimitGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY RATE LIMITING GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[LOCAL RATE LIMITING]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  Enforced per Envoy instance (no shared state).")
	fmt.Println("  Uses a token bucket algorithm.")
	fmt.Println("  Best for: single-instance deployments, edge rate limiting.")
	fmt.Println("  Pros : zero external deps, low latency overhead")
	fmt.Println("  Cons : not globally consistent across replicas")

	fmt.Println("\n[GLOBAL RATE LIMITING]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  Calls an external gRPC rate-limit service (e.g. Lyft ratelimit).")
	fmt.Println("  Shared state — consistent across all Envoy replicas.")
	fmt.Println("  Best for: multi-replica deployments, per-user quotas.")
	fmt.Println("  Pros : accurate global enforcement")
	fmt.Println("  Cons : extra hop, single point of failure if not HA")

	fmt.Println("\n[TOKEN BUCKET PARAMETERS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  %-20s %s\n", "max_tokens", "Maximum burst capacity")
	fmt.Printf("  %-20s %s\n", "tokens_per_fill", "Tokens added per interval")
	fmt.Printf("  %-20s %s\n", "fill_interval", "How often tokens are added (e.g. 1s)")

	fmt.Println("\n[USEFUL ADMIN QUERIES]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # Check local rate limit stats")
	fmt.Println("  curl http://localhost:9901/stats | grep rate_limit")
	fmt.Println("")
	fmt.Println("  # Check listener stats")
	fmt.Println("  curl http://localhost:9901/stats | grep listener")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	mode := flag.String("mode", "local", "Rate limit mode: local | global")
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	maxTokens := flag.Int("max-tokens", 100, "Max tokens in bucket (local mode)")
	tokensPerFill := flag.Int("tokens-per-fill", 10, "Tokens added per fill interval (local mode)")
	fillInterval := flag.String("fill-interval", "1s", "Token fill interval (local mode)")
	rlService := flag.String("rl-service", "ratelimit.default.svc.cluster.local", "Rate limit service host (global mode)")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print rate limit strategy guide")
	flag.Parse()

	if *guide {
		printRateLimitGuide()
		return
	}

	cfg := RateLimitConfig{
		ClusterName:  *cluster,
		ListenerPort: *listenerPort,
		AdminPort:    *adminPort,
		TokenBucket: TokenBucketConfig{
			MaxTokens:     *maxTokens,
			TokensPerFill: *tokensPerFill,
			FillInterval:  *fillInterval,
		},
		RateLimitService: *rlService,
		Descriptors: []RateLimitDescriptor{
			{Key: "remote_address", Value: ""},
			{Key: "header_match", Value: "x-api-key"},
		},
	}

	var config map[string]interface{}
	switch *mode {
	case "global":
		cfg.GlobalRateLimit = true
		config = GenerateGlobalRateLimitConfig(cfg)
	default:
		config = GenerateLocalRateLimitConfig(cfg)
	}

	data, err := toJSON(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling config: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, []byte(data), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config written to %s\n", *output)
	} else {
		fmt.Println(data)
	}
}

// toJSON serialises v as indented JSON (Envoy accepts JSON as a valid config format)
func toJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
