package operator

import (
	"context"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/operator/reconciliations/api_gateway"
	"github.com/pkg/errors"
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
			apiClient := createFakeClient()

			agr := &APIGatewayReconciler{
				Client: apiClient,
				Scheme: getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{
					status: controllers.ReadyStatus(),
				},
				log: logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should not requeue a CR without finalizers, because it's considered to be in deletion", func() {
			// given
			istioCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
				},
			}

			apiClient := createFakeClient(istioCR)

			agr := &APIGatewayReconciler{
				Client: apiClient,
				Scheme: getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{
					status: controllers.ReadyStatus(),
				},
				log: logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should set status to ready when API-Gateway reconciliation succeed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					Finalizers: []string{
						api_gateway.ApiGatewayFinalizer,
					},
				},
			}

			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client: apiClient,
				Scheme: getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{
					status: controllers.ReadyStatus(),
				},
				log: logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(apiClient.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).Should(Equal(operatorv1alpha1.Ready))
		})

		It("Should set error status and return an error when API-Gateway reconciliation failed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					Finalizers: []string{
						api_gateway.ApiGatewayFinalizer,
					},
				},
			}

			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client: apiClient,
				Scheme: getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{
					status: controllers.ErrorStatus(errors.New("API-Gateway test error"), "Test error description"),
				},
				log: logr.Discard(),
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("API-Gateway test error"))
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(apiClient.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).Should(Equal(operatorv1alpha1.Error))
			Expect(apiGatewayCR.Status.Description).To(ContainSubstring("Test error description"))
		})
	})
})

type apiGatewayReconciliationMock struct {
	status controllers.Status
}

func (i *apiGatewayReconciliationMock) Reconcile(_ context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) controllers.Status {
	return i.status
}
