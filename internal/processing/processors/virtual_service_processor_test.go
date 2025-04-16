package processors_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"time"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	timeout10s gatewayv1beta1.Timeout = 10
	timeout20s gatewayv1beta1.Timeout = 20
)

var _ = Describe("Virtual Service Processor", func() {
	It("should create virtual service when no virtual service exists", func() {
		// given
		processor := processors.VirtualServiceProcessor{
			ApiRule: &gatewayv1beta1.APIRule{},
			Creator: mockVirtualServiceCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), GetFakeClient())

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should update virtual service when virtual service exists", func() {
		// given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoAuth,
				},
			},
		}

		allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
		rules := []gatewayv1beta1.Rule{allowRule}

		apiRule := GetAPIRuleFor(rules)

		vs := networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.OwnerLabel: fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
				},
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&vs).Build()

		processor := processors.VirtualServiceProcessor{
			ApiRule: apiRule,
			Creator: mockVirtualServiceCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
	})
})

var _ = Describe("GetVirtualServiceHttpTimeout", func() {
	It("should return default of 180s when no timeout is set", func() {
		// given
		apiRuleSpec := gatewayv1beta1.APIRuleSpec{}
		rule := gatewayv1beta1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(time.Second * 180))
	})

	It("should return timeout from rule when timeout is set on rule only", func() {
		// given
		apiRuleSpec := gatewayv1beta1.APIRuleSpec{}
		rule := gatewayv1beta1.Rule{
			Timeout: &timeout10s,
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(time.Second * 10))
	})

	It("should return timeout from apiRule when timeout is set on apiRule only", func() {
		// given
		apiRuleSpec := gatewayv1beta1.APIRuleSpec{
			Timeout: &timeout20s,
		}
		rule := gatewayv1beta1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(time.Second * 20))
	})

	It("should return timeout from rule when timeout is set on both apiRule and rule", func() {
		// given
		apiRuleSpec := gatewayv1beta1.APIRuleSpec{
			Timeout: &timeout20s,
		}
		rule := gatewayv1beta1.Rule{
			Timeout: &timeout10s,
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(time.Second * 10))
	})

})

type mockVirtualServiceCreator struct {
}

func (r mockVirtualServiceCreator) Create(_ *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return builders.VirtualService().Get(), nil
}
