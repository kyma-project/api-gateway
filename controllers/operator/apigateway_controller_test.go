package operator

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API-Gateway Controller", func() {
	Context("Reconcile", func() {
		It("Should not return an error when CR was not found", func() {
			// given
			c := createFakeClient()
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should add finalizer when API-Gateway CR is not marked for deletion", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})

		It("Should set status to Ready when API-Gateway reconciliation succeed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Finalizers: []string{ApiGatewayFinalizer},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).Should(Equal(operatorv1alpha1.Ready))
		})

		It("Should delete API-Gateway CR if there are no blocking resources", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
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
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(errors.IsNotFound(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR))).To(BeTrue())
		})

		It("Should not delete API-Gateway CR if there are any APIRules on cluster", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
				Name:              apiGatewayCRName,
				Namespace:         testNamespace,
				DeletionTimestamp: &now,
				Finalizers:        []string{ApiGatewayFinalizer},
			},
			}
			apiRule := &gatewayv1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-rule",
					Namespace: "default",
				},
			}

			c := createFakeClient(apiGatewayCR, apiRule)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("could not delete API-Gateway CR since there are custom resources that block its deletion"))
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are custom resources that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})

		It("Should not delete API-Gateway CR if there is any ORY Oathkeeper Rules on cluster", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
				Name:              apiGatewayCRName,
				Namespace:         testNamespace,
				DeletionTimestamp: &now,
				Finalizers:        []string{ApiGatewayFinalizer},
			},
			}
			oryRule := &oryv1alpha1.Rule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ory-rule",
					Namespace: "default",
				},
			}

			c := createFakeClient(apiGatewayCR, oryRule)
			agr := &APIGatewayReconciler{
				Client: c,
				Scheme: getTestScheme(),
				log:    logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("could not delete API-Gateway CR since there are custom resources that block its deletion"))
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are custom resources that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})
	})
})
