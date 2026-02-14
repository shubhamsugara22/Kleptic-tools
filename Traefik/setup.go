package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		exitErr("failed to get working directory", err)
	}
	if strings.HasSuffix(rootDir, "Traefik") == false {
		// Ensure we run from the Traefik folder if invoked elsewhere.
		if err := os.Chdir(filepath.Join(rootDir, "Traefik")); err == nil {
			rootDir, _ = os.Getwd()
		}
	}

	networkName := getenvDefault("TRAEFIK_NETWORK", "traefik-network")
	traefikVersion := getenvDefault("TRAEFIK_VERSION", "v3.0")
	dashboardPort := getenvDefault("TRAEFIK_DASHBOARD_PORT", "8080")
	acmeEmail := getenvDefault("TRAEFIK_ACME_EMAIL", "your-email@example.com")

	ensureDocker()
	composeBin := resolveCompose()

	ensureNetwork(networkName)
	ensureTraefikConfig(networkName)
	ensureComposeConfig(traefikVersion, dashboardPort, networkName)
	ensureAcme()
	injectAcmeEmail(acmeEmail)

	runCmd(composeBin, "up", "-d")
	fmt.Printf("Traefik is running. Dashboard: http://localhost:%s\n", dashboardPort)
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

func ensureTraefikConfig(network string) {
	if fileExists("traefik.yml") {
		return
	}
	content := fmt.Sprintf(`api:
  dashboard: true
  insecure: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: %s

log:
  level: INFO

accessLog: {}
`, network)
	writeFile("traefik.yml", content)
}

func ensureComposeConfig(version, dashboardPort, network string) {
	if fileExists("docker-compose.yml") {
		return
	}
	content := fmt.Sprintf(`version: '3'

services:
  traefik:
    image: traefik:%s
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "%s:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/traefik.yml:ro
      - ./acme.json:/acme.json
    networks:
      - %s

networks:
  %s:
    external: true
`, version, dashboardPort, network, network)
	writeFile("docker-compose.yml", content)
}

func ensureAcme() {
	if fileExists("acme.json") {
		return
	}
	writeFile("acme.json", "")
	if runtime.GOOS != "windows" {
		_ = os.Chmod("acme.json", 0o600)
	}
}

func injectAcmeEmail(email string) {
	data, err := os.ReadFile("traefik.yml")
	if err != nil {
		return
	}
	if !bytes.Contains(data, []byte("your-email@example.com")) {
		return
	}
	updated := strings.ReplaceAll(string(data), "your-email@example.com", email)
	writeFile("traefik.yml", updated)
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
