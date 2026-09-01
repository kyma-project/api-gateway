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

	Context("When APIGateway CR generation is greater than ObservedGeneration (spec changed)", func() {
		It("should return true", func() {
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
					Conditions: []metav1.Condition{*readyCond},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName})).Should(BeTrue())
		})
	})

	Context("When APIGateway CR cannot be fetched", func() {
		It("should return false on error", func() {
			c := createFakeClient()
			agr := &APIGatewayReconciler{Client: c, Scheme: getTestScheme(), log: logr.Discard()}

			Expect(agr.shouldSetProcessing(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: "non-existent"})).Should(BeFalse())
		})
	})
})
