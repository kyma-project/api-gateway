package ingress

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

// lookupIP is the resolver seam: net.LookupIP in production, overridden in unit tests.
var lookupIP = net.LookupIP

// LoadBalancer describes the istio ingress load balancer: the DNS targets it advertises
// (each status.loadBalancer.ingress[*] as IP if present, else hostname) and the IP
// families it serves (spec.ipFamilies). On an IP-based dual-stack LB, Targets holds one
// IP per family; on a hostname-based LB (e.g. AWS NLB) it holds the single hostname that
// itself carries A and AAAA.
type LoadBalancer struct {
	Targets    []string
	IPFamilies []string
}

// DualStack reports whether the ingress advertises both address families.
func (lb LoadBalancer) DualStack() bool { return len(lb.IPFamilies) == 2 }

// GetLoadBalancer reads the ingress Service once and returns its targets and IP families.
func GetLoadBalancer(ctx context.Context, r *resources.Resources, svcName, svcNamespace string) (LoadBalancer, error) {
	svc := &unstructured.Unstructured{}
	svc.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"})
	if err := r.Get(ctx, svcName, svcNamespace, svc); err != nil {
		return LoadBalancer{}, fmt.Errorf("service %s/%s was not found: %w", svcNamespace, svcName, err)
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil {
		return LoadBalancer{}, fmt.Errorf("could not get load balancer ingress from service %s/%s: %w", svcNamespace, svcName, err)
	}
	if !found {
		return LoadBalancer{}, fmt.Errorf("load balancer ingress not found in service %s/%s", svcNamespace, svcName)
	}

	targets, err := loadBalancerTargets(ingress)
	if err != nil {
		return LoadBalancer{}, fmt.Errorf("service %s/%s: %w", svcNamespace, svcName, err)
	}
	families, err := ipFamiliesFromSpec(svc.Object)
	if err != nil {
		return LoadBalancer{}, fmt.Errorf("service %s/%s: %w", svcNamespace, svcName, err)
	}
	return LoadBalancer{Targets: targets, IPFamilies: families}, nil
}

func loadBalancerTargets(ingress []any) ([]string, error) {
	if len(ingress) == 0 {
		return nil, fmt.Errorf("load balancer ingress is empty")
	}
	targets := make([]string, 0, len(ingress))
	for i, entry := range ingress {
		lbIngress, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("could not parse load balancer ingress object at index %d", i)
		}
		target, err := getLoadBalancerTarget(lbIngress)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func getLoadBalancerTarget(lbIngress map[string]any) (string, error) {
	if ip, found, _ := unstructured.NestedString(lbIngress, "ip"); found && net.ParseIP(ip) != nil {
		return ip, nil
	}
	hostname, found, err := unstructured.NestedString(lbIngress, "hostname")
	if err != nil {
		return "", fmt.Errorf("could not get DNS based load balancer hostname: %w", err)
	}
	if !found {
		return "", fmt.Errorf("neither ip nor hostname found in load balancer ingress object")
	}
	return hostname, nil
}

// ipFamiliesFromSpec returns Service spec.ipFamilies, defaulting to ["IPv4"] when unset,
// mirroring the operator's stack-type derivation in internal/reconciliations/gateway/dns_entry.go.
func ipFamiliesFromSpec(svcObj map[string]any) ([]string, error) {
	families, found, err := unstructured.NestedStringSlice(svcObj, "spec", "ipFamilies")
	if err != nil {
		return nil, fmt.Errorf("could not get spec.ipFamilies: %w", err)
	}
	if !found || len(families) == 0 {
		return []string{"IPv4"}, nil
	}
	return families, nil
}

// WaitUntilDNSReady waits until a random probe host under the wildcard domain resolves to
// the load balancer targets. When requireBoth is set (dual-stack ingress) the probe must
// resolve to at least one IPv4 AND one IPv6 address among the targets; otherwise a single
// matching family suffices. Returns the 1-based attempt on which it became ready.
func WaitUntilDNSReady(domain string, targets []string, requireBoth bool, retryOpts ...retry.Option) (int, error) {
	var attempt uint
	err := retry.Do(func() error {
		attempt++
		v4, v6, err := probeMatch(domain, targets)
		if err != nil {
			return fmt.Errorf("checking DNS readiness for %s -> %v: %w", domain, targets, err)
		}
		if requireBoth && (!v4 || !v6) {
			return fmt.Errorf("domain %s does not yet resolve over both IPv4 and IPv6 for targets %v", domain, targets)
		}
		if !requireBoth && (!v4 && !v6) {
			return fmt.Errorf("domain %s does not yet resolve to targets %v", domain, targets)
		}
		return nil
	}, retryOpts...)
	if err != nil {
		return 0, err
	}
	return int(attempt), nil
}

// probeMatch resolves a fresh probe host under domain and reports which IP families of the
// probe's addresses also appear among the targets' addresses (IP targets compared directly,
// hostname targets resolved).
func probeMatch(domain string, targets []string) (v4, v6 bool, err error) {
	if len(targets) == 0 {
		return false, false, fmt.Errorf("no targets provided")
	}
	hostProbe := fmt.Sprintf("probe-%s.%s", randomString(5), domain)
	probeIPs, err := lookupIP(hostProbe)
	if err != nil {
		return false, false, fmt.Errorf("DNS lookup for %s failed: %w", hostProbe, err)
	}
	expected, err := resolveTargets(targets)
	if err != nil {
		return false, false, err
	}
	for _, p := range probeIPs {
		for _, e := range expected {
			if !p.Equal(e) {
				continue
			}
			if p.To4() != nil {
				v4 = true
			} else {
				v6 = true
			}
		}
	}
	return v4, v6, nil
}

// resolveTargets flattens targets into the IPs they point at: IP targets are parsed
// directly, hostname targets are resolved.
func resolveTargets(targets []string) ([]net.IP, error) {
	var ips []net.IP
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if ip := net.ParseIP(target); ip != nil {
			ips = append(ips, ip)
			continue
		}
		resolved, err := lookupIP(target)
		if err != nil {
			return nil, fmt.Errorf("DNS lookup for target %s failed: %w", target, err)
		}
		ips = append(ips, resolved...)
	}
	return ips, nil
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
