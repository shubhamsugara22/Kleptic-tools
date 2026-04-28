package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// TLSMode selects the TLS configuration pattern
type TLSMode string

const (
	TLSTerminate   TLSMode = "terminate"   // TLS termination at Envoy (HTTPS in, HTTP out)
	TLSPassthrough TLSMode = "passthrough" // SNI-based passthrough (no termination)
	MTLS           TLSMode = "mtls"        // Mutual TLS — both client and upstream certs required
	TLSOriginate   TLSMode = "originate"   // Envoy originates TLS to upstream (HTTP in, HTTPS out)
)

// CertConfig holds paths for TLS certificates
type CertConfig struct {
	CertChainPath  string // path to server cert (PEM)
	PrivateKeyPath string // path to server private key (PEM)
	CACertPath     string // path to CA cert for client verification (mTLS)
	SNINames       []string
}

// TLSConfig is the top-level config for TLS generation
type TLSConfig struct {
	ClusterName   string
	ListenerPort  int
	AdminPort     int
	Mode          TLSMode
	Certs         CertConfig
	UpstreamHost  string
	UpstreamPort  int
	MinTLSVersion string // e.g. "TLSv1_2", "TLSv1_3"
	AlpnProtocols []string
}

// buildDownstreamTLSContext returns the TLS context for the listener (incoming connections)
func buildDownstreamTLSContext(cfg TLSConfig) map[string]interface{} {
	tlsParams := map[string]interface{}{
		"tls_minimum_protocol_version": cfg.MinTLSVersion,
		"cipher_suites": []string{
			"ECDHE-ECDSA-AES128-GCM-SHA256",
			"ECDHE-RSA-AES128-GCM-SHA256",
			"ECDHE-ECDSA-AES256-GCM-SHA384",
		},
	}

	ctx := map[string]interface{}{
		"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext",
		"common_tls_context": map[string]interface{}{
			"tls_params": tlsParams,
			"tls_certificates": []map[string]interface{}{
				{
					"certificate_chain": map[string]string{
						"filename": cfg.Certs.CertChainPath,
					},
					"private_key": map[string]string{
						"filename": cfg.Certs.PrivateKeyPath,
					},
				},
			},
			"alpn_protocols": cfg.AlpnProtocols,
		},
	}

	// For mTLS, require and verify client certificate
	if cfg.Mode == MTLS {
		ctx["require_client_certificate"] = true
		commonCtx := ctx["common_tls_context"].(map[string]interface{})
		commonCtx["validation_context"] = map[string]interface{}{
			"trusted_ca": map[string]string{
				"filename": cfg.Certs.CACertPath,
			},
		}
	}

	return ctx
}

// buildUpstreamTLSContext returns the TLS context for the cluster (outgoing connections)
func buildUpstreamTLSContext(cfg TLSConfig) map[string]interface{} {
	ctx := map[string]interface{}{
		"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext",
		"common_tls_context": map[string]interface{}{
			"tls_params": map[string]interface{}{
				"tls_minimum_protocol_version": cfg.MinTLSVersion,
			},
			"alpn_protocols": cfg.AlpnProtocols,
		},
	}

	// For mTLS originate — present our own client cert to upstream
	if cfg.Mode == MTLS {
		commonCtx := ctx["common_tls_context"].(map[string]interface{})
		commonCtx["tls_certificates"] = []map[string]interface{}{
			{
				"certificate_chain": map[string]string{"filename": cfg.Certs.CertChainPath},
				"private_key":       map[string]string{"filename": cfg.Certs.PrivateKeyPath},
			},
		}
		commonCtx["validation_context"] = map[string]interface{}{
			"trusted_ca": map[string]string{"filename": cfg.Certs.CACertPath},
		}
	}

	return ctx
}

// GenerateTLSConfig produces an Envoy static config for the chosen TLS mode
func GenerateTLSConfig(cfg TLSConfig) map[string]interface{} {
	var listenerFilterChain map[string]interface{}
	var clusterTransportSocket map[string]interface{}

	switch cfg.Mode {
	case TLSPassthrough:
		// SNI-based passthrough — use TLS inspector + tcp_proxy
		listenerFilterChain = map[string]interface{}{
			"filter_chain_match": map[string]interface{}{
				"server_names": cfg.Certs.SNINames,
			},
			"filters": []map[string]interface{}{
				{
					"name": "envoy.filters.network.tcp_proxy",
					"typed_config": map[string]interface{}{
						"@type":       "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy",
						"stat_prefix": "passthrough",
						"cluster":     cfg.ClusterName,
					},
				},
			},
		}

	case TLSOriginate:
		// Plain HTTP listener, but originate TLS to upstream
		listenerFilterChain = buildHTTPFilterChain(cfg)
		clusterTransportSocket = map[string]interface{}{
			"name":         "envoy.transport_sockets.tls",
			"typed_config": buildUpstreamTLSContext(cfg),
		}

	case MTLS:
		// mTLS on both sides
		downstreamCtx := buildDownstreamTLSContext(cfg)
		listenerFilterChain = map[string]interface{}{
			"transport_socket": map[string]interface{}{
				"name":         "envoy.transport_sockets.tls",
				"typed_config": downstreamCtx,
			},
			"filters": buildHTTPFilterChain(cfg)["filters"],
		}
		clusterTransportSocket = map[string]interface{}{
			"name":         "envoy.transport_sockets.tls",
			"typed_config": buildUpstreamTLSContext(cfg),
		}

	default: // terminate
		downstreamCtx := buildDownstreamTLSContext(cfg)
		listenerFilterChain = map[string]interface{}{
			"transport_socket": map[string]interface{}{
				"name":         "envoy.transport_sockets.tls",
				"typed_config": downstreamCtx,
			},
			"filters": buildHTTPFilterChain(cfg)["filters"],
		}
	}

	listenerConfig := map[string]interface{}{
		"name": "tls_listener",
		"address": map[string]interface{}{
			"socket_address": map[string]interface{}{
				"address":    "0.0.0.0",
				"port_value": cfg.ListenerPort,
			},
		},
		"filter_chains": []map[string]interface{}{listenerFilterChain},
	}

	// TLS passthrough requires the TLS inspector to read SNI before routing
	if cfg.Mode == TLSPassthrough {
		listenerConfig["listener_filters"] = []map[string]interface{}{
			{"name": "envoy.filters.listener.tls_inspector"},
		}
	}

	cluster := map[string]interface{}{
		"name":            cfg.ClusterName,
		"connect_timeout": "1s",
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
										"address":    cfg.UpstreamHost,
										"port_value": cfg.UpstreamPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if clusterTransportSocket != nil {
		cluster["transport_socket"] = clusterTransportSocket
	}

	return map[string]interface{}{
		"static_resources": map[string]interface{}{
			"listeners": []map[string]interface{}{listenerConfig},
			"clusters":  []map[string]interface{}{cluster},
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

// buildHTTPFilterChain returns an HCM filter chain (reused across modes)
func buildHTTPFilterChain(cfg TLSConfig) map[string]interface{} {
	return map[string]interface{}{
		"filters": []map[string]interface{}{
			{
				"name": "envoy.filters.network.http_connection_manager",
				"typed_config": map[string]interface{}{
					"@type":       "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
					"stat_prefix": "ingress_https",
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
	}
}

func printTLSGuide() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  ENVOY TLS GUIDE")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n[TLS MODES]")
	fmt.Println(strings.Repeat("-", 40))
	modes := [][]string{
		{"terminate", "HTTPS in → Envoy decrypts → HTTP out to upstream. Most common."},
		{"passthrough", "TLS passthrough via SNI routing. Upstream handles decryption."},
		{"mtls", "Mutual TLS: verify both client and upstream certs. Best for zero-trust."},
		{"originate", "HTTP in → Envoy encrypts → HTTPS out. Useful when upstream requires TLS."},
	}
	for _, m := range modes {
		fmt.Printf("  %-15s %s\n", m[0], m[1])
	}

	fmt.Println("\n[GENERATE SELF-SIGNED CERTS FOR TESTING]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # CA key + cert")
	fmt.Println("  openssl genrsa -out ca.key 4096")
	fmt.Println("  openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj '/CN=MyCA'")
	fmt.Println("")
	fmt.Println("  # Server key + cert signed by CA")
	fmt.Println("  openssl genrsa -out server.key 2048")
	fmt.Println("  openssl req -new -key server.key -out server.csr -subj '/CN=localhost'")
	fmt.Println("  openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt")
	fmt.Println("")
	fmt.Println("  # Client key + cert (mTLS)")
	fmt.Println("  openssl genrsa -out client.key 2048")
	fmt.Println("  openssl req -new -key client.key -out client.csr -subj '/CN=client'")
	fmt.Println("  openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt")

	fmt.Println("\n[TLS VERSION STRINGS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  TLSv1_2  (recommended minimum)")
	fmt.Println("  TLSv1_3  (most secure, may break old clients)")

	fmt.Println("\n[VERIFY TLS]")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  # Test TLS termination")
	fmt.Println("  curl -v --cacert ca.crt https://localhost:10000/")
	fmt.Println("")
	fmt.Println("  # Test mTLS (provide client cert)")
	fmt.Println("  curl -v --cacert ca.crt --cert client.crt --key client.key https://localhost:10000/")
	fmt.Println("")
	fmt.Println("  # Check TLS stats via admin")
	fmt.Println("  curl http://localhost:9901/stats | grep ssl")

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	cluster := flag.String("cluster", "backend_cluster", "Upstream cluster name")
	listenerPort := flag.Int("port", 10000, "Listener port")
	adminPort := flag.Int("admin-port", 9901, "Admin port")
	mode := flag.String("mode", "terminate", "TLS mode: terminate | passthrough | mtls | originate")
	certChain := flag.String("cert", "/etc/envoy/certs/server.crt", "Server certificate chain path")
	privateKey := flag.String("key", "/etc/envoy/certs/server.key", "Server private key path")
	caCert := flag.String("ca", "/etc/envoy/certs/ca.crt", "CA certificate path (for mTLS)")
	sniNames := flag.String("sni", "example.com,www.example.com", "Comma-separated SNI names (passthrough mode)")
	upstreamHost := flag.String("upstream-host", "service.example.com", "Upstream host")
	upstreamPort := flag.Int("upstream-port", 8080, "Upstream port")
	minTLS := flag.String("min-tls", "TLSv1_2", "Minimum TLS version: TLSv1_2 | TLSv1_3")
	output := flag.String("output", "", "Write config to file (default: stdout)")
	guide := flag.Bool("guide", false, "Print TLS guide")
	flag.Parse()

	if *guide {
		printTLSGuide()
		return
	}

	sniList := strings.Split(*sniNames, ",")

	cfg := TLSConfig{
		ClusterName:   *cluster,
		ListenerPort:  *listenerPort,
		AdminPort:     *adminPort,
		Mode:          TLSMode(*mode),
		UpstreamHost:  *upstreamHost,
		UpstreamPort:  *upstreamPort,
		MinTLSVersion: *minTLS,
		AlpnProtocols: []string{"h2", "http/1.1"},
		Certs: CertConfig{
			CertChainPath:  *certChain,
			PrivateKeyPath: *privateKey,
			CACertPath:     *caCert,
			SNINames:       sniList,
		},
	}

	config := GenerateTLSConfig(cfg)

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
