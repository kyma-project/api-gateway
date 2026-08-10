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
			// given
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
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeFalse())
		})
	})

	Context("When APIGateway CR has no Ready condition (initial install)", func() {
		It("should return true", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 1,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State: operatorv1alpha1.Ready,
					// No conditions set yet
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeTrue())
		})
	})

	Context("When APIGateway CR generation equals ObservedGeneration", func() {
		It("should return false (no changes since last successful reconcile)", func() {
			// given
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 5

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 5, // Same as ObservedGeneration
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeFalse())
		})
	})

	Context("When APIGateway CR generation is less than ObservedGeneration", func() {
		It("should return false (no changes)", func() {
			// given
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 10

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 8, // Less than ObservedGeneration
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeFalse())
		})
	})

	Context("When APIGateway CR generation is greater than ObservedGeneration", func() {
		It("should return true (spec has changed since last reconcile)", func() {
			// given
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 3

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 7, // Greater than ObservedGeneration
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeTrue())
		})
	})

	Context("When APIGateway CR cannot be fetched", func() {
		It("should return false on error", func() {
			// given
			c := createFakeClient() // Empty client, so CR won't be found
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: "non-existent"})

			// then
			Expect(result).Should(BeFalse())
		})
	})

	Context("When APIGateway CR has Ready condition with zero ObservedGeneration", func() {
		It("should return true when current generation is greater than zero", func() {
			// given
			readyCond := conditions.ReconcileSucceeded.Condition()
			readyCond.ObservedGeneration = 0

			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Generation: 1,
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then
			Expect(result).Should(BeTrue())
		})
	})

	Context("When Ready condition exists but is not the first condition", func() {
		It("should correctly identify and process the Ready condition", func() {
			// given - create other conditions before Ready
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
				},
				Status: operatorv1alpha1.APIGatewayStatus{
					State:      operatorv1alpha1.Ready,
					Conditions: []metav1.Condition{cond1, *readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result := agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})

			// then - should return true because generation (4) > ObservedGeneration (3)
			Expect(result).Should(BeTrue())
		})
	})
})
