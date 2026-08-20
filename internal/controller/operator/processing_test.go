package operator

import (
	"context"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/conditions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("shouldSetProcessing", func() {
	const testNamespace = "test-namespace"
	const apiGatewayCRName = "test-api-gateway"

	Context("When APIGateway CR is in deletion", func() {
		It("should return false", func() {
			now := metav1.Now()
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:              apiGatewayCRName,
					Namespace:         testNamespace,
					DeletionTimestamp: &now,
					Finalizers:        []string{ApiGatewayFinalizer},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})
	})

	Context("When APIGateway CR has no Ready condition (initial install)", func() {
		It("should return true", func() {
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 1,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State: operatorv1alpha1.Ready,
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeTrue())
		})
	})

	Context("When APIGateway CR generation equals ObservedGeneration (periodic reconcile)", func() {
		It("should return false", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 5

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 5,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})
	})

	Context("When APIGateway CR generation is less than ObservedGeneration", func() {
		It("should return false", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 10

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 8,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})
	})

	Context("When APIGateway CR cannot be fetched", func() {
		It("should return false on error", func() {
			c := createFakeClient()
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: "non-existent"})).Should(BeFalse())
		})
	})

	Context("When spec changed (generation increased)", func() {
		enableTrue := true
		enableFalse := false

		It("should return true when enableKymaGateway is being disabled", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":true}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableFalse,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeTrue())
		})

		It("should return true when enableKymaGateway is set to nil after being true", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":true}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: nil,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeTrue())
		})

		It("should return false when enableKymaGateway is being enabled", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":false}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableTrue,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})

		It("should return false when only networkPoliciesEnabled is toggled (no downtime)", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3
			networkTrue := true

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":true}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway:      &enableTrue,
					NetworkPoliciesEnabled: &networkTrue,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})

		It("should return false when enableKymaGateway is false and no annotation exists (was always disabled)", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					// no annotation
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableFalse,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})

		It("should return false when enableKymaGateway is true and no annotation exists (first enable, gateway being created)", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableTrue,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})

		It("should return false when Ready condition exists but ObservedGeneration is zero and spec enables gateway", func() {
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 0

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 1,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":true}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableTrue,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeFalse())
		})
	})

	Context("When Ready condition exists but is not the first condition", func() {
		It("should correctly identify the Ready condition and detect gateway being disabled", func() {
			enableFalse := false
			cond1 := metav1.Condition{
				Type:               "SomeOtherType",
				Status:             metav1.ConditionTrue,
				ObservedGeneration: 2,
				Reason:             "Test",
				Message:            "Test condition",
			}
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 4,
					Annotations: map[string]string{
						operatorv1alpha1.LastAppliedConfigAnnotation: `{"enableKymaGateway":true}`,
					},
				},
				Spec: operatorv1alpha1.APIGatewaySpec{
					EnableKymaGateway: &enableFalse,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{cond1, *readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeTrue())
		})
	})
})
