package helpers

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"net"
)

// GetLoadBalancerIp returns the IP of the load balancer from the load balancer ingress object
func GetLoadBalancerIp(loadBalancerIngress map[string]interface{}) (net.IP, error) {
	loadBalancerIP, err := getIpBasedLoadBalancerIp(loadBalancerIngress)

	if err == nil {
		return loadBalancerIP, nil
	} else {
		return getDnsBasedLoadBalancerIp(loadBalancerIngress)
	}
}

func getIpBasedLoadBalancerIp(lbIngress map[string]interface{}) (net.IP, error) {
	loadBalancerIP, found, err := unstructured.NestedString(lbIngress, "ip")

	if err != nil || !found {
		return nil, fmt.Errorf("could not get IP based load balancer IP: %s", err)
	}

	ip := net.ParseIP(loadBalancerIP)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse IP from load balancer IP %s", loadBalancerIP)
	}

	return ip, nil
}

func getDnsBasedLoadBalancerIp(lbIngress map[string]interface{}) (net.IP, error) {
	loadBalancerHostname, found, err := unstructured.NestedString(lbIngress, "hostname")

	if err != nil || !found {
		return nil, fmt.Errorf("could not get DNS based load balancer IP: %s", err)
	}

	ips, err := net.LookupIP(loadBalancerHostname)
	if err != nil || len(ips) < 1 {
		return nil, fmt.Errorf("could not get IPs by load balancer hostname: %s", err)
	}

	return ips[0], nil
}

func GetLoadBalancerIngress(k8sClient dynamic.Interface, svcName string, svcNamespace string) (map[string]interface{}, error) {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	svc, err := k8sClient.Resource(res).Namespace(svcNamespace).Get(context.Background(), svcName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("service %s was not found in namespace %s: %w", svcName, svcNamespace, err)
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil || !found {
		return nil, fmt.Errorf("could not get load balancer status from the service %s: %w", svcName, err)
	}
	loadBalancerIngress, _ := ingress[0].(map[string]interface{})

	return loadBalancerIngress, nil
}
