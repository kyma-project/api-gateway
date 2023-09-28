package operator

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/described_errors"
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
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &StatusMock{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should call update status to processing when CR is not deleted", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
				},
			}

			statusMock := StatusMock{}
			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &statusMock,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(statusMock.updatedToProcessingCalled).Should(BeTrue())
		})

		It("Should return an error when update status to processing failed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
				},
			}

			statusMock := StatusMock{
				processingError: errors.New("Update to processing error"),
			}
			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &statusMock,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("Update to processing error"))
			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(statusMock.updatedToProcessingCalled).Should(BeTrue())
		})

		It("Should call update status to deleting when CR is deleted", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
					Finalizers: []string{"apigateways.operator.kyma-project.io/test-mock"},
				},
			}

			statusMock := StatusMock{}
			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &statusMock,
			}

			// when
			_, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(statusMock.updatedToDeletingCalled).Should(BeTrue())
		})

		It("Should return an error when update status to deleting failed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
					Finalizers: []string{"apigateways.operator.kyma-project.io/test-mock"},
				},
			}

			statusMock := StatusMock{
				deletingError: errors.New("Update to deleting error"),
			}
			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &statusMock,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("Update to deleting error"))
			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(statusMock.updatedToDeletingCalled).Should(BeTrue())
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
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &StatusMock{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should return an error when update status to ready failed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					Finalizers: []string{
						"apigateways.operator.kyma-project.io/api-gateway-reconciliation",
					},
				},
			}
			statusMock := StatusMock{
				readyError: errors.New("Update to ready error"),
			}
			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            &statusMock,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("Update to ready error"))
			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(statusMock.updatedToReadyCalled).Should(BeTrue())
		})

		It("Should set status to ready when API-Gateway reconciliation succeed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					Finalizers: []string{
						"apigateways.operator.kyma-project.io/api-gateway-reconciliation",
					},
				},
			}

			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client:                   apiClient,
				Scheme:                   getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{},
				log:                      logr.Discard(),
				statusHandler:            newStatusHandler(apiClient),
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
						"apigateways.operator.kyma-project.io/api-gateway-reconciliation",
					},
				},
			}

			apiClient := createFakeClient(apiGatewayCR)

			agr := &APIGatewayReconciler{
				Client: apiClient,
				Scheme: getTestScheme(),
				apiGatewayReconciliation: &apiGatewayReconciliationMock{
					err: described_errors.NewDescribedError(errors.New("API-Gateway test error"), "Test error description"),
				},
				log:           logr.Discard(),
				statusHandler: newStatusHandler(apiClient),
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
	err described_errors.DescribedError
}

func (i *apiGatewayReconciliationMock) Reconcile(_ context.Context, apiGatewayCR operatorv1alpha1.APIGateway, _ string) (operatorv1alpha1.APIGateway, described_errors.DescribedError) {
	return apiGatewayCR, i.err
}

type StatusMock struct {
	processingError           error
	updatedToProcessingCalled bool
	readyError                error
	updatedToReadyCalled      bool
	deletingError             error
	updatedToDeletingCalled   bool
	errorError                error
	updatedToErrorCalled      bool
}

func (s *StatusMock) updateToProcessing(_ context.Context, _ string, _ *operatorv1alpha1.APIGateway) error {
	s.updatedToProcessingCalled = true
	return s.processingError
}

func (s *StatusMock) updateToError(_ context.Context, _ described_errors.DescribedError, _ *operatorv1alpha1.APIGateway) error {
	s.updatedToErrorCalled = true
	return s.errorError
}

func (s *StatusMock) updateToDeleting(_ context.Context, _ *operatorv1alpha1.APIGateway) error {
	s.updatedToDeletingCalled = true
	return s.deletingError
}

func (s *StatusMock) updateToReady(_ context.Context, _ *operatorv1alpha1.APIGateway) error {
	s.updatedToReadyCalled = true
	return s.readyError
}
