package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// LoadBalancingStrategy defines the type of load balancing
type LoadBalancingStrategy string

const (
	RoundRobin      LoadBalancingStrategy = "ROUND_ROBIN"
	LeastRequest    LoadBalancingStrategy = "LEAST_REQUEST"
	RingHash        LoadBalancingStrategy = "RING_HASH"
	Random          LoadBalancingStrategy = "RANDOM"
	Maglev          LoadBalancingStrategy = "MAGLEV"
	ClusterProvided LoadBalancingStrategy = "CLUSTER_PROVIDED"
)

// EndpointConfig represents a backend endpoint
type EndpointConfig struct {
	Address string
	Port    int
	Weight  int
}

// LoadBalancerConfig contains the configuration for load balancing
type LoadBalancerConfig struct {
	Strategy  LoadBalancingStrategy
	Name      string
	Endpoints []EndpointConfig
	Port      int
	AdminPort int
}

// EnvoyConfig represents the full Envoy configuration structure
type EnvoyConfig struct {
	StaticResources StaticResources `yaml:"static_resources"`
	Admin           AdminConfig     `yaml:"admin"`
}

type StaticResources struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type Listener struct {
	Name         string        `yaml:"name"`
	Address      SocketAddress `yaml:"address"`
	FilterChains []FilterChain `yaml:"filter_chains"`
}

type SocketAddress struct {
	SocketAddress InnerSocketAddress `yaml:"socket_address"`
}

type InnerSocketAddress struct {
	Address   string `yaml:"address"`
	PortValue int    `yaml:"port_value"`
}

type FilterChain struct {
	Filters []Filter `yaml:"filters"`
}

type Filter struct {
	Name        string      `yaml:"name"`
	TypedConfig interface{} `yaml:"typed_config"`
}

type HttpConnectionManager struct {
	Type        string       `yaml:"@type"`
	StatPrefix  string       `yaml:"stat_prefix"`
	HttpFilters []HttpFilter `yaml:"http_filters"`
	RouteConfig RouteConfig  `yaml:"route_config"`
	AccessLog   []AccessLog  `yaml:"access_log,omitempty"`
}

type HttpFilter struct {
	Name string `yaml:"name"`
}

type RouteConfig struct {
	Name         string        `yaml:"name"`
	VirtualHosts []VirtualHost `yaml:"virtual_hosts"`
}

type VirtualHost struct {
	Name    string   `yaml:"name"`
	Domains []string `yaml:"domains"`
	Routes  []Route  `yaml:"routes"`
}

type Route struct {
	Match Match       `yaml:"match"`
	Route RouteAction `yaml:"route"`
}

type Match struct {
	Prefix string `yaml:"prefix"`
}

type RouteAction struct {
	Cluster     string       `yaml:"cluster"`
	Timeout     string       `yaml:"timeout,omitempty"`
	RetryPolicy *RetryPolicy `yaml:"retry_policy,omitempty"`
}

type RetryPolicy struct {
	RetryOn       string `yaml:"retry_on"`
	NumRetries    int    `yaml:"num_retries"`
	PerTryTimeout string `yaml:"per_try_timeout"`
}

type AccessLog struct {
	Name        string      `yaml:"name"`
	TypedConfig interface{} `yaml:"typed_config"`
}

type Cluster struct {
	Name             string            `yaml:"name"`
	Type             string            `yaml:"type"`
	ConnectTimeout   string            `yaml:"connect_timeout,omitempty"`
	LoadAssignment   LoadAssignment    `yaml:"load_assignment"`
	LbPolicy         string            `yaml:"lb_policy,omitempty"`
	LbConfig         interface{}       `yaml:"lb_config,omitempty"`
	OutlierDetection *OutlierDetection `yaml:"outlier_detection,omitempty"`
	HealthChecks     []HealthCheck     `yaml:"health_checks,omitempty"`
}

type LoadAssignment struct {
	ClusterName string                `yaml:"cluster_name"`
	Endpoints   []LocalityLbEndpoints `yaml:"endpoints"`
}

type LocalityLbEndpoints struct {
	LbEndpoints []LbEndpoint `yaml:"lb_endpoints"`
}

type LbEndpoint struct {
	Endpoint            Endpoint `yaml:"endpoint"`
	LoadBalancingWeight int      `yaml:"load_balancing_weight,omitempty"`
}

type Endpoint struct {
	Address SocketAddress `yaml:"address"`
}

type OutlierDetection struct {
	Consecutive5xx                 int    `yaml:"consecutive_5xx,omitempty"`
	Interval                       string `yaml:"interval,omitempty"`
	BaseEjectionTime               string `yaml:"base_ejection_time,omitempty"`
	SplitExternalLocalOriginErrors bool   `yaml:"split_external_local_origin_errors,omitempty"`
}

type HealthCheck struct {
	Timeout            string          `yaml:"timeout"`
	Interval           string          `yaml:"interval"`
	UnhealthyThreshold int             `yaml:"unhealthy_threshold"`
	HealthyThreshold   int             `yaml:"healthy_threshold"`
	HttpHealthCheck    HttpHealthCheck `yaml:"http_health_check"`
}

type HttpHealthCheck struct {
	Path string `yaml:"path"`
}

type AdminConfig struct {
	AccessLogPath string        `yaml:"access_log_path"`
	Address       SocketAddress `yaml:"address"`
}

// RingHashConfig for ring hash load balancing
type RingHashConfig struct {
	Type            string `yaml:"@type"`
	HashFunction    string `yaml:"hash_function,omitempty"`
	MinimumRingSize int    `yaml:"minimum_ring_size,omitempty"`
	MaximumRingSize int    `yaml:"maximum_ring_size,omitempty"`
	// For older Envoy versions without lb_config
}

// LeastRequestConfig for least request load balancing
type LeastRequestConfig struct {
	Type        string `yaml:"@type"`
	ChoiceCount int    `yaml:"choice_count,omitempty"`
}

// MagleVConfig for Maglev load balancing
type MaglevConfig struct {
	Type      string `yaml:"@type"`
	TableSize int    `yaml:"table_size,omitempty"`
}

// GenerateEnvoyConfig creates an Envoy configuration for the given strategy
func GenerateEnvoyConfig(config LoadBalancerConfig) string {
	// Create clusters with appropriate load balancing strategy
	var clusters []Cluster

	cluster := createCluster(config.Name, config.Strategy, config.Endpoints)
	clusters = append(clusters, cluster)

	// Create listener
	listener := Listener{
		Name: "listener_0",
		Address: SocketAddress{
			SocketAddress: InnerSocketAddress{
				Address:   "0.0.0.0",
				PortValue: config.Port,
			},
		},
		FilterChains: []FilterChain{
			{
				Filters: []Filter{
					{
						Name: "envoy.filters.network.http_connection_manager",
						TypedConfig: map[string]interface{}{
							"@type":       "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
							"stat_prefix": "ingress_http",
							"http_filters": []map[string]interface{}{
								{
									"name": "envoy.filters.http.router",
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
												"match": map[string]string{
													"prefix": "/",
												},
												"route": map[string]interface{}{
													"cluster": config.Name,
													"timeout": "30s",
													"retry_policy": map[string]interface{}{
														"retry_on":        "5xx",
														"num_retries":     3,
														"per_try_timeout": "10s",
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
		},
	}

	// Build the full config
	envoyConfig := map[string]interface{}{
		"static_resources": map[string]interface{}{
			"listeners": []Listener{listener},
			"clusters":  clusters,
		},
		"admin": map[string]interface{}{
			"access_log_path": "/tmp/admin_access.log",
			"address": map[string]interface{}{
				"socket_address": map[string]interface{}{
					"address":    "0.0.0.0",
					"port_value": config.AdminPort,
				},
			},
		},
	}

	// Convert to JSON then to YAML-like format
	jsonData, _ := json.MarshalIndent(envoyConfig, "", "  ")
	return convertJSONToYAML(string(jsonData))
}

func createCluster(name string, strategy LoadBalancingStrategy, endpoints []EndpointConfig) Cluster {
	// Create load assignment endpoints
	var lbEndpoints []LbEndpoint
	for _, ep := range endpoints {
		lbEndpoint := LbEndpoint{
			Endpoint: Endpoint{
				Address: SocketAddress{
					SocketAddress: InnerSocketAddress{
						Address:   ep.Address,
						PortValue: ep.Port,
					},
				},
			},
			LoadBalancingWeight: ep.Weight,
		}
		lbEndpoints = append(lbEndpoints, lbEndpoint)
	}

	cluster := Cluster{
		Name:           name,
		Type:           "STATIC",
		ConnectTimeout: "1s",
		LoadAssignment: LoadAssignment{
			ClusterName: name,
			Endpoints: []LocalityLbEndpoints{
				{
					LbEndpoints: lbEndpoints,
				},
			},
		},
		LbPolicy: string(strategy),
		OutlierDetection: &OutlierDetection{
			Consecutive5xx:                 5,
			Interval:                       "30s",
			BaseEjectionTime:               "30s",
			SplitExternalLocalOriginErrors: true,
		},
		HealthChecks: []HealthCheck{
			{
				Timeout:            "1s",
				Interval:           "10s",
				UnhealthyThreshold: 2,
				HealthyThreshold:   2,
				HttpHealthCheck: HttpHealthCheck{
					Path: "/health",
				},
			},
		},
	}

	return cluster
}

func convertJSONToYAML(jsonStr string) string {
	// Simple JSON to YAML conversion
	// In production, use a proper YAML library
	yamlStr := jsonStr
	yamlStr = replace(yamlStr, `"`, "")
	yamlStr = replace(yamlStr, "{", "")
	yamlStr = replace(yamlStr, "}", "")
	yamlStr = replace(yamlStr, "[", "")
	yamlStr = replace(yamlStr, "]", "")
	yamlStr = replace(yamlStr, ",", "")
	return yamlStr
}

func replace(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// LoadBalancingStrategyGuide provides detailed information about each strategy
type LoadBalancingStrategyGuide struct {
	Strategy      LoadBalancingStrategy
	Description   string
	UseCase       string
	Pros          []string
	Cons          []string
	Configuration string
}

var StrategyGuides = []LoadBalancingStrategyGuide{
	{
		Strategy:    RoundRobin,
		Description: "Cycles through endpoints in order",
		UseCase:     "Simple load distribution when all endpoints have equal capacity",
		Pros: []string{
			"Simple and predictable",
			"Good for homogeneous backends",
			"No computation overhead",
			"Fair distribution",
		},
		Cons: []string{
			"Doesn't account for endpoint health",
			"Ignores backend load",
			"Not suitable for heterogeneous endpoints",
		},
		Configuration: `lb_policy: ROUND_ROBIN
# No additional configuration needed`,
	},
	{
		Strategy:    LeastRequest,
		Description: "Routes to endpoint with least active requests",
		UseCase:     "When endpoint capacity varies or handling long-lived connections",
		Pros: []string{
			"Adapts to endpoint load",
			"Good for variable capacity backends",
			"Prevents overloading slow endpoints",
			"Configurable choice count",
		},
		Cons: []string{
			"Requires tracking active connections",
			"Slightly higher latency for routing decision",
			"Less predictable distribution",
		},
		Configuration: `lb_policy: LEAST_REQUEST
lb_config:
  "@type": type.googleapis.com/envoy.extensions.load_balancers.least_request.v3.LeastRequest
  choice_count: 2  # Number of random backends to compare`,
	},
	{
		Strategy:    RingHash,
		Description: "Uses consistent hashing based on header values",
		UseCase:     "Session persistence, consistent routing for stateful backends",
		Pros: []string{
			"Session affinity/stickiness",
			"Consistent routing for same client",
			"Minimal redistribution on endpoint changes",
			"Good for stateful services",
		},
		Cons: []string{
			"Uneven distribution if ring size not configured",
			"Hash collisions possible",
			"Requires hash key configuration",
		},
		Configuration: `lb_policy: RING_HASH
lb_config:
  "@type": type.googleapis.com/envoy.extensions.load_balancers.ring_hash.v3.RingHash
  hash_function: XX_HASH  # or MURMUR_HASH_2
  minimum_ring_size: 1024
  maximum_ring_size: 8192`,
	},
	{
		Strategy:    Random,
		Description: "Randomly selects an endpoint",
		UseCase:     "Simple load distribution without order dependency",
		Pros: []string{
			"Very simple implementation",
			"Good statistical distribution",
			"No state required",
			"Fast routing decision",
		},
		Cons: []string{
			"Less predictable than round-robin",
			"Doesn't consider endpoint health",
			"Poor for small number of endpoints",
		},
		Configuration: `lb_policy: RANDOM
# No additional configuration needed`,
	},
	{
		Strategy:    Maglev,
		Description: "Consistent hashing using Maglev algorithm",
		UseCase:     "Session persistence with better distribution than ring hash",
		Pros: []string{
			"Superior consistency vs ring hash",
			"Better load distribution",
			"Lower collision rate",
			"Scales better",
		},
		Cons: []string{
			"More complex algorithm",
			"Higher CPU overhead",
			"Requires Envoy with Maglev support",
		},
		Configuration: `lb_policy: MAGLEV
lb_config:
  "@type": type.googleapis.com/envoy.extensions.load_balancers.maglev.v3.Maglev
  table_size: 65537  # Prime number recommended`,
	},
	{
		Strategy:    ClusterProvided,
		Description: "Delegates load balancing to EDS (Endpoint Discovery Service)",
		UseCase:     "Dynamic endpoint management with external load balancing logic",
		Pros: []string{
			"Flexible, custom load balancing",
			"Dynamic updates via xDS",
			"Integration with service discovery",
		},
		Cons: []string{
			"Requires control plane",
			"More complex setup",
		},
		Configuration: `lb_policy: CLUSTER_PROVIDED
# Requires external control plane like Istio`,
	},
}

// PrintAllStrategies prints detailed guide for all strategies
func PrintAllStrategies() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ENVOY LOAD BALANCING STRATEGIES - COMPREHENSIVE GUIDE")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	for i, guide := range StrategyGuides {
		fmt.Printf("%d. %s\n", i+1, guide.Strategy)
		fmt.Printf("   %s\n", guide.Description)
		fmt.Printf("   Use Case: %s\n\n", guide.UseCase)

		fmt.Println("   Pros:")
		for _, pro := range guide.Pros {
			fmt.Printf("   ✓ %s\n", pro)
		}

		fmt.Println("\n   Cons:")
		for _, con := range guide.Cons {
			fmt.Printf("   ✗ %s\n", con)
		}

		fmt.Println("\n   Configuration:")
		fmt.Println("   ```yaml")
		fmt.Println("   " + guide.Configuration)
		fmt.Println("   ```\n")

		if i < len(StrategyGuides)-1 {
			fmt.Println(strings.Repeat("-", 80))
		}
	}
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func main() {
	strategyFlag := flag.String("strategy", "ROUND_ROBIN", "Load balancing strategy (ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, MAGLEV, CLUSTER_PROVIDED)")
	nameFlag := flag.String("name", "backend_cluster", "Cluster name")
	portFlag := flag.Int("port", 10000, "Listener port")
	adminPortFlag := flag.Int("admin-port", 9901, "Admin interface port")
	outputFlag := flag.String("output", "", "Output file for generated configuration")
	guideFlag := flag.Bool("guide", false, "Show detailed guide for all strategies")

	flag.Parse()

	if *guideFlag {
		PrintAllStrategies()
		return
	}

	// Validate strategy
	strategy := LoadBalancingStrategy(*strategyFlag)
	validStrategy := false
	for _, validStrategyName := range []LoadBalancingStrategy{
		RoundRobin, LeastRequest, RingHash, Random, Maglev, ClusterProvided,
	} {
		if strategy == validStrategyName {
			validStrategy = true
			break
		}
	}

	if !validStrategy {
		fmt.Printf("Error: Invalid strategy '%s'\n", strategy)
		fmt.Println("Valid strategies: ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, MAGLEV, CLUSTER_PROVIDED")
		os.Exit(1)
	}

	// Create example endpoints
	endpoints := []EndpointConfig{
		{Address: "127.0.0.1", Port: 8080, Weight: 1},
		{Address: "127.0.0.1", Port: 8081, Weight: 1},
		{Address: "127.0.0.1", Port: 8082, Weight: 1},
	}

	// Create config
	config := LoadBalancerConfig{
		Strategy:  strategy,
		Name:      *nameFlag,
		Endpoints: endpoints,
		Port:      *portFlag,
		AdminPort: *adminPortFlag,
	}

	// Generate configuration
	envoyConfig := GenerateEnvoyConfig(config)

	if *outputFlag != "" {
		// Write to file
		err := os.WriteFile(*outputFlag, []byte(envoyConfig), 0644)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration written to: %s\n", *outputFlag)
	} else {
		// Print to stdout
		fmt.Println(envoyConfig)
	}

	// Print strategy information
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("LOAD BALANCING STRATEGY: %s\n", strategy)
	fmt.Println(strings.Repeat("=", 80))

	for _, guide := range StrategyGuides {
		if guide.Strategy == strategy {
			fmt.Printf("\nDescription: %s\n", guide.Description)
			fmt.Printf("Use Case: %s\n\n", guide.UseCase)
			fmt.Println("Pros:")
			for _, pro := range guide.Pros {
				fmt.Printf("  ✓ %s\n", pro)
			}
			fmt.Println("\nCons:")
			for _, con := range guide.Cons {
				fmt.Printf("  ✗ %s\n", con)
			}
			break
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Configuration Details:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("  Listener Port:  %d\n", *portFlag)
	fmt.Printf("  Admin Port:     %d\n", *adminPortFlag)
	fmt.Printf("  Cluster Name:   %s\n", *nameFlag)
	fmt.Printf("  Backends:       %d endpoints\n", len(endpoints))
	fmt.Printf("  Strategy:       %s\n", strategy)
	fmt.Println()

	fmt.Println("For more detailed information about all strategies, run:")
	fmt.Println("  go run setup.go --guide")
}
