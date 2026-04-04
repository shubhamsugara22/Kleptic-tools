package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		exitErr("failed to get working directory", err)
	}
	if strings.HasSuffix(rootDir, "Loki") == false {
		// Ensure we run from the Loki folder if invoked elsewhere.
		if err := os.Chdir(filepath.Join(rootDir, "Loki")); err == nil {
			rootDir, _ = os.Getwd()
		}
	}

	networkName := getenvDefault("LOKI_NETWORK", "loki-network")
	lokiVersion := getenvDefault("LOKI_VERSION", "latest")
	nginxHTTPPort := getenvDefault("NGINX_HTTP_PORT", "8080")
	nginxHTTPSPort := getenvDefault("NGINX_HTTPS_PORT", "8443")
	grafanaEnabled := getenvDefault("GRAFANA_ENABLED", "false")
	grafanaPort := getenvDefault("GRAFANA_PORT", "3000")

	ensureDocker()
	composeBin := resolveCompose()

	ensureNetwork(networkName)
	ensureLokiConfig()
	ensurePromtailConfig()
	ensurePrometheusConfig()
	ensureNginxConfig()
	ensureComposeConfig(lokiVersion, networkName, grafanaEnabled, grafanaPort, nginxHTTPPort, nginxHTTPSPort)

	runCmd(composeBin, "up", "-d")
	fmt.Printf("Loki is available behind Nginx: http://localhost:%s and https://localhost:%s\n", nginxHTTPPort, nginxHTTPSPort)
	if grafanaEnabled == "true" {
		fmt.Printf("Grafana is running on http://localhost:%s\n", grafanaPort)
	}
}

func getenvDefault(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

func ensureDocker() {
	if _, err := exec.LookPath("docker"); err != nil {
		exitErr("Docker is required but not installed", err)
	}
}

func resolveCompose() string {
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "compose", "version")
		if err := cmd.Run(); err == nil {
			return "docker compose"
		}
	}
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return "docker-compose"
	}
	exitErr("Docker Compose is required but not installed", nil)
	return ""
}

func ensureNetwork(network string) {
	cmd := exec.Command("docker", "network", "inspect", network)
	if err := cmd.Run(); err == nil {
		return
	}
	runCmd("docker", "network", "create", network)
}

func ensureLokiConfig() {
	if fileExists("loki-config.yml") {
		return
	}
	content := `auth_enabled: true

ingester:
  chunk_idle_period: 3m
  chunk_retain_period: 1m
  max_chunk_age: 1h
  chunk_encoding: snappy

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

server:
  http_listen_port: 3100
  log_level: info

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks

chunk_store_config:
  max_look_back_period: 0s

table_manager:
	retention_deletes_enabled: true
	retention_period: 168h
`
	writeFile("loki-config.yml", content)
}

func ensurePromtailConfig() {
	if fileExists("promtail-config.yml") {
		return
	}
	content := `server:
	http_listen_port: 9080
	grpc_listen_port: 0

positions:
	filename: /tmp/positions.yaml

clients:
	- url: http://loki:3100/loki/api/v1/push

scrape_configs:
	- job_name: system-logs
		static_configs:
			- targets:
					- localhost
				labels:
					job: varlogs
					__path__: /var/log/*.log
`
	writeFile("promtail-config.yml", content)
}

func ensurePrometheusConfig() {
	if fileExists("prometheus.yml") {
		return
	}
	content := `global:
	scrape_interval: 15s

scrape_configs:
	- job_name: 'prometheus'
		static_configs:
			- targets: ['localhost:9090']

	- job_name: 'loki'
		static_configs:
			- targets: ['loki:3100']

	- job_name: 'promtail'
		static_configs:
			- targets: ['promtail:9080']
`
	writeFile("prometheus.yml", content)
}

func ensureNginxConfig() {
	if fileExists("nginx.conf") {
		return
	}
	content := `events {}
http {
		log_format audit '$remote_addr - $remote_user [$time_local] "$request" '
										 '$status $body_bytes_sent "$http_referer" '
										 '"$http_user_agent" "$http_x_forwarded_for"';

		access_log /var/log/nginx/access.log audit;
		error_log /var/log/nginx/error.log warn;

		limit_req_zone $binary_remote_addr zone=loki_rate:10m rate=20r/s;

		server {
				listen 8080;
				server_name _;
				return 301 https://$host:8443$request_uri;
		}

		server {
				listen 8443 ssl;
				server_name _;

				ssl_certificate /etc/nginx/certs/loki.crt;
				ssl_certificate_key /etc/nginx/certs/loki.key;
				ssl_protocols TLSv1.2 TLSv1.3;
				ssl_ciphers HIGH:!aNULL:!MD5;

				auth_basic "Loki Auth";
				auth_basic_user_file /etc/nginx/.htpasswd;

				location / {
						limit_req zone=loki_rate burst=40 nodelay;
						proxy_pass http://loki:3100;
						proxy_set_header Host $host;
						proxy_set_header X-Real-IP $remote_addr;
						proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
						proxy_set_header X-Forwarded-Proto $scheme;
				}
		}
}
`
	writeFile("nginx.conf", content)
}

func ensureComposeConfig(version, network, grafanaEnabled, grafanaPort, nginxHTTPPort, nginxHTTPSPort string) {
	if fileExists("docker-compose.yml") {
		return
	}

	lokiService := fmt.Sprintf(`version: '3'

services:
	loki:
		image: grafana/loki:%s
		container_name: loki
		restart: unless-stopped
		volumes:
			- ./loki-config.yml:/etc/loki/local-config.yml:ro
			- loki-data:/loki
		environment:
			- LOKI_CONFIG_FILE=/etc/loki/local-config.yml
		networks:
			- %s

	promtail:
		image: grafana/promtail:latest
		container_name: promtail
		restart: unless-stopped
		volumes:
			- ./promtail-config.yml:/etc/promtail/promtail-config.yml:ro
			- /var/log:/var/log:ro
		command: -config.file=/etc/promtail/promtail-config.yml
		depends_on:
			- loki
		networks:
			- %s

	prometheus:
		image: prom/prometheus:latest
		container_name: prometheus
		restart: unless-stopped
		ports:
			- "9090:9090"
		volumes:
			- ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
		networks:
			- %s

	nginx:
		image: nginx:latest
		container_name: loki-nginx
		restart: unless-stopped
		ports:
			- "%s:8080"
			- "%s:8443"
		volumes:
			- ./nginx.conf:/etc/nginx/nginx.conf:ro
			- ./nginx.htpasswd:/etc/nginx/.htpasswd:ro
			- ./certs:/etc/nginx/certs:ro
		depends_on:
			- loki
		networks:
			- %s
`, version, network, network, network, nginxHTTPPort, nginxHTTPSPort, network)

	var fullContent string

	if grafanaEnabled == "true" {
		fullContent = fmt.Sprintf(`%s
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "%s:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - loki
    networks:
      - %s

volumes:
  loki-data:
  grafana-data:

networks:
  %s:
    external: true
`, lokiService, grafanaPort, network, network)
	} else {
		fullContent = fmt.Sprintf(`%s
volumes:
  loki-data:

networks:
  %s:
    external: true
`, lokiService, network)
	}

	writeFile("docker-compose.yml", fullContent)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		exitErr("failed to write file: "+path, err)
	}
}

func runCmd(command string, args ...string) {
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], append(parts[1:], args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		exitErr("command failed: "+command+" "+strings.Join(args, " "), err)
	}
}

func exitErr(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}
