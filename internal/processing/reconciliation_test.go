package processing_test

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	istioHandler "github.com/kyma-incubator/api-gateway/internal/processing/istio"
	oryHandler "github.com/kyma-incubator/api-gateway/internal/processing/ory"
	"github.com/kyma-incubator/api-gateway/internal/validation"
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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped, ory)
			},
		}
		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped, ory)
			},
		}
		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusSkipped, ory)
			},
		}

		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("error during processor execution"))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})

	It("should return api status error and vs/ar status error when error during apply of changes", func() {
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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusOK, ory)
			},
		}

		client := fake.NewClientBuilder().Build()

		// when
		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).ToNot(BeEmpty())
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusOK, ory)
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(toBeUpdatedVs, toBeDeletedVs).Build()

		// when
		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})

	It("should return status error for update on non existing VS", func() {
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
			getStatusBaseMock: func() processing.ReconciliationStatus {
				return mockStatusBase(gatewayv1beta1.StatusOK, ory)
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		// when
 		status := processing.Reconcile(context.TODO(), client, testLogger(), cmd, &gatewayv1beta1.APIRule{})

		// then
		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())

	})
})

type MockReconciliationCommand struct {
	validateMock      func() ([]validation.Failure, error)
	getStatusBaseMock func() processing.ReconciliationStatus
	processorMocks    func() []processing.ReconciliationProcessor
}

func (r MockReconciliationCommand) Validate(_ context.Context, _ client.Client, _ *gatewayv1beta1.APIRule) ([]validation.Failure, error) {
	return r.validateMock()
}

func (r MockReconciliationCommand) GetProcessors() []processing.ReconciliationProcessor {
	return r.processorMocks()
}

type MockReconciliationProcessor struct {
	evaluate func() ([]*processing.ObjectChange, error)
}

func (r MockReconciliationProcessor) EvaluateReconciliation(_ context.Context, _ client.Client, _ *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, error) {
	return r.evaluate()
}

func (c MockReconciliationCommand) GetStatusBase(_ gatewayv1beta1.StatusCode) processing.ReconciliationStatus {
	return c.getStatusBaseMock()
}

func testLogger() *logr.Logger {
	logger := ctrl.Log.WithName("test")
	return &logger
}

type handler int

const (
	istio handler = iota
	ory
)

func mockStatusBase(statusCode gatewayv1beta1.StatusCode, handler handler) processing.ReconciliationStatus {
	if handler == istio {
		return istioHandler.IstioStatusBase(statusCode)
	}
	return oryHandler.OryStatusBase(statusCode)
}
