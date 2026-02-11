package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// Kong Admin API endpoint
	kongAdminURL = "http://localhost:8001"
	// Kong Proxy endpoint
	kongProxyURL = "http://localhost:8000"
)

// Service represents a Kong service
type Service struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Protocol string `json:"protocol,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
}

// Route represents a Kong route
type Route struct {
	Name      string   `json:"name"`
	Paths     []string `json:"paths"`
	Methods   []string `json:"methods,omitempty"`
	Protocols []string `json:"protocols,omitempty"`
}

// Plugin represents a Kong plugin
type Plugin struct {
	Name    string                 `json:"name"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Enabled bool                   `json:"enabled"`
}

func main() {
	fmt.Println("🚀 Kong Gateway Demo - Basic Functionality")
	fmt.Println("==========================================\n")

	// Step 1: Check Kong Gateway status
	fmt.Println("📊 Step 1: Checking Kong Gateway status...")
	if err := checkKongStatus(); err != nil {
		log.Fatalf("Kong Gateway is not running: %v\n", err)
	}
	fmt.Println("✅ Kong Gateway is running!\n")

	// Step 2: Create a service
	fmt.Println("🔧 Step 2: Creating a service...")
	service := Service{
		Name: "example-service",
		URL:  "http://httpbin.org",
	}
	serviceID, err := createService(service)
	if err != nil {
		log.Printf("⚠️  Warning: Could not create service: %v\n", err)
	} else {
		fmt.Printf("✅ Service created with ID: %s\n\n", serviceID)
	}

	// Step 3: Create a route
	fmt.Println("🛣️  Step 3: Creating a route for the service...")
	route := Route{
		Name:    "example-route",
		Paths:   []string{"/httpbin"},
		Methods: []string{"GET", "POST"},
	}
	routeID, err := createRoute("example-service", route)
	if err != nil {
		log.Printf("⚠️  Warning: Could not create route: %v\n", err)
	} else {
		fmt.Printf("✅ Route created with ID: %s\n\n", routeID)
	}

	// Step 4: Test the route through Kong proxy
	fmt.Println("🌐 Step 4: Testing the route through Kong Gateway...")
	time.Sleep(1 * time.Second) // Give Kong a moment to register the route
	if err := testProxyRequest(); err != nil {
		log.Printf("⚠️  Warning: Could not test proxy request: %v\n", err)
	} else {
		fmt.Println("✅ Successfully made request through Kong Gateway!\n")
	}

	// Step 5: Add a plugin (Rate Limiting)
	fmt.Println("🔌 Step 5: Adding rate limiting plugin...")
	plugin := Plugin{
		Name: "rate-limiting",
		Config: map[string]interface{}{
			"minute": 5,
			"policy": "local",
		},
		Enabled: true,
	}
	pluginID, err := addPlugin("example-service", plugin)
	if err != nil {
		log.Printf("⚠️  Warning: Could not add plugin: %v\n", err)
	} else {
		fmt.Printf("✅ Rate limiting plugin added with ID: %s\n\n", pluginID)
	}

	// Step 6: List all services
	fmt.Println("📋 Step 6: Listing all services...")
	if err := listServices(); err != nil {
		log.Printf("⚠️  Warning: Could not list services: %v\n", err)
	}

	// Step 7: Test rate limiting
	fmt.Println("\n🧪 Step 7: Testing rate limiting (making 6 requests)...")
	testRateLimiting()

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("=====================================")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   - Visit http://localhost:8001 for Kong Admin API")
	fmt.Println("   - Visit http://localhost:8000/httpbin/get for proxied requests")
	fmt.Println("   - Explore more plugins and configurations")
}

// checkKongStatus checks if Kong Gateway is running
func checkKongStatus() error {
	resp, err := http.Get(kongAdminURL + "/status")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Kong returned status: %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	fmt.Printf("   Database: %v\n", status["database"])
	fmt.Printf("   Server: %v\n", status["server"])
	return nil
}

// createService creates a new service in Kong
func createService(service Service) (string, error) {
	jsonData, err := json.Marshal(service)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(kongAdminURL+"/services", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return "", fmt.Errorf("failed to create service: %s (status: %d)", body, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// createRoute creates a new route for a service
func createRoute(serviceName string, route Route) (string, error) {
	jsonData, err := json.Marshal(route)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/services/%s/routes", kongAdminURL, serviceName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return "", fmt.Errorf("failed to create route: %s (status: %d)", body, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// testProxyRequest makes a test request through Kong proxy
func testProxyRequest() error {
	resp, err := http.Get(kongProxyURL + "/httpbin/get")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy request failed with status: %d", resp.StatusCode)
	}

	fmt.Printf("   Response status: %d\n", resp.StatusCode)
	fmt.Printf("   Response preview: %.100s...\n", string(body))
	return nil
}

// addPlugin adds a plugin to a service
func addPlugin(serviceName string, plugin Plugin) (string, error) {
	jsonData, err := json.Marshal(plugin)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/services/%s/plugins", kongAdminURL, serviceName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return "", fmt.Errorf("failed to add plugin: %s (status: %d)", body, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// listServices lists all services in Kong
func listServices() error {
	resp, err := http.Get(kongAdminURL + "/services")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	services := result["data"].([]interface{})
	fmt.Printf("   Total services: %d\n", len(services))
	
	for i, svc := range services {
		service := svc.(map[string]interface{})
		fmt.Printf("   %d. Name: %s, Host: %s\n", i+1, service["name"], service["host"])
	}

	return nil
}

// testRateLimiting tests the rate limiting by making multiple requests
func testRateLimiting() {
	for i := 1; i <= 6; i++ {
		resp, err := http.Get(kongProxyURL + "/httpbin/get")
		if err != nil {
			log.Printf("   Request %d failed: %v\n", i, err)
			continue
		}

		fmt.Printf("   Request %d: Status %d", i, resp.StatusCode)
		
		// Check rate limit headers
		if remaining := resp.Header.Get("X-RateLimit-Remaining-Minute"); remaining != "" {
			fmt.Printf(" | Remaining: %s", remaining)
		}
		
		if resp.StatusCode == http.StatusTooManyRequests {
			fmt.Printf(" ⛔ Rate limit exceeded!")
		} else {
			fmt.Printf(" ✅")
		}
		fmt.Println()
		
		resp.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}
}
