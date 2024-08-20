package v2alpha1

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Validate gateway", func() {
	It("Should succeed if spec is empty", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{},
		}
		gatewayList := networkingv1beta1.GatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec"))
		Expect(problems[0].Message).To(Equal("Gateway not specified"))
	})

	It("Should fail if gateway does not exist", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To("namespace/gateway"),
			},
		}
		gatewayList := networkingv1beta1.GatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.gateway"))
		Expect(problems[0].Message).To(Equal("Gateway not found"))
	})

	It("Should succeed if gateway exist", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To("namespace/gateway"),
			},
		}
		gatewayList := networkingv1beta1.GatewayList{
			Items: []*networkingv1beta1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway",
						Namespace: "namespace",
					},
				},
			},
		}

		//when
		problems := validateGateway(".spec", gatewayList, apiRule)

		//then
		Expect(problems).To(BeEmpty())
	})
})
