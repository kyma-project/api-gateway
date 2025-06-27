package processing_test

import (
	"context"
	"fmt"
	"net/http"

	apirulev2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/internal/processing"
	testUtils "github.com/kyma-project/api-gateway/internal/processing/processing_test"
)

var _ = Describe("APIRule subresources deletion", func() {

	It("should delete subresources when APIRule is deleted and don't delete other resources", func() {
		// given
		path := "/*"
		methods := []apirulev2alpha1.HttpMethod{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
		host := apirulev2alpha1.Host("myService.myDomain.com")
		namespace := "some-namespace"

		rules := []apirulev2alpha1.Rule{
			{
				Path:    path,
				Methods: methods,
				NoAuth:  ptr.To(true),
			},
		}
		apiRule := apirulev2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-apirule",
				UID:       types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"),
				Namespace: namespace,
			},
			TypeMeta: v1.TypeMeta{
				APIVersion: "gateway.kyma-project.io/v2alpha11",
				Kind:       "APIRule",
			},
			Spec: apirulev2alpha1.APIRuleSpec{
				Hosts:   []*apirulev2alpha1.Host{&host},
				Gateway: ptr.To("some-namespace/some-gateway"),
				Service: &apirulev2alpha1.Service{
					Name: ptr.To("example-service"),
					Port: ptr.To(uint32(8080)),
				},
				Rules: rules,
			},
		}

		apiRuleObjectMeta := v1.ObjectMeta{
			Name:      "test-apirule-psdh34",
			Namespace: namespace,
			Labels: map[string]string{
				"apirule.gateway.kyma-project.io/v1beta1": fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace),
			},
		}
		notApiRuleObjectMeta := v1.ObjectMeta{
			Name:      "test-other-apirule",
			Namespace: namespace,
			Labels: map[string]string{
				"apirule.gateway.kyma-project.io/v1beta1": fmt.Sprintf("%s.%s", "some-other", apiRule.Namespace),
			},
		}

		apiRuleVS := networkingv1beta1.VirtualService{
			ObjectMeta: apiRuleObjectMeta,
		}

		otherVS := networkingv1beta1.VirtualService{
			ObjectMeta: notApiRuleObjectMeta,
		}

		apiRuleRule := rulev1alpha1.Rule{
			ObjectMeta: apiRuleObjectMeta,
		}

		otherRule := rulev1alpha1.Rule{
			ObjectMeta: notApiRuleObjectMeta,
		}

		apiRuleAP := securityv1beta1.AuthorizationPolicy{
			ObjectMeta: apiRuleObjectMeta,
		}

		otherAP := securityv1beta1.AuthorizationPolicy{
			ObjectMeta: notApiRuleObjectMeta,
		}

		apiRuleRA := securityv1beta1.RequestAuthentication{
			ObjectMeta: apiRuleObjectMeta,
		}

		otherRA := securityv1beta1.RequestAuthentication{
			ObjectMeta: notApiRuleObjectMeta,
		}

		client := testUtils.GetFakeClient(&apiRuleVS, &otherVS, &apiRuleRule, &otherRule, &apiRuleAP, &otherAP, &apiRuleRA, &otherRA)

		// when
		err := processing.DeleteAPIRuleSubresources(client, context.Background(), apiRule)
		Expect(err).ShouldNot(HaveOccurred())

		// then
		vsList := networkingv1beta1.VirtualServiceList{}
		err = client.List(context.Background(), &vsList)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(vsList.Items).To(HaveLen(1))
		Expect(vsList.Items[0].Name).To(Equal("test-other-apirule"))

		ruleList := rulev1alpha1.RuleList{}
		err = client.List(context.Background(), &ruleList)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(ruleList.Items).To(HaveLen(1))
		Expect(ruleList.Items[0].Name).To(Equal("test-other-apirule"))

		apList := securityv1beta1.AuthorizationPolicyList{}
		err = client.List(context.Background(), &apList)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(apList.Items).To(HaveLen(1))
		Expect(apList.Items[0].Name).To(Equal("test-other-apirule"))

		raList := securityv1beta1.RequestAuthenticationList{}
		err = client.List(context.Background(), &raList)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(raList.Items).To(HaveLen(1))
		Expect(raList.Items[0].Name).To(Equal("test-other-apirule"))
	})
})
