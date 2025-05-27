package helpers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var _ = Describe("FindServiceNamespace", func() {
	It("Finds namespace from rule service if its specified", func() {
		apiRule := gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "apiRuleNS",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: &gatewayv1beta1.Service{
					Namespace: ptr.To("specServiceNS"),
				},
			},
		}
		rule := gatewayv1beta1.Rule{
			Service: &gatewayv1beta1.Service{
				Namespace: ptr.To("ruleNS"),
			},
		}
		serviceNS := FindServiceNamespace(&apiRule, &rule)
		Expect(serviceNS).To(Equal("ruleNS"))
	})

	It("Finds namespace from spec service if its specified and rule does not have it", func() {
		apiRule := gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "apiRuleNS",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: &gatewayv1beta1.Service{
					Namespace: ptr.To("specServiceNS"),
				},
			},
		}
		rule := gatewayv1beta1.Rule{
			Service: &gatewayv1beta1.Service{},
		}
		serviceNS := FindServiceNamespace(&apiRule, &rule)
		Expect(serviceNS).To(Equal("specServiceNS"))
	})

	It("Finds namespace from API Rule directly if nowhere else specified", func() {
		apiRule := gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "apiRuleNS",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: &gatewayv1beta1.Service{},
			},
		}
		rule := gatewayv1beta1.Rule{
			Service: &gatewayv1beta1.Service{},
		}
		serviceNS := FindServiceNamespace(&apiRule, &rule)
		Expect(serviceNS).To(Equal("apiRuleNS"))
	})
})

var _ = Describe("GetLabelSelectorFromService", func() {
	It("Should get workload selector with match labels retrieved from a core service", func() {
		coreService := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceName",
				Namespace: "serviceNS",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"selectorKey": "selectorValue"},
			},
		}
		k8sClient := createFakeClient(&coreService)
		service := gatewayv1beta1.Service{
			Name:      ptr.To("serviceName"),
			Namespace: ptr.To("serviceNS"),
		}
		apiRule := gatewayv1beta1.APIRule{}
		rule := gatewayv1beta1.Rule{}
		workloadSelector, err := GetLabelSelectorFromService(context.Background(), k8sClient, &service, &apiRule, &rule)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadSelector.MatchLabels).To(HaveKeyWithValue("selectorKey", "selectorValue"))
	})

	It("Should return error if service name not specified", func() {
		coreService := corev1.Service{}
		k8sClient := createFakeClient(&coreService)
		service := gatewayv1beta1.Service{
			Namespace: ptr.To("serviceNS"),
		}
		apiRule := gatewayv1beta1.APIRule{}
		rule := gatewayv1beta1.Rule{}
		workloadSelector, err := GetLabelSelectorFromService(context.Background(), k8sClient, &service, &apiRule, &rule)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("service name is required but missing"))
		Expect(workloadSelector.MatchLabels).To(BeNil())
	})

	It("Should get workload selector with match labels retrieved from a core service when no namespace specified in service", func() {
		coreService := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceName",
				Namespace: "serviceNS",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"selectorKey": "selectorValue"},
			},
		}
		k8sClient := createFakeClient(&coreService)
		service := gatewayv1beta1.Service{
			Name: ptr.To("serviceName"),
		}
		apiRule := gatewayv1beta1.APIRule{
			Spec: gatewayv1beta1.APIRuleSpec{
				Service: &gatewayv1beta1.Service{
					Namespace: ptr.To("serviceNS"),
				},
			},
		}
		rule := gatewayv1beta1.Rule{}
		workloadSelector, err := GetLabelSelectorFromService(context.Background(), k8sClient, &service, &apiRule, &rule)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadSelector.MatchLabels).To(HaveKeyWithValue("selectorKey", "selectorValue"))
	})

	It("Should get workload selector with match labels retrieved from a core service when no namespace speficied (use default)", func() {
		coreService := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceName",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"selectorKey": "selectorValue"},
			},
		}
		k8sClient := createFakeClient(&coreService)
		service := gatewayv1beta1.Service{
			Name: ptr.To("serviceName"),
		}
		apiRule := gatewayv1beta1.APIRule{}
		rule := gatewayv1beta1.Rule{}
		workloadSelector, err := GetLabelSelectorFromService(context.Background(), k8sClient, &service, &apiRule, &rule)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadSelector.MatchLabels).To(HaveKeyWithValue("selectorKey", "selectorValue"))
	})

	It("Should return nil workload selector when core service has no selector speficied", func() {
		coreService := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceName",
				Namespace: "serviceNS",
			},
		}
		k8sClient := createFakeClient(&coreService)
		service := gatewayv1beta1.Service{
			Name:      ptr.To("serviceName"),
			Namespace: ptr.To("serviceNS"),
		}
		apiRule := gatewayv1beta1.APIRule{}
		rule := gatewayv1beta1.Rule{}
		workloadSelector, err := GetLabelSelectorFromService(context.Background(), k8sClient, &service, &apiRule, &rule)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadSelector).To(BeNil())
	})
})
