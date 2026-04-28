package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// JWKSMode selects how the JWKS (key set) is provided
type JWKSMode string

const (
	JWKSRemote JWKSMode = "remote" // fetch from JWKS URI
	JWKSInline JWKSMode = "inline" // embed the JWKS JSON directly
)

// JWTProvider defines one JWT issuer / JWKS combination
type JWTProvider struct {
	Name          string
	Issuer        string
	Audiences     []string
	JWKSMode      JWKSMode
	JWKSURI       string   // used when mode=remote
	JWKSInline    string   // used when mode=inline (JSON string of the JWK set)
	JWKSCluster   string   // cluster name that resolves the JWKS endpoint
	ForwardJWT    bool     // forward decoded JWT payload as x-jwt-payload header
	HeaderName    string   // JWT location header (default: Authorization)
	ClaimToHeader []string // map claim→header, format "claim:header-name"
}

// JWTAuthConfig is the top-level config for JWT auth generation
type JWTAuthConfig struct {
	ClusterName  string
	ListenerPort int
	AdminPort    int
	Providers    []JWTProvider
}

// buildJWTProvider converts a JWTProvider to the envoy jwt_authn provider block
func buildJWTProvider(p JWTProvider) map[string]interface{} {
	provider := map[string]interface{}{
		"issuer":    p.Issuer,
		"audiences": p.Audiences,
		"forward":   p.ForwardJWT,
	}

	// JWT token location
	if p.HeaderName != "" && p.HeaderName != "Authorization" {
		provider["from_headers"] = []map[string]interface{}{
			{"name": p.HeaderName},
		}
	} else {
		// default: Authorization: Bearer <token>
		provider["from_headers"] = []map[string]interface{}{
			{
				"name":         "Authorization",
				"value_prefix": "Bearer ",
			},
		}
	}

	// JWKS source
	if p.JWKSMode == JWKSInline {
		provider["local_jwks"] = map[string]interface{}{
			"inline_string": p.JWKSInline,
		}
	} else {
		provider["remote_jwks"] = map[string]interface{}{
			"http_uri": map[string]interface{}{
				"uri":     p.JWKSURI,
				"cluster": p.JWKSCluster,
				"timeout": "5s",
			},
			"cache_duration": "300s",
		}
	}

	// Claim-to-header forwarding
	if len(p.ClaimToHeader) > 0 {
		claimHeaders := []map[string]interface{}{}
		for _, ch := range p.ClaimToHeader {
			parts := strings.SplitN(ch, ":", 2)
			if len(parts) == 2 {
				claimHeaders = append(claimHeaders, map[string]interface{}{
					"header_name": parts[1],
					"claim_name":  parts[0],
				})
			}
		}
		if len(claimHeaders) > 0 {
			provider["claim_to_headers"] = claimHeaders
		}
	}

	return provider
}

// GenerateJWTAuthConfig produces an Envoy static config with JWT authentication
func GenerateJWTAuthConfig(cfg JWTAuthConfig) map[string]interface{} {
	// Build the providers map
	providers := map[string]interface{}{}
	rules := []map[string]interface{}{}
	jwksClusters := []map[string]interface{}{}
	seenClusters := map[string]bool{}

	for _, p := range cfg.Providers {
		providers[p.Name] = buildJWTProvider(p)

		// One rule per provider: match all paths
		rules = append(rules, map[string]interface{}{
			"match": map[string]string{"prefix": "/"},
			"requires": map[string]interface{}{
				"provider_name": p.Name,
			},
		})

		// Add a JWKS cluster if remote mode and not already added
		if p.JWKSMode == JWKSRemote && !seenClusters[p.JWKSCluster] {
			seenClusters[p.JWKSCluster] = true
			jwksClusters = append(jwksClusters, map[string]interface{}{
				"name":            p.JWKSCluster,
				"connect_timeout": "5s",
				"type":            "LOGICAL_DNS",
				"lb_policy":       "ROUND_ROBIN",
				"load_assignment": map[string]interface{}{
					"cluster_name": p.JWKSCluster,
					"endpoints": []map[string]interface{}{
						{
							"lb_endpoints": []map[string]interface{}{
								{
									"endpoint": map[string]interface{}{
										"address": map[string]interface{}{
											"socket_address": map[string]interface{}{
												"address":    p.JWKSCluster,
												"port_value": 443,
											},
										},
									},
								},
							},
						},
					},
				},
				"transport_socket": map[string]interface{}{
					"name": "envoy.transport_sockets.tls",
					"typed_config": map[string]interface{}{
						"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext",
					},
				},
			})
		}
	}

	jwtFilter := map[string]interface{}{
		"name": "envoy.filters.http.jwt_authn",
		"typed_config": map[string]interface{}{
			"@type":     "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication",
			"providers": providers,
			"rules":     rules,
		},
	}

	allClusters := []map[string]interface{}{
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
	allClusters = append(allClusters, jwksClusters...)

	return map[string]interface{}{
		"static_resources": map[string]interface{}{
			"listeners": []map[string]interface{}{
				{
					"name": "jwt_listener",
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
											jwtFilter,
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
			"clusters": allClusters,
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

func printJWTGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY JWT AUTHENTICATION GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[HOW IT WORKS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  Envoy's jwt_authn filter validates JWT tokens BEFORE forwarding to upstream.")
	fmt.Println("  - Validates signature against JWKS (remote or inline)")
	fmt.Println("  - Checks iss (issuer) and aud (audience) claims")
	fmt.Println("  - Optionally forwards payload as x-jwt-payload or as custom headers")
	fmt.Println("  - Returns 401 for missing/invalid tokens, 403 for wrong issuer/audience")

	fmt.Println("\n[JWKS MODES]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  remote  Fetch JWKS from a URL (e.g. Auth0 /.well-known/jwks.json)")
	fmt.Println("          Cached for 300s by default. Requires a JWKS cluster definition.")
	fmt.Println("")
	fmt.Println("  inline  Embed the full JWKS JSON in the config.")
	fmt.Println("          Best for static/self-hosted keys. No network round-trip.")

	fmt.Println("\n[EXAMPLE INLINE JWKS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(`  {
    "keys": [{
      "kty": "RSA",
      "use": "sig",
      "kid": "1",
      "n":   "<base64url-encoded modulus>",
      "e":   "AQAB"
    }]
  }`)

	fmt.Println("\n[GENERATE A TEST JWT (requires jwt-cli)]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  jwt encode --secret mysecret --iss https://my-issuer.example.com --aud api")

	fmt.Println("\n[TEST WITH CURL]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # Valid token")
	fmt.Println("  curl -H 'Authorization: Bearer <token>' http://localhost:10000/")
	fmt.Println("")
	fmt.Println("  # Missing token → 401")
	fmt.Println("  curl http://localhost:10000/")

	fmt.Println("\n[CLAIM TO HEADER FORWARDING]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  Use --claim-header 'sub:x-user-id' to forward the 'sub' claim as x-user-id.")
	fmt.Println("  Multiple mappings: --claim-header 'sub:x-user-id' --claim-header 'email:x-email'")
	fmt.Println("  Upstream receives these as regular HTTP headers after JWT validation.")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")

	// Single-provider convenience flags
	providerName := flag.String("provider", "default_provider", "JWT provider name")
	issuer := flag.String("issuer", "https://my-issuer.example.com", "JWT issuer (iss claim)")
	audiences := flag.String("audiences", "api", "Comma-separated audience values")
	jwksMode := flag.String("jwks-mode", "remote", "JWKS mode: remote | inline")
	jwksURI := flag.String("jwks-uri", "https://my-issuer.example.com/.well-known/jwks.json", "Remote JWKS URI")
	jwksCluster := flag.String("jwks-cluster", "jwks_cluster", "Cluster name for JWKS endpoint")
	jwksInline := flag.String("jwks-inline", "", "Inline JWKS JSON string (mode=inline)")
	forwardJWT := flag.Bool("forward-jwt", true, "Forward JWT payload as x-jwt-payload header")
	headerName := flag.String("header", "Authorization", "Header to read JWT from")
	claimHeaders := flag.String("claim-headers", "", "Comma-separated claim:header mappings, e.g. sub:x-user-id,email:x-email")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print JWT authentication guide")
	flag.Parse()

	if *guide {
		printJWTGuide()
		return
	}

	var claimToHeader []string
	if *claimHeaders != "" {
		claimToHeader = strings.Split(*claimHeaders, ",")
	}

	provider := JWTProvider{
		Name:          *providerName,
		Issuer:        *issuer,
		Audiences:     strings.Split(*audiences, ","),
		JWKSMode:      JWKSMode(*jwksMode),
		JWKSURI:       *jwksURI,
		JWKSCluster:   *jwksCluster,
		JWKSInline:    *jwksInline,
		ForwardJWT:    *forwardJWT,
		HeaderName:    *headerName,
		ClaimToHeader: claimToHeader,
	}

	cfg := JWTAuthConfig{
		ClusterName:  *cluster,
		ListenerPort: *listenerPort,
		AdminPort:    *adminPort,
		Providers:    []JWTProvider{provider},
	}

	config := GenerateJWTAuthConfig(cfg)

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
