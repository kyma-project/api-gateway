package processing_test

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	v1beta1Status "github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile", func() {
	It("should return api status error and vs/ar status skipped when an error happens during validation", func() {
		// given
		cmd := MockReconciliationCommand{
			validateMock: func() ([]validation.Failure, error) { return nil, fmt.Errorf("error during validation") },
			getStatusBaseMock: func() status.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped)
			},
		}
		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("error during validation"))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())
	})

	It("should return api status error and vs/ar status skipped when validation failed", func() {
		// given
		failures := []validation.Failure{{
			AttributePath: "some.path",
			Message:       "The value is not allowed",
		}}
		cmd := MockReconciliationCommand{
			validateMock: func() ([]validation.Failure, error) { return failures, nil },
			getStatusBaseMock: func() status.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped)
			},
		}
		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("Validation error: Attribute \"some.path\": The value is not allowed"))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})

	It("should return api status error and vs/ar status skipped when processor reconciliation returns error", func() {
		// given
		p := MockReconciliationProcessor{
			evaluate: func() ([]*processing.ObjectChange, error) {
				return []*processing.ObjectChange{}, fmt.Errorf("error during processor execution")
			},
		}

		cmd := MockReconciliationCommand{
			validateMock:   func() ([]validation.Failure, error) { return []validation.Failure{}, nil },
			processorMocks: func() []processing.ReconciliationProcessor { return []processing.ReconciliationProcessor{p} },
			getStatusBaseMock: func() status.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped)
			},
		}

		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("error during processor execution"))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})
	Context("when VirtualService is missing kind", func() {
		It("should return api status error when error happened during apply of changes on VS", func() {
			// given
			c := []*processing.ObjectChange{processing.NewObjectCreateAction(builders.VirtualService().Get())}
			p := MockReconciliationProcessor{
				evaluate: func() ([]*processing.ObjectChange, error) {
					return c, nil
				},
			}

			cmd := MockReconciliationCommand{
				validateMock:   func() ([]validation.Failure, error) { return []validation.Failure{}, nil },
				processorMocks: func() []processing.ReconciliationProcessor { return []processing.ReconciliationProcessor{p} },
				getStatusBaseMock: func() status.ReconciliationStatus {
					return mockStatusBase(gatewayv1beta1.StatusOK)
				},
			}

			client := fake.NewClientBuilder().Build()

			// when
			status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

			// then
			Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(status.ApiRuleStatus.Description).ToNot(BeEmpty())
			Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(status.AuthorizationPolicyStatus).To(BeNil())
			Expect(status.RequestAuthenticationStatus).To(BeNil())

		})
	})

	It("should return status ok for create, update and delete", func() {
		// given
		toBeUpdatedVs := builders.VirtualService().Name("toBeUpdated").Get()
		toBeDeletedVs := builders.VirtualService().Name("toBeDeleted").Get()
		c := []*processing.ObjectChange{
			processing.NewObjectCreateAction(builders.VirtualService().Name("test").Get()),
			processing.NewObjectUpdateAction(toBeUpdatedVs),
			processing.NewObjectDeleteAction(toBeDeletedVs),
		}
		p := MockReconciliationProcessor{
			evaluate: func() ([]*processing.ObjectChange, error) {
				return c, nil
			},
		}

		cmd := MockReconciliationCommand{
			validateMock:   func() ([]validation.Failure, error) { return []validation.Failure{}, nil },
			processorMocks: func() []processing.ReconciliationProcessor { return []processing.ReconciliationProcessor{p} },
			getStatusBaseMock: func() status.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusOK)
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(toBeUpdatedVs, toBeDeletedVs).Build()

		// when
		status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})

	It("should return status error on APIRule and VS for update on non existing VS", func() {
		// give
		toBeUpdatedVs := builders.VirtualService().Name("toBeUpdated").Get()
		toBeUpdatedVs.Kind = "VirtualService"
		c := []*processing.ObjectChange{
			processing.NewObjectUpdateAction(toBeUpdatedVs),
		}
		p := MockReconciliationProcessor{
			evaluate: func() ([]*processing.ObjectChange, error) {
				return c, nil
			},
		}

		cmd := MockReconciliationCommand{
			validateMock:   func() ([]validation.Failure, error) { return []validation.Failure{}, nil },
			processorMocks: func() []processing.ReconciliationProcessor { return []processing.ReconciliationProcessor{p} },
			getStatusBaseMock: func() status.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusOK)
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		// when
		status := processing.Reconcile(context.Background(), client, testLogger(), cmd).(status.ReconciliationV1beta1Status)

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("Error has happened on subresource VirtualService"))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})
})

type MockReconciliationCommand struct {
	validateMock      func() ([]validation.Failure, error)
	getStatusBaseMock func() status.ReconciliationStatus
	processorMocks    func() []processing.ReconciliationProcessor
}

func (r MockReconciliationCommand) Validate(_ context.Context, _ client.Client) ([]validation.Failure, error) {
	return r.validateMock()
}

func (r MockReconciliationCommand) GetProcessors() []processing.ReconciliationProcessor {
	return r.processorMocks()
}

type MockReconciliationProcessor struct {
	evaluate func() ([]*processing.ObjectChange, error)
}

func (r MockReconciliationProcessor) EvaluateReconciliation(_ context.Context, _ client.Client) ([]*processing.ObjectChange, error) {
	return r.evaluate()
}

func (r MockReconciliationCommand) GetStatusBase(string) status.ReconciliationStatus {
	return r.getStatusBaseMock()
}

func testLogger() *logr.Logger {
	logger := ctrl.Log.WithName("test")
	return &logger
}

func mockStatusBase(statusCode gatewayv1beta1.StatusCode) status.ReconciliationStatus {
	return v1beta1Status.ReconciliationV1beta1Status{
		ApiRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		},
		VirtualServiceStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		},
		AccessRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		},
	}
}
