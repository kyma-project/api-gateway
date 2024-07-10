package v2alpha1

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

var _ = Describe("Validate", func() {

	sampleServiceName := "some-service"
	host := v2alpha1.Host(sampleServiceName + ".test.dev")

	It("should invoke rules validation", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules:   nil,
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := (&APIRuleValidator{ApiRule: apiRule}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{})

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

})
