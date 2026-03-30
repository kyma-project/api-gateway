package gateway

import (
	"context"
	"testing"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func createFakeClientForDnsTests(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	require.NoError(t, v1alpha1.AddToScheme(scheme.Scheme))
	require.NoError(t, corev1.AddToScheme(scheme.Scheme))
	require.NoError(t, v1alpha3.AddToScheme(scheme.Scheme))
	require.NoError(t, v1beta1.AddToScheme(scheme.Scheme))
	require.NoError(t, dnsv1alpha1.AddToScheme(scheme.Scheme))
	require.NoError(t, certv1alpha1.AddToScheme(scheme.Scheme))
	require.NoError(t, networkingv1beta1.AddToScheme(scheme.Scheme))
	require.NoError(t, apiextensionsv1.AddToScheme(scheme.Scheme))
	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
}

const (
	disclaimerKey   = "apigateways.operator.kyma-project.io/managed-by-disclaimer"
	disclaimerValue = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."

	gardenerClassKey   = "dns.gardener.cloud/class"
	gardenerClassValue = "garden"
)

func TestReconcileDnsEntry(t *testing.T) {
	tests := []struct {
		name                      string
		targets                   []string
		ipStackType               string
		wantTargets               []string
		wantAdditionalAnnotations map[string]string // additional annotations expected beyond Gardener class and disclaimer
	}{
		{
			name:        "single IPv4 address - no ip-stack annotation",
			targets:     []string{"10.0.0.1"},
			ipStackType: ipStackTypeIPv4,
			wantTargets: []string{"10.0.0.1"},
		},
		{
			name:        "multiple IPv4 addresses - no ip-stack annotation",
			targets:     []string{"10.0.0.1", "10.0.0.2"},
			ipStackType: ipStackTypeIPv4,
			wantTargets: []string{"10.0.0.1", "10.0.0.2"},
		},
		{
			name:        "hostname target (DNS-based LB) - no ip-stack annotation",
			targets:     []string{"some.host.name"},
			ipStackType: ipStackTypeIPv4,
			wantTargets: []string{"some.host.name"},
		},
		{
			name:                      "IPv6 address - ip-stack annotation set to ipv6",
			targets:                   []string{"2001:db8::1"},
			ipStackType:               ipStackTypeIPv6,
			wantTargets:               []string{"2001:db8::1"},
			wantAdditionalAnnotations: map[string]string{ipStackAnnotation: ipStackTypeIPv6},
		},
		{
			name:                      "dual-stack addresses - ip-stack annotation set to dual-stack",
			targets:                   []string{"10.0.0.1", "2001:db8::1"},
			ipStackType:               ipStackTypeDualStack,
			wantTargets:               []string{"10.0.0.1", "2001:db8::1"},
			wantAdditionalAnnotations: map[string]string{ipStackAnnotation: ipStackTypeDualStack},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			k8sClient := createFakeClientForDnsTests(t)

			err := reconcileDnsEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", tc.targets, tc.ipStackType)
			require.NoError(t, err)

			created := dnsv1alpha1.DNSEntry{}
			require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &created))

			assert.Equal(t, "*.test-domain.com", created.Spec.DNSName)
			assert.ElementsMatch(t, tc.wantTargets, created.Spec.Targets)
			assert.Equal(t, "garden", created.Annotations["dns.gardener.cloud/class"])

			if tc.wantAdditionalAnnotations == nil {
				tc.wantAdditionalAnnotations = map[string]string{}
			}
			tc.wantAdditionalAnnotations[disclaimerKey] = disclaimerValue
			tc.wantAdditionalAnnotations[gardenerClassKey] = gardenerClassValue

			assert.Equal(t, tc.wantAdditionalAnnotations, created.Annotations)

			if tc.ipStackType == ipStackTypeIPv4 {
				assert.NotContains(t, created.Annotations, ipStackAnnotation, "IPv4 entries must not carry the ip-stack annotation")
			}
		})
	}
}

func TestReconcileDnsEntry_ReappliesAnnotations(t *testing.T) {
	k8sClient := createFakeClientForDnsTests(t)

	require.NoError(t, reconcileDnsEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", []string{"10.0.0.1"}, ipStackTypeIPv4))

	// Simulate a manual edit that strips all annotations.
	dnsEntry := dnsv1alpha1.DNSEntry{}
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &dnsEntry))
	dnsEntry.Annotations = nil
	require.NoError(t, k8sClient.Update(context.Background(), &dnsEntry))

	// Reconcile again — annotations must be restored.
	require.NoError(t, reconcileDnsEntry(context.Background(), k8sClient, "test", "test-ns", "test-domain.com", []string{"10.0.0.1"}, ipStackTypeIPv4))

	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "test-ns"}, &dnsEntry))
	assert.Equal(t, "garden", dnsEntry.Annotations["dns.gardener.cloud/class"])
}

func TestFetchIstioIngressGatewayIp(t *testing.T) {
	tests := []struct {
		name            string
		service         corev1.Service
		wantAddresses   []string
		wantStackType   string
		wantErrContains string
	}{
		{
			name: "single IPv4 LoadBalancer IP",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{IP: testIstioIngressGatewayLoadBalancerIp}},
					},
				},
			},
			wantAddresses: []string{testIstioIngressGatewayLoadBalancerIp},
			wantStackType: ipStackTypeIPv4,
		},
		{
			name: "multiple IPv4 LoadBalancer IPs",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "172.0.0.1"},
							{IP: "172.0.0.2"},
						},
					},
				},
			},
			wantAddresses: []string{"172.0.0.1", "172.0.0.2"},
			wantStackType: ipStackTypeIPv4,
		},
		{
			name: "IPv6 LoadBalancer IP",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv6Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{IP: "2001:db8::1"}},
					},
				},
			},
			wantAddresses: []string{"2001:db8::1"},
			wantStackType: ipStackTypeIPv6,
		},
		{
			name: "dual-stack LoadBalancer IPs",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "172.0.0.1"},
							{IP: "2001:db8::1"},
						},
					},
				},
			},
			wantAddresses: []string{"172.0.0.1", "2001:db8::1"},
			wantStackType: ipStackTypeDualStack,
		},
		{
			name: "DNS-based hostname LoadBalancer (IPv4)",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{Hostname: "some.host.name"}},
					},
				},
			},
			wantAddresses: []string{"some.host.name"},
			wantStackType: ipStackTypeIPv4,
		},
		{
			name: "ingress entry with both IP and hostname - both appended as targets",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{IP: "172.0.0.1", Hostname: "some.host.name"}},
					},
				},
			},
			wantAddresses: []string{"172.0.0.1", "some.host.name"},
			wantStackType: ipStackTypeIPv4,
		},
		{
			name: "multiple ingress entries each with IP and hostname",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "172.0.0.1", Hostname: "host1.example.com"},
							{IP: "2001:db8::1", Hostname: "host2.example.com"},
						},
					},
				},
			},
			wantAddresses: []string{"172.0.0.1", "host1.example.com", "2001:db8::1", "host2.example.com"},
			wantStackType: ipStackTypeDualStack,
		},
		{
			name: "no ingress entries returns error",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
			},
			wantStackType:   ipStackTypeIPv4,
			wantErrContains: "no ingress exists for",
		},
		{
			name: "ingress entry with neither IP nor hostname returns error",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "istio-ingressgateway", Namespace: "istio-system"},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{}},
					},
				},
			},
			wantStackType:   ipStackTypeIPv4,
			wantErrContains: "no ingress targets found for",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			k8sClient := createFakeClientForDnsTests(t, &tc.service)

			addresses, stackType, err := fetchIstioIngressGatewayIp(context.Background(), k8sClient)

			assert.Equal(t, tc.wantStackType, stackType)

			if tc.wantErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErrContains)
				assert.Nil(t, addresses)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantAddresses, addresses)
			}
		})
	}
}
