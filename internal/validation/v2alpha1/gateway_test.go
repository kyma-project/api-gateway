package v2alpha1

import (
	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Validate gateway", func() {
	It("Should fail if neither gateway nor externalGateway is specified", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{},
		}
		gatewayList := networkingv1beta1.GatewayList{}
		externalGwList := externalv1alpha1.ExternalGatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, externalGwList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec"))
		Expect(problems[0].Message).To(Equal("Either gateway or externalGateway must be specified"))
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
		externalGwList := externalv1alpha1.ExternalGatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, externalGwList, apiRule)

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
		externalGwList := externalv1alpha1.ExternalGatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, externalGwList, apiRule)

		//then
		Expect(problems).To(BeEmpty())
	})

	It("Should succeed if externalGateway exists", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				ExternalGateway: ptr.To("namespace/external-gateway"),
			},
		}
		gatewayList := networkingv1beta1.GatewayList{}
		externalGwList := externalv1alpha1.ExternalGatewayList{
			Items: []externalv1alpha1.ExternalGateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-gateway",
						Namespace: "namespace",
					},
				},
			},
		}

		//when
		problems := validateGateway(".spec", gatewayList, externalGwList, apiRule)

		//then
		Expect(problems).To(BeEmpty())
	})

	It("Should fail if externalGateway does not exist", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				ExternalGateway: ptr.To("namespace/external-gateway"),
			},
		}
		gatewayList := networkingv1beta1.GatewayList{}
		externalGwList := externalv1alpha1.ExternalGatewayList{}

		//when
		problems := validateGateway(".spec", gatewayList, externalGwList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.externalGateway"))
		Expect(problems[0].Message).To(Equal("ExternalGateway not found"))
	})
})
