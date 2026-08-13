package customdomain

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"

	"github.com/avast/retry-go/v4"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func GetLoadBalancerTarget(ctx context.Context, r *resources.Resources, svcName, svcNamespace string) (string, error) {
	svc := &unstructured.Unstructured{}
	svc.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"})

	if err := r.Get(ctx, svcName, svcNamespace, svc); err != nil {
		return "", fmt.Errorf("service %s/%s was not found: %w", svcNamespace, svcName, err)
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil || !found || len(ingress) == 0 {
		return "", fmt.Errorf("could not get load balancer ingress from service %s/%s: %w", svcNamespace, svcName, err)
	}

	loadBalancerIngress, ok := ingress[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("could not parse load balancer ingress object from service %s/%s", svcNamespace, svcName)
	}

	loadBalancerTarget, err := getLoadBalancerTarget(loadBalancerIngress)
	if err != nil {
		return "", err
	}

	return loadBalancerTarget, nil
}

// this either returns ip or a hostname
func getLoadBalancerTarget(lbIngress map[string]any) (string, error) {
	if ip, err := getIPBasedLoadBalancerIP(lbIngress); err == nil {
		return ip.String(), nil
	}

	loadBalancerHostname, found, err := unstructured.NestedString(lbIngress, "hostname")
	if err != nil || !found {
		return "", fmt.Errorf("could not get DNS based load balancer hostname: %w", err)
	}
	return loadBalancerHostname, nil
}

func getIPBasedLoadBalancerIP(lbIngress map[string]any) (net.IP, error) {
	loadBalancerIP, found, err := unstructured.NestedString(lbIngress, "ip")
	if err != nil || !found {
		return nil, fmt.Errorf("could not get IP based load balancer IP: %w", err)
	}

	ip := net.ParseIP(loadBalancerIP)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse IP from load balancer IP %s", loadBalancerIP)
	}

	return ip, nil
}

// WaitUntilDNSReady waits until a random hostname under the given wildcard domain resolves to expectedTarget.
// The expected target may be either an IP address or a hostname.
// It returns the attempt number on which DNS resolved successfully (1-based).
func WaitUntilDNSReady(domain string, expectedTarget string, retryOpts ...retry.Option) (int, error) {
	var attempt uint
	err := retry.Do(func() error {
		attempt++
		ready, err := isDNSReady(domain, expectedTarget)
		if err != nil {
			return fmt.Errorf("error while checking if domain %s resolves to %s: %w", domain, expectedTarget, err)
		}
		if !ready {
			return fmt.Errorf("domain %s is not ready yet for target %s", domain, expectedTarget)
		}
		return nil
	}, retryOpts...)
	if err != nil {
		return 0, err
	}
	return int(attempt), nil
}

func isDNSReady(domain string, expectedTarget string) (bool, error) {
	hostProbe := fmt.Sprintf("probe-%s.%s", randomString(5), domain)
	return doesHostResolveToTarget(hostProbe, expectedTarget)
}

func doesHostResolveToTarget(hostProbe, expectedTarget string) (bool, error) {
	expectedTarget = strings.TrimSpace(expectedTarget)
	if expectedTarget == "" {
		return false, fmt.Errorf("expected target must not be empty")
	}

	if expectedIP := net.ParseIP(expectedTarget); expectedIP != nil {
		return doesHostResolveToIP(hostProbe, expectedIP)
	}

	resolvedCNAME, err := net.LookupCNAME(hostProbe)
	if err == nil && normalizeHostname(resolvedCNAME) == normalizeHostname(expectedTarget) {
		return true, nil
	}

	hostIPs, err := net.LookupIP(hostProbe)
	if err != nil {
		return false, nil
	}
	targetIPs, err := net.LookupIP(expectedTarget)
	if err != nil {
		return false, nil
	}

	for _, hostIP := range hostIPs {
		for _, targetIP := range targetIPs {
			if hostIP.Equal(targetIP) {
				return true, nil
			}
		}
	}

	return false, nil
}

func doesHostResolveToIP(hostProbe string, expectedIP net.IP) (bool, error) {
	ips, err := net.LookupIP(hostProbe)
	if err != nil {
		return false, nil
	}

	for _, ip := range ips {
		if ip.Equal(expectedIP) {
			return true, nil
		}
	}

	return false, nil
}
func normalizeHostname(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func randomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			b[i] = letters[0]
			continue
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
