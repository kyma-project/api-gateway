package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	istionetworkingv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testApiRuleName      = "test-apirule"
	testApiRuleNamespace = "default"
	testGatewayNamespace = "kyma-system"
	testGatewayName      = "kyma-gateway"
)

// v2DependencyCRDs are the CRDs dependencies.APIRuleV2().AreAvailable checks for.
var v2DependencyCRDs = []string{
	"virtualservices.networking.istio.io",
	"authorizationpolicies.security.istio.io",
	"requestauthentications.security.istio.io",
}

func conditionsTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		gatewayv1beta1.AddToScheme,
		gatewayv2alpha1.AddToScheme,
		networkingv1beta1.AddToScheme,
		apiextensionsv1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("add to scheme: %v", err)
		}
	}
	return s
}

func testApiRule() *gatewayv2alpha1.APIRule {
	return &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testApiRuleName,
			Namespace:  testApiRuleNamespace,
			Generation: 3,
			Finalizers: []string{apiGatewayFinalizer},
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Hosts:   []*gatewayv2alpha1.Host{ptr.To(gatewayv2alpha1.Host("test.example.com"))},
			Gateway: ptr.To(testGatewayNamespace + "/" + testGatewayName),
			Rules: []gatewayv2alpha1.Rule{
				{
					Path:    "/.*",
					Methods: []gatewayv2alpha1.HttpMethod{"GET"},
					NoAuth:  ptr.To(true),
				},
			},
		},
	}
}

func testGateway() *networkingv1beta1.Gateway {
	return &networkingv1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testGatewayName,
			Namespace: testGatewayNamespace,
		},
		Spec: istionetworkingv1beta1.Gateway{
			Servers: []*istionetworkingv1beta1.Server{
				{
					Hosts: []string{"*.example.com"},
					Port:  &istionetworkingv1beta1.Port{Number: 443, Protocol: "HTTPS", Name: "https"},
				},
			},
		},
	}
}

func newTestReconciler(c client.Client, scheme *runtime.Scheme) *APIRuleReconciler {
	return &APIRuleReconciler{
		Client:                 c,
		APIReader:              c,
		Log:                    logr.Discard(),
		Scheme:                 scheme,
		Config:                 &helpers.Config{},
		ReconciliationConfig:   processing.ReconciliationConfig{},
		OnErrorReconcilePeriod: 0,
		// Metrics is only dereferenced on the v1beta1 path; the v2alpha1 branches
		// under test return before any metrics access, so a nil collector is safe.
		Metrics: nil,
	}
}

func reconcileTestApiRule(t *testing.T, objects []client.Object, crdNames []string) *gatewayv2alpha1.APIRule {
	t.Helper()
	return reconcileTestApiRuleWithClient(t, objects, crdNames, nil)
}

// reconcileTestApiRuleWithClient builds a fake client seeded with the test APIRule,
// the given objects and CRDs, optionally wraps it via wrap (e.g. to inject Get
// errors), runs a single Reconcile and returns the persisted APIRule.
func reconcileTestApiRuleWithClient(t *testing.T, objects []client.Object, crdNames []string, wrap func(client.Client) client.Client) *gatewayv2alpha1.APIRule {
	t.Helper()
	scheme := conditionsTestScheme(t)

	apiRule := testApiRule()
	all := append([]client.Object{apiRule}, objects...)

	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(apiRule).
		WithObjects(all...)
	c := builder.Build()

	for _, name := range crdNames {
		crd := &apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if err := c.Create(context.Background(), crd); err != nil {
			t.Fatalf("create CRD %s: %v", name, err)
		}
	}

	reconcileClient := client.Client(c)
	if wrap != nil {
		reconcileClient = wrap(c)
	}

	r := newTestReconciler(reconcileClient, scheme)
	// Reconcile may return a non-nil (backoff) error when it writes an error/warning
	// status; the status is persisted regardless, which is what we assert on.
	if _, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: testApiRuleNamespace, Name: testApiRuleName},
	}); err != nil {
		t.Logf("Reconcile returned error (expected for error/warning status): %v", err)
	}

	got := &gatewayv2alpha1.APIRule{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: testApiRuleNamespace, Name: testApiRuleName}, got); err != nil {
		t.Fatalf("get reconciled APIRule: %v", err)
	}
	return got
}

func requireCondition(t *testing.T, conditions []metav1.Condition, condType string) metav1.Condition {
	t.Helper()
	c := apimeta.FindStatusCondition(conditions, condType)
	if c == nil {
		t.Fatalf("condition %q not found; have %d conditions", condType, len(conditions))
	}
	return *c
}

// When a required dependency CRD is missing, the dependency branch must surface
// DependenciesReady=False/DependenciesMissing and drive the aggregate Ready to
// False/ReconcileFailed (rather than leaving it stuck at ReconcileProcessing).
func TestReconcileV2Alpha1_DependenciesMissing(t *testing.T) {
	// Gateway present so discovery passes; CRDs seeded except virtualservices.
	got := reconcileTestApiRule(t,
		[]client.Object{testGateway()},
		[]string{
			"authorizationpolicies.security.istio.io",
			"requestauthentications.security.istio.io",
		},
	)

	if got.Status.State != gatewayv2alpha1.Warning {
		t.Errorf("State: got %q, want %q", got.Status.State, gatewayv2alpha1.Warning)
	}

	depCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionFalse {
		t.Errorf("Dependencies status: got %q, want %q", depCond.Status, metav1.ConditionFalse)
	}
	if depCond.Reason != gatewayv2alpha1.ReasonDependenciesMissing {
		t.Errorf("Dependencies reason: got %q, want %q", depCond.Reason, gatewayv2alpha1.ReasonDependenciesMissing)
	}
	if depCond.ObservedGeneration != 3 {
		t.Errorf("Dependencies observedGeneration: got %d, want 3", depCond.ObservedGeneration)
	}

	readyCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want %q", readyCond.Status, metav1.ConditionFalse)
	}
	if readyCond.Reason != gatewayv2alpha1.ReasonReconcileFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, gatewayv2alpha1.ReasonReconcileFailed)
	}
	if readyCond.ObservedGeneration != 3 {
		t.Errorf("Ready observedGeneration: got %d, want 3", readyCond.ObservedGeneration)
	}
}

// When the referenced Gateway does not exist, discovery must surface the
// Ready=False/GatewayNotFound condition and set Error state.
func TestReconcileV2Alpha1_GatewayNotFound(t *testing.T) {
	// No Gateway object; all dependency CRDs present (discovery fails first anyway).
	got := reconcileTestApiRule(t, nil, v2DependencyCRDs)

	if got.Status.State != gatewayv2alpha1.Error {
		t.Errorf("State: got %q, want %q", got.Status.State, gatewayv2alpha1.Error)
	}

	readyCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want %q", readyCond.Status, metav1.ConditionFalse)
	}
	if readyCond.Reason != gatewayv2alpha1.ReasonGatewayNotFound {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, gatewayv2alpha1.ReasonGatewayNotFound)
	}
}

// crdGetErrorClient returns a non-NotFound error on Get of a CustomResourceDefinition,
// simulating a dependency discovery failure that is not a missing CRD.
type crdGetErrorClient struct {
	client.Client
}

func (c crdGetErrorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		return apierrors.NewInternalError(errors.New("boom"))
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

// When dependency discovery fails with a non-NotFound error, the dependency branch
// must surface DependenciesReady=False/DependenciesReconcileFailed, set Error state,
// and drive the aggregate Ready to False/ReconcileFailed.
func TestReconcileV2Alpha1_DependenciesError(t *testing.T) {
	// Gateway present so discovery passes; the wrapped client errors on CRD Get.
	got := reconcileTestApiRuleWithClient(t,
		[]client.Object{testGateway()},
		nil,
		func(c client.Client) client.Client { return crdGetErrorClient{Client: c} },
	)

	if got.Status.State != gatewayv2alpha1.Error {
		t.Errorf("State: got %q, want %q", got.Status.State, gatewayv2alpha1.Error)
	}

	depCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionFalse {
		t.Errorf("Dependencies status: got %q, want %q", depCond.Status, metav1.ConditionFalse)
	}
	if depCond.Reason != gatewayv2alpha1.ReasonDependenciesReconcileFailed {
		t.Errorf("Dependencies reason: got %q, want %q", depCond.Reason, gatewayv2alpha1.ReasonDependenciesReconcileFailed)
	}

	readyCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want %q", readyCond.Status, metav1.ConditionFalse)
	}
	if readyCond.Reason != gatewayv2alpha1.ReasonReconcileFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, gatewayv2alpha1.ReasonReconcileFailed)
	}
}

// When all dependencies are available, the controller must record
// DependenciesReady=True so every downstream path carries it, regardless of how
// sub-resource reconciliation turns out.
func TestReconcileV2Alpha1_DependenciesReady(t *testing.T) {
	// Gateway present and all dependency CRDs installed.
	got := reconcileTestApiRule(t, []client.Object{testGateway()}, v2DependencyCRDs)

	depCond := requireCondition(t, got.Status.Conditions, gatewayv2alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionTrue {
		t.Errorf("Dependencies status: got %q, want %q", depCond.Status, metav1.ConditionTrue)
	}
	if depCond.Reason != gatewayv2alpha1.ReasonDependenciesAvailable {
		t.Errorf("Dependencies reason: got %q, want %q", depCond.Reason, gatewayv2alpha1.ReasonDependenciesAvailable)
	}
	if depCond.ObservedGeneration != 3 {
		t.Errorf("Dependencies observedGeneration: got %d, want 3", depCond.ObservedGeneration)
	}
}
