package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Colors for terminal output
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[0;32m"
	colorBlue   = "\033[0;34m"
	colorYellow = "\033[1;33m"
)

var (
	envoyConfigContent = `static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.router
          route_config:
            name: local_route
            virtual_hosts:
            - name: backend
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: backend_service
  clusters:
  - name: backend_service
    type: STATIC
    load_assignment:
      cluster_name: backend_service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
admin:
  access_log_path: /tmp/admin_access.log
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
`
)

func main() {
	fmt.Printf("%s=== Envoy Proxy Setup ===%s\n\n", colorBlue, colorReset)

	osType := runtime.GOOS
	fmt.Printf("%sDetected OS: %s%s\n\n", colorYellow, osType, colorReset)

	// Check for Docker
	dockerAvailable := isCommandAvailable("docker")
	if dockerAvailable {
		fmt.Printf("%s✓ Docker is installed%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%s✗ Docker is not installed%s\n", colorYellow, colorReset)
	}

	fmt.Println("")

	// Install Envoy based on OS
	installEnvoy(osType, dockerAvailable)

	// Check if Envoy is installed
	if isCommandAvailable("envoy") {
		fmt.Printf("%s✓ Envoy is installed%s\n", colorGreen, colorReset)
		if version, err := getEnvoyVersion(); err == nil {
			fmt.Printf("  %s\n\n", version)
		} else {
			fmt.Println("")
		}
	} else {
		fmt.Printf("%s⚠ Envoy command not found in PATH%s\n", colorYellow, colorReset)
		fmt.Printf("  Ensure Envoy is properly installed and added to your PATH\n\n")
	}

	// Create sample configuration
	createSampleConfig()

	// Setup complete
	fmt.Printf("\n%s=== Setup Complete ===%s\n\n", colorGreen, colorReset)
	fmt.Println("Next steps:")
	fmt.Println("1. Review and modify envoy.yaml as needed")
	fmt.Println("2. Start Envoy:")
	fmt.Println("   envoy -c envoy.yaml")
	fmt.Println("3. Access admin interface:")
	fmt.Println("   http://localhost:9901")
	fmt.Println("")
	fmt.Println("For more information, visit: https://www.envoyproxy.io/")
}

func isCommandAvailable(name string) bool {
	cmd := "which"
	if runtime.GOOS == "windows" {
		cmd = "where"
	}

	command := exec.Command(cmd, name)
	if err := command.Run(); err != nil {
		return false
	}
	return true
}

func getEnvoyVersion() (string, error) {
	cmd := exec.Command("envoy", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func installEnvoy(osType string, dockerAvailable bool) {
	fmt.Printf("%sInstalling Envoy on %s...%s\n\n", colorBlue, osType, colorReset)

	switch osType {
	case "linux":
		installEnvoyLinux(dockerAvailable)
	case "darwin":
		installEnvoyMacOS(dockerAvailable)
	case "windows":
		installEnvoyWindows(dockerAvailable)
	default:
		fmt.Printf("%sUnknown OS: %s%s\n", colorYellow, osType, colorReset)
		fmt.Println("Please visit: https://www.envoyproxy.io/docs/envoy/latest/install/install")
	}
}

func installEnvoyLinux(dockerAvailable bool) {
	// Try apt-get
	if isCommandAvailable("apt-get") {
		fmt.Println("Installing Envoy via apt...")
		cmd := exec.Command("apt-get", "update")
		_ = cmd.Run()

		cmd = exec.Command("apt-get", "install", "-y", "envoy")
		if err := cmd.Run(); err == nil {
			fmt.Printf("%s✓ Envoy installed via apt%s\n", colorGreen, colorReset)
			return
		}
	}

	// Try yum
	if isCommandAvailable("yum") {
		fmt.Println("Installing Envoy via yum...")
		cmd := exec.Command("yum", "install", "-y", "envoy")
		if err := cmd.Run(); err == nil {
			fmt.Printf("%s✓ Envoy installed via yum%s\n", colorGreen, colorReset)
			return
		}
	}

	fmt.Printf("%sCould not find package manager. Please install Envoy manually.%s\n", colorYellow, colorReset)
	fmt.Println("Visit: https://www.envoyproxy.io/docs/envoy/latest/install/install")
}

func installEnvoyMacOS(dockerAvailable bool) {
	if isCommandAvailable("brew") {
		fmt.Println("Installing Envoy via Homebrew...")
		cmd := exec.Command("brew", "tap", "envoyproxy/envoy")
		_ = cmd.Run()

		cmd = exec.Command("brew", "install", "envoy")
		if err := cmd.Run(); err == nil {
			fmt.Printf("%s✓ Envoy installed via Homebrew%s\n", colorGreen, colorReset)
			return
		}
	}

	fmt.Printf("%sBrew not found. Please install Homebrew first.%s\n", colorYellow, colorReset)
	fmt.Println("Visit: https://brew.sh")
}

func installEnvoyWindows(dockerAvailable bool) {
	if isCommandAvailable("choco") {
		fmt.Println("Installing Envoy via Chocolatey...")
		cmd := exec.Command("choco", "install", "envoy", "-y")
		if err := cmd.Run(); err == nil {
			fmt.Printf("%s✓ Envoy installed via Chocolatey%s\n", colorGreen, colorReset)
			return
		}
	}

	if dockerAvailable {
		fmt.Println("Docker available! You can run Envoy in a container:")
		fmt.Println(`docker run -v $(pwd)/envoy.yaml:/etc/envoy/envoy.yaml -p 10000:10000 -p 9901:9901 envoyproxy/envoy:v1.27-latest`)
		return
	}

	fmt.Printf("%sCould not auto-install Envoy on Windows.%s\n", colorYellow, colorReset)
	fmt.Println("Please use one of these methods:")
	fmt.Println("1. Install Chocolatey and run: choco install envoy")
	fmt.Println("2. Use Docker: docker pull envoyproxy/envoy")
	fmt.Println("3. Download from: https://www.envoyproxy.io/docs/envoy/latest/install/install")
}

func createSampleConfig() {
	configPath := filepath.Join(".", "envoy.yaml")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s✓ envoy.yaml already exists%s\n\n", colorGreen, colorReset)
		return
	}

	fmt.Printf("%sCreating sample envoy.yaml configuration...%s\n\n", colorBlue, colorReset)

	err := ioutil.WriteFile(configPath, []byte(envoyConfigContent), 0644)
	if err != nil {
		fmt.Printf("%s✗ Error creating envoy.yaml: %v%s\n\n", colorYellow, err, colorReset)
		return
	}

	fmt.Printf("%s✓ Sample configuration created: envoy.yaml%s\n", colorGreen, colorReset)
}
