package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	networkingv1beta1 "istio.io/api/networking/v1beta1"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// IstioTrafficHandler manages Istio traffic configurations
type IstioTrafficHandler struct {
	client    *versionedclient.Clientset
	namespace string
}

// NewIstioTrafficHandler creates a new traffic handler
func NewIstioTrafficHandler(namespace string) (*IstioTrafficHandler, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = home + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %v", err)
	}

	istioClient, err := versionedclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Istio client: %v", err)
	}

	return &IstioTrafficHandler{
		client:    istioClient,
		namespace: namespace,
	}, nil
}

// CreateCanaryDeployment creates a canary deployment with traffic splitting
func (h *IstioTrafficHandler) CreateCanaryDeployment(serviceName string, stableWeight, canaryWeight int32) error {
	ctx := context.Background()

	virtualService := &networkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-canary",
			Namespace: h.namespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{serviceName},
			Http: []*networkingv1beta1.HTTPRoute{
				{
					Route: []*networkingv1beta1.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1.Destination{
								Host:   serviceName,
								Subset: "stable",
							},
							Weight: stableWeight,
						},
						{
							Destination: &networkingv1beta1.Destination{
								Host:   serviceName,
								Subset: "canary",
							},
							Weight: canaryWeight,
						},
					},
				},
			},
		},
	}

	_, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Create(ctx, virtualService, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create virtual service: %v", err)
	}

	fmt.Printf("✓ Created canary deployment for %s (stable: %d%%, canary: %d%%)\n", serviceName, stableWeight, canaryWeight)
	return nil
}

// UpdateTrafficWeights updates traffic distribution between versions
func (h *IstioTrafficHandler) UpdateTrafficWeights(serviceName string, stableWeight, canaryWeight int32) error {
	ctx := context.Background()

	vsName := serviceName + "-canary"
	vs, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Get(ctx, vsName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get virtual service: %v", err)
	}

	// Update weights
	vs.Spec.Http[0].Route[0].Weight = stableWeight
	vs.Spec.Http[0].Route[1].Weight = canaryWeight

	_, err = h.client.NetworkingV1beta1().VirtualServices(h.namespace).Update(ctx, vs, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update virtual service: %v", err)
	}

	fmt.Printf("✓ Updated traffic weights for %s (stable: %d%%, canary: %d%%)\n", serviceName, stableWeight, canaryWeight)
	return nil
}

// CreateDestinationRule creates destination rule with subsets for canary deployment
func (h *IstioTrafficHandler) CreateDestinationRule(serviceName string) error {
	ctx := context.Background()

	destinationRule := &networkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: h.namespace,
		},
		Spec: networkingv1beta1.DestinationRule{
			Host: serviceName,
			Subsets: []*networkingv1beta1.Subset{
				{
					Name: "stable",
					Labels: map[string]string{
						"version": "stable",
					},
				},
				{
					Name: "canary",
					Labels: map[string]string{
						"version": "canary",
					},
				},
			},
			TrafficPolicy: &networkingv1beta1.TrafficPolicy{
				LoadBalancer: &networkingv1beta1.LoadBalancerSettings{
					LbPolicy: &networkingv1beta1.LoadBalancerSettings_Simple{
						Simple: networkingv1beta1.LoadBalancerSettings_ROUND_ROBIN,
					},
				},
				ConnectionPool: &networkingv1beta1.ConnectionPoolSettings{
					Tcp: &networkingv1beta1.ConnectionPoolSettings_TCPSettings{
						MaxConnections: 100,
					},
					Http: &networkingv1beta1.ConnectionPoolSettings_HTTPSettings{
						Http1MaxPendingRequests: 50,
						MaxRequestsPerConnection: 2,
					},
				},
			},
		},
	}

	_, err := h.client.NetworkingV1beta1().DestinationRules(h.namespace).Create(ctx, destinationRule, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create destination rule: %v", err)
	}

	fmt.Printf("✓ Created destination rule for %s\n", serviceName)
	return nil
}

// ApplyCircuitBreaker applies circuit breaker settings to a service
func (h *IstioTrafficHandler) ApplyCircuitBreaker(serviceName string, maxConnections, maxPendingRequests int32) error {
	ctx := context.Background()

	destinationRule := &networkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-circuit-breaker",
			Namespace: h.namespace,
		},
		Spec: networkingv1beta1.DestinationRule{
			Host: serviceName,
			TrafficPolicy: &networkingv1beta1.TrafficPolicy{
				ConnectionPool: &networkingv1beta1.ConnectionPoolSettings{
					Tcp: &networkingv1beta1.ConnectionPoolSettings_TCPSettings{
						MaxConnections: maxConnections,
					},
					Http: &networkingv1beta1.ConnectionPoolSettings_HTTPSettings{
						Http1MaxPendingRequests:  maxPendingRequests,
						MaxRequestsPerConnection: 1,
					},
				},
				OutlierDetection: &networkingv1beta1.OutlierDetection{
					Consecutive_5XxErrors: &networkingv1beta1.UInt32Value{Value: 5},
					Interval:              &networkingv1beta1.Duration{Seconds: 30},
					BaseEjectionTime:      &networkingv1beta1.Duration{Seconds: 30},
					MaxEjectionPercent:    50,
					MinHealthPercent:      0,
				},
			},
		},
	}

	_, err := h.client.NetworkingV1beta1().DestinationRules(h.namespace).Create(ctx, destinationRule, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create circuit breaker: %v", err)
	}

	fmt.Printf("✓ Applied circuit breaker to %s\n", serviceName)
	return nil
}

// CreateRetryPolicy creates a retry policy for a service
func (h *IstioTrafficHandler) CreateRetryPolicy(serviceName string, attempts int32, timeout string) error {
	ctx := context.Background()

	virtualService := &networkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-retry",
			Namespace: h.namespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{serviceName},
			Http: []*networkingv1beta1.HTTPRoute{
				{
					Route: []*networkingv1beta1.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1.Destination{
								Host: serviceName,
							},
						},
					},
					Retries: &networkingv1beta1.HTTPRetry{
						Attempts:      attempts,
						PerTryTimeout: &networkingv1beta1.Duration{Seconds: 2},
					},
					Timeout: &networkingv1beta1.Duration{Seconds: 10},
				},
			},
		},
	}

	_, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Create(ctx, virtualService, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create retry policy: %v", err)
	}

	fmt.Printf("✓ Created retry policy for %s (attempts: %d)\n", serviceName, attempts)
	return nil
}

// CreateHeaderBasedRouting creates header-based routing rules
func (h *IstioTrafficHandler) CreateHeaderBasedRouting(serviceName, headerName, headerValue, targetSubset string) error {
	ctx := context.Background()

	virtualService := &networkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-header-routing",
			Namespace: h.namespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{serviceName},
			Http: []*networkingv1beta1.HTTPRoute{
				{
					Match: []*networkingv1beta1.HTTPMatchRequest{
						{
							Headers: map[string]*networkingv1beta1.StringMatch{
								headerName: {
									MatchType: &networkingv1beta1.StringMatch_Exact{
										Exact: headerValue,
									},
								},
							},
						},
					},
					Route: []*networkingv1beta1.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1.Destination{
								Host:   serviceName,
								Subset: targetSubset,
							},
						},
					},
				},
				{
					Route: []*networkingv1beta1.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1.Destination{
								Host:   serviceName,
								Subset: "default",
							},
						},
					},
				},
			},
		},
	}

	_, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Create(ctx, virtualService, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create header-based routing: %v", err)
	}

	fmt.Printf("✓ Created header-based routing for %s (header: %s=%s -> %s)\n", serviceName, headerName, headerValue, targetSubset)
	return nil
}

// GradualCanaryRollout performs gradual canary rollout
func (h *IstioTrafficHandler) GradualCanaryRollout(serviceName string, stages []int32, interval time.Duration) error {
	fmt.Printf("Starting gradual canary rollout for %s\n", serviceName)

	for i, canaryWeight := range stages {
		stableWeight := 100 - canaryWeight
		fmt.Printf("\n[Stage %d/%d] Shifting to %d%% canary, %d%% stable\n", i+1, len(stages), canaryWeight, stableWeight)

		err := h.UpdateTrafficWeights(serviceName, stableWeight, canaryWeight)
		if err != nil {
			return fmt.Errorf("failed at stage %d: %v", i+1, err)
		}

		if i < len(stages)-1 {
			fmt.Printf("Waiting %v before next stage...\n", interval)
			time.Sleep(interval)
		}
	}

	fmt.Printf("\n✓ Canary rollout completed successfully!\n")
	return nil
}

// ListVirtualServices lists all virtual services in the namespace
func (h *IstioTrafficHandler) ListVirtualServices() error {
	ctx := context.Background()

	vsList, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list virtual services: %v", err)
	}

	fmt.Printf("\nVirtual Services in namespace '%s':\n", h.namespace)
	fmt.Println("-----------------------------------")
	for _, vs := range vsList.Items {
		fmt.Printf("- %s (hosts: %v)\n", vs.Name, vs.Spec.Hosts)
	}

	return nil
}

// DeleteVirtualService deletes a virtual service
func (h *IstioTrafficHandler) DeleteVirtualService(name string) error {
	ctx := context.Background()

	err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete virtual service: %v", err)
	}

	fmt.Printf("✓ Deleted virtual service: %s\n", name)
	return nil
}

func main() {
	// Set namespace (default to "default")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	handler, err := NewIstioTrafficHandler(namespace)
	if err != nil {
		log.Fatalf("Failed to create traffic handler: %v", err)
	}

	// Example usage
	serviceName := "myapp"

	fmt.Println("=== Istio Traffic Handler Demo ===\n")

	// 1. Create destination rule with subsets
	fmt.Println("1. Creating destination rule...")
	if err := handler.CreateDestinationRule(serviceName); err != nil {
		log.Printf("Error: %v", err)
	}

	// 2. Create canary deployment with initial traffic split
	fmt.Println("\n2. Creating canary deployment...")
	if err := handler.CreateCanaryDeployment(serviceName, 90, 10); err != nil {
		log.Printf("Error: %v", err)
	}

	// 3. Apply circuit breaker
	fmt.Println("\n3. Applying circuit breaker...")
	if err := handler.ApplyCircuitBreaker(serviceName, 100, 50); err != nil {
		log.Printf("Error: %v", err)
	}

	// 4. Create retry policy
	fmt.Println("\n4. Creating retry policy...")
	if err := handler.CreateRetryPolicy(serviceName+"-api", 3, "10s"); err != nil {
		log.Printf("Error: %v", err)
	}

	// 5. List all virtual services
	fmt.Println("\n5. Listing virtual services...")
	if err := handler.ListVirtualServices(); err != nil {
		log.Printf("Error: %v", err)
	}

	// 6. Perform gradual canary rollout (commented out to avoid automatic execution)
	/*
		fmt.Println("\n6. Starting gradual canary rollout...")
		stages := []int32{10, 25, 50, 75, 100}
		interval := 30 * time.Second
		if err := handler.GradualCanaryRollout(serviceName, stages, interval); err != nil {
			log.Printf("Error: %v", err)
		}
	*/

	fmt.Println("\n=== Demo completed ===")
	fmt.Println("\nNote: Modify the main function to suit your specific needs.")
	fmt.Println("Set NAMESPACE environment variable to target a specific namespace.")
}
