package networkpolicy_test

import (
	"testing"

	apigatewayv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/networkpolicy"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var testOwnerCR = &apigatewayv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: "test"}}

func TestOperatorPolicy_Handle(t *testing.T) {
	tc := []struct {
		name                   string
		shouldError            bool
		numOfPolicies          int
		networkPoliciesEnabled bool
		objects                []runtime.Object
	}{
		{
			name:                   "should create NetworkPolicies if not present",
			shouldError:            false,
			numOfPolicies:          1,
			networkPoliciesEnabled: true,
		},
		{
			name:                   "should not add NetworkPolicies if they already exist",
			shouldError:            false,
			numOfPolicies:          1,
			networkPoliciesEnabled: true,
			objects: []runtime.Object{
				&networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kyma-project.io--api-gateway-allow",
						Namespace: "kyma-system",
						Labels:    map[string]string{networkpolicy.OwningResourceLabel: testOwnerCR.GetName()},
					},
				},
			},
		},
		{
			name:                   "should finish without updates if policies are disabled",
			numOfPolicies:          0,
			shouldError:            false,
			networkPoliciesEnabled: false,
		},
		{
			name:                   "should delete NetworkPolicies if they exist and provisioning policies is disabled",
			shouldError:            false,
			numOfPolicies:          0,
			networkPoliciesEnabled: false,
			objects: []runtime.Object{
				&networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kyma-project.io--api-gateway-allow",
						Namespace: "kyma-system",
						Labels:    map[string]string{networkpolicy.OwningResourceLabel: testOwnerCR.GetName()},
					},
				},
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewFakeClient(tt.objects...)

			_ = networkingv1.AddToScheme(c.Scheme())
			_ = apigatewayv1alpha1.AddToScheme(c.Scheme())

			r := networkpolicy.OperatorPolicy{
				Client:  c,
				Enabled: tt.networkPoliciesEnabled,
				Owner:   testOwnerCR,
			}
			err := r.Handle(t.Context())
			if err != nil && !tt.shouldError {
				t.Errorf("got unexpected error: %v", err)
			}

			nps := networkingv1.NetworkPolicyList{}
			err = c.List(t.Context(), &nps)
			if err != nil {
				t.Errorf("got unexpected error while listing NetPolicies: %v", err)
			}
			if len(nps.Items) != tt.numOfPolicies {
				t.Errorf("got %d NetPolicies, expected %d", len(nps.Items), tt.numOfPolicies)
			}

			enp := networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kyma-project.io--api-gateway-allow",
					Namespace: "kyma-system",
				},
			}

			assertNetPol(t, c, tt.networkPoliciesEnabled, enp)
		})
	}
}

func TestOperatorPolicy_Handle_SpecUpdate(t *testing.T) {
	np := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kyma-project.io--api-gateway-allow",
			Namespace: "kyma-system",
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{CIDR: "192.168.0.0/16"},
						},
					},
				},
			},
		},
	}
	c := fake.NewFakeClient(&np)

	_ = networkingv1.AddToScheme(c.Scheme())
	_ = apigatewayv1alpha1.AddToScheme(c.Scheme())

	r := networkpolicy.OperatorPolicy{
		Client:  c,
		Enabled: true,
		Owner:   testOwnerCR,
	}
	err := r.Handle(t.Context())
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}

	got := networkingv1.NetworkPolicy{}
	err = c.Get(t.Context(), client.ObjectKeyFromObject(&np), &got)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if len(got.Spec.Ingress) != 2 {
		t.Errorf("got %d Ingress, expected 2", len(got.Spec.Ingress))
	}
}

func assertNetPol(t *testing.T, c client.Client, enabled bool, p networkingv1.NetworkPolicy) {
	err := c.Get(t.Context(), client.ObjectKeyFromObject(&p), &p)
	if err != nil && enabled {
		t.Errorf("got unexpected error while getting resulting NetPolicy: %v", err)
	}

	gotOwner := p.GetLabels()[networkpolicy.OwningResourceLabel]
	if gotOwner != testOwnerCR.GetName() && enabled {
		t.Errorf("unexpected name of owned-resource label: %s", p.GetLabels()[networkpolicy.OwningResourceLabel])
	}
}
