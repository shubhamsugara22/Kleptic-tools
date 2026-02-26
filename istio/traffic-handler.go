package main

import (
	"context"
	"fmt"
	"log"
	"os"

	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ListVirtualServices lists all virtual services in the namespace
func (h *IstioTrafficHandler) ListVirtualServices() error {
	ctx := context.Background()

	vsList, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list virtual services: %v", err)
	}

	fmt.Printf("\nVirtual Services in namespace '%s':\n", h.namespace)
	fmt.Println("==========================================")
	if len(vsList.Items) == 0 {
		fmt.Println("No VirtualServices found")
	} else {
		for _, vs := range vsList.Items {
			fmt.Printf("- Name: %s\n", vs.Name)
			fmt.Printf("  Hosts: %v\n", vs.Spec.Hosts)
			if len(vs.Spec.Http) > 0 && len(vs.Spec.Http[0].Route) > 0 {
				fmt.Printf("  Routes:\n")
				for _, route := range vs.Spec.Http[0].Route {
					if route.Destination != nil {
						weight := route.Weight
						subset := ""
						if route.Destination.Subset != "" {
							subset = fmt.Sprintf(" (subset: %s)", route.Destination.Subset)
						}
						fmt.Printf("    -> %s%s: %d%%\n", route.Destination.Host, subset, weight)
					}
				}
			}
			fmt.Println()
		}
	}

	return nil
}

// ListDestinationRules lists all destination rules in the namespace
func (h *IstioTrafficHandler) ListDestinationRules() error {
	ctx := context.Background()

	drList, err := h.client.NetworkingV1beta1().DestinationRules(h.namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list destination rules: %v", err)
	}

	fmt.Printf("\nDestination Rules in namespace '%s':\n", h.namespace)
	fmt.Println("==========================================")
	if len(drList.Items) == 0 {
		fmt.Println("No DestinationRules found")
	} else {
		for _, dr := range drList.Items {
			fmt.Printf("- Name: %s\n", dr.Name)
			fmt.Printf("  Host: %s\n", dr.Spec.Host)
			if len(dr.Spec.Subsets) > 0 {
				fmt.Printf("  Subsets:\n")
				for _, subset := range dr.Spec.Subsets {
					fmt.Printf("    - %s (labels: %v)\n", subset.Name, subset.Labels)
				}
			}
			fmt.Println()
		}
	}

	return nil
}

// ListGateways lists all gateways in the namespace
func (h *IstioTrafficHandler) ListGateways() error {
	ctx := context.Background()

	gwList, err := h.client.NetworkingV1beta1().Gateways(h.namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list gateways: %v", err)
	}

	fmt.Printf("\nGateways in namespace '%s':\n", h.namespace)
	fmt.Println("==========================================")
	if len(gwList.Items) == 0 {
		fmt.Println("No Gateways found")
	} else {
		for _, gw := range gwList.Items {
			fmt.Printf("- Name: %s\n", gw.Name)
			fmt.Printf("  Selector: %v\n", gw.Spec.Selector)
			if len(gw.Spec.Servers) > 0 {
				fmt.Printf("  Servers:\n")
				for _, server := range gw.Spec.Servers {
					if server.Port != nil {
						fmt.Printf("    - Port: %d (%s)\n", server.Port.Number, server.Port.Protocol)
					}
					fmt.Printf("      Hosts: %v\n", server.Hosts)
				}
			}
			fmt.Println()
		}
	}

	return nil
}

// DeleteVirtualService deletes a virtual service
func (h *IstioTrafficHandler) DeleteVirtualService(name string) error {
	ctx := context.Background()

	err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete virtual service: %v", err)
	}

	fmt.Printf("✓ Deleted virtual service: %s\n", name)
	return nil
}

// DeleteDestinationRule deletes a destination rule
func (h *IstioTrafficHandler) DeleteDestinationRule(name string) error {
	ctx := context.Background()

	err := h.client.NetworkingV1beta1().DestinationRules(h.namespace).Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete destination rule: %v", err)
	}

	fmt.Printf("✓ Deleted destination rule: %s\n", name)
	return nil
}

// GetVirtualService retrieves a specific virtual service
func (h *IstioTrafficHandler) GetVirtualService(name string) error {
	ctx := context.Background()

	vs, err := h.client.NetworkingV1beta1().VirtualServices(h.namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get virtual service: %v", err)
	}

	fmt.Printf("\nVirtualService: %s\n", vs.Name)
	fmt.Println("==========================================")
	fmt.Printf("Namespace: %s\n", vs.Namespace)
	fmt.Printf("Hosts: %v\n", vs.Spec.Hosts)

	if len(vs.Spec.Http) > 0 {
		fmt.Println("\nHTTP Routes:")
		for i, http := range vs.Spec.Http {
			fmt.Printf("  Route %d:\n", i+1)
			for _, route := range http.Route {
				if route.Destination != nil {
					weight := route.Weight
					fmt.Printf("    - Destination: %s\n", route.Destination.Host)
					if route.Destination.Subset != "" {
						fmt.Printf("      Subset: %s\n", route.Destination.Subset)
					}
					fmt.Printf("      Weight: %d%%\n", weight)
				}
			}
		}
	}

	return nil
}

// ShowStatus displays Istio configuration status
func (h *IstioTrafficHandler) ShowStatus() error {
	fmt.Println("\n=== Istio Traffic Management Status ===")
	fmt.Printf("Namespace: %s\n", h.namespace)
	fmt.Println()

	if err := h.ListVirtualServices(); err != nil {
		log.Printf("Error listing VirtualServices: %v", err)
	}

	if err := h.ListDestinationRules(); err != nil {
		log.Printf("Error listing DestinationRules: %v", err)
	}

	if err := h.ListGateways(); err != nil {
		log.Printf("Error listing Gateways: %v", err)
	}

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

	fmt.Println("=== Istio Traffic Handler ===")
	fmt.Println()

	// Show current status
	if err := handler.ShowStatus(); err != nil {
		log.Printf("Error: %v", err)
	}

	fmt.Println("\n=== Operations completed ===")
	fmt.Println("\nNote: This tool lists existing Istio resources.")
	fmt.Println("To create new resources, use kubectl apply with YAML files.")
	fmt.Println("See traffic-management.md for configuration examples.")
	fmt.Println("\nSet NAMESPACE environment variable to target a specific namespace.")
}
