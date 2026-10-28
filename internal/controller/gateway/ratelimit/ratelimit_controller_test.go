package ratelimit_test

import (
	"context"
	"errors"
	"testing"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/controller/gateway/ratelimit"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"time"
)

const (
	testRateLimitName      = "test-ratelimit"
	testRateLimitNamespace = "default"
	rateLimitDependencyCRD = "envoyfilters.networking.istio.io"
)

func rateLimitTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		corev1.AddToScheme,
		ratelimitv1alpha1.AddToScheme,
		operatorv1alpha1.AddToScheme,
		networkingv1alpha3.AddToScheme,
		apiextensionsv1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("add to scheme: %v", err)
		}
	}
	return s
}

func testRateLimit() *ratelimitv1alpha1.RateLimit {
	return &ratelimitv1alpha1.RateLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testRateLimitName,
			Namespace:  testRateLimitNamespace,
			Generation: 3,
		},
		Spec: ratelimitv1alpha1.RateLimitSpec{
			SelectorLabels: map[string]string{"app": "test"},
			Local: ratelimitv1alpha1.LocalConfig{
				DefaultBucket: ratelimitv1alpha1.BucketSpec{
					MaxTokens:     10,
					TokensPerFill: 10,
					FillInterval:  &metav1.Duration{Duration: time.Minute},
				},
			},
		},
	}
}

func testRateLimitAPIGateway() *operatorv1alpha1.APIGateway {
	ag := &operatorv1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-apigateway",
		},
	}
	ag.Status.State = operatorv1alpha1.Ready
	return ag
}

// testRateLimitPod is a pod with the istio sidecar annotation and the selector labels
// used by testRateLimit, so that Validate passes in fake-client tests.
func testRateLimitPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: testRateLimitNamespace,
			Annotations: map[string]string{
				"sidecar.istio.io/status": "",
			},
			Labels: map[string]string{"app": "test"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c", Image: "busybox"}},
		},
	}
}

// reconcileTestRateLimit seeds a fake client with the test RateLimit, the given extra
// objects and CRDs, runs a single Reconcile, and returns the persisted RateLimit.
func reconcileTestRateLimit(t *testing.T, objects []client.Object, crdNames []string) *ratelimitv1alpha1.RateLimit {
	t.Helper()
	return reconcileTestRateLimitWithClient(t, objects, crdNames, nil)
}

func reconcileTestRateLimitWithClient(t *testing.T, objects []client.Object, crdNames []string, wrap func(client.Client) client.Client) *ratelimitv1alpha1.RateLimit {
	t.Helper()
	scheme := rateLimitTestScheme(t)

	rl := testRateLimit()
	all := append([]client.Object{rl}, objects...)

	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(rl).
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

	rec := &ratelimit.RateLimitReconciler{
		Client:          reconcileClient,
		Scheme:          scheme,
		ReconcilePeriod: 0,
	}

	if _, err := rec.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: testRateLimitNamespace, Name: testRateLimitName},
	}); err != nil {
		t.Logf("Reconcile returned error (expected for error status paths): %v", err)
	}

	got := &ratelimitv1alpha1.RateLimit{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: testRateLimitNamespace, Name: testRateLimitName}, got); err != nil {
		t.Fatalf("get reconciled RateLimit: %v", err)
	}
	return got
}

func requireRLCondition(t *testing.T, conditions []metav1.Condition, condType string) metav1.Condition {
	t.Helper()
	c := apimeta.FindStatusCondition(conditions, condType)
	if c == nil {
		t.Fatalf("condition %q not found; have %d conditions", condType, len(conditions))
	}
	return *c
}

// A missing dependency CRD must surface DependenciesReady=False and drive Ready to False.
func TestReconcileRateLimit_DependenciesMissing(t *testing.T) {
	got := reconcileTestRateLimit(t, nil, nil) // no CRDs seeded

	depCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionFalse {
		t.Errorf("DependenciesReady status: got %q, want False", depCond.Status)
	}
	if depCond.Reason != ratelimitv1alpha1.ReasonDependenciesMissing {
		t.Errorf("DependenciesReady reason: got %q, want %q", depCond.Reason, ratelimitv1alpha1.ReasonDependenciesMissing)
	}
	if depCond.ObservedGeneration != 3 {
		t.Errorf("DependenciesReady observedGeneration: got %d, want 3", depCond.ObservedGeneration)
	}

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want False", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonReconcileFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonReconcileFailed)
	}
}

// A non-NotFound error on the CRD check must surface DependenciesReady=False/DependenciesReconcileFailed.
func TestReconcileRateLimit_DependenciesError(t *testing.T) {
	// Wrap the client so that Get on a CRD returns a synthetic server error.
	wrap := func(inner client.Client) client.Client {
		return &crdGetErrorClient{Client: inner}
	}
	got := reconcileTestRateLimitWithClient(t, nil, nil, wrap)

	depCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionFalse {
		t.Errorf("DependenciesReady status: got %q, want False", depCond.Status)
	}
	if depCond.Reason != ratelimitv1alpha1.ReasonDependenciesReconcileFailed {
		t.Errorf("DependenciesReady reason: got %q, want %q", depCond.Reason, ratelimitv1alpha1.ReasonDependenciesReconcileFailed)
	}

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want False", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonReconcileFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonReconcileFailed)
	}
}

// No APIGateway CR in the cluster must surface DependenciesReady=True and Ready=False/APIGatewayNotReady.
func TestReconcileRateLimit_APIGatewayNotReady(t *testing.T) {
	got := reconcileTestRateLimit(t,
		nil, // no APIGateway seeded
		[]string{rateLimitDependencyCRD},
	)

	depCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionTrue {
		t.Errorf("DependenciesReady status: got %q, want True", depCond.Status)
	}

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want False", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonAPIGatewayNotReady {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonAPIGatewayNotReady)
	}
}

// Validation failure (no pods matching selector) must surface Ready=False/ValidationFailed.
func TestReconcileRateLimit_ValidationFailed(t *testing.T) {
	// CRD and APIGateway present, but no pod seeded — Validate rejects the spec.
	got := reconcileTestRateLimit(t,
		[]client.Object{testRateLimitAPIGateway()},
		[]string{rateLimitDependencyCRD},
	)

	depCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionTrue {
		t.Errorf("DependenciesReady status: got %q, want True", depCond.Status)
	}

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want False", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonValidationFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonValidationFailed)
	}
}

// EnvoyFilter create/update failure must surface Ready=False/EnvoyFilterReconcileFailed.
func TestReconcileRateLimit_EnvoyFilterFailed(t *testing.T) {
	wrap := func(inner client.Client) client.Client {
		return &envoyFilterCreateErrorClient{Client: inner}
	}
	got := reconcileTestRateLimitWithClient(t,
		[]client.Object{testRateLimitAPIGateway(), testRateLimitPod()},
		[]string{rateLimitDependencyCRD},
		wrap,
	)

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready status: got %q, want False", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonEnvoyFilterReconcileFailed {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonEnvoyFilterReconcileFailed)
	}
}

// Full happy path: EnvoyFilter is created successfully → Ready=True/ReconcileSucceeded.
func TestReconcileRateLimit_EnvoyFilterReconciled(t *testing.T) {
	got := reconcileTestRateLimit(t,
		[]client.Object{testRateLimitAPIGateway(), testRateLimitPod()},
		[]string{rateLimitDependencyCRD},
	)

	depCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeDependencies)
	if depCond.Status != metav1.ConditionTrue {
		t.Errorf("DependenciesReady status: got %q, want True", depCond.Status)
	}

	readyCond := requireRLCondition(t, got.Status.Conditions, ratelimitv1alpha1.ConditionTypeReady)
	if readyCond.Status != metav1.ConditionTrue {
		t.Errorf("Ready status: got %q, want True", readyCond.Status)
	}
	if readyCond.Reason != ratelimitv1alpha1.ReasonReconcileSucceeded {
		t.Errorf("Ready reason: got %q, want %q", readyCond.Reason, ratelimitv1alpha1.ReasonReconcileSucceeded)
	}
	if readyCond.ObservedGeneration != 3 {
		t.Errorf("Ready observedGeneration: got %d, want 3", readyCond.ObservedGeneration)
	}
}

// crdGetErrorClient wraps a client and returns a synthetic internal-server-error
// whenever Get is called for a CustomResourceDefinition object.
type crdGetErrorClient struct {
	client.Client
}

func (c *crdGetErrorClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		return apierrors.NewInternalError(errors.New("synthetic CRD lookup error"))
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

var _ client.Client = &crdGetErrorClient{}

// envoyFilterCreateErrorClient wraps a client and returns a synthetic error
// whenever Create or Update is called for an EnvoyFilter object.
type envoyFilterCreateErrorClient struct {
	client.Client
}

func (c *envoyFilterCreateErrorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if _, ok := obj.(*networkingv1alpha3.EnvoyFilter); ok {
		return apierrors.NewInternalError(errors.New("synthetic EnvoyFilter create error"))
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *envoyFilterCreateErrorClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if _, ok := obj.(*networkingv1alpha3.EnvoyFilter); ok {
		return apierrors.NewInternalError(errors.New("synthetic EnvoyFilter update error"))
	}
	return c.Client.Update(ctx, obj, opts...)
}

var _ client.Client = &envoyFilterCreateErrorClient{}
