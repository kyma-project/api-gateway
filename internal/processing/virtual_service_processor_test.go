package processing_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Virtual Service Processor", func() {
	It("should create virtual service when no virtual service exists", func() {
		// given
		apiRule := &gatewayv1beta1.APIRule{}

		processor := processing.VirtualServiceProcessor{
			Creator: mockVirtualServiceCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), GetEmptyFakeClient(), apiRule)

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
					Name: "allow",
				},
			},
		}

		allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
		rules := []gatewayv1beta1.Rule{allowRule}

		apiRule := GetAPIRuleFor(rules)

		vs := networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
				},
			},
		}

		scheme := runtime.NewScheme()
		err := networkingv1beta1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&vs).Build()

		processor := processing.VirtualServiceProcessor{
			Creator: mockVirtualServiceCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
	})
})

type mockVirtualServiceCreator struct {
}

func (r mockVirtualServiceCreator) Create(_ *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return builders.VirtualService().Get()
}
