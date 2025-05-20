package v2alpha1

import (
	"context"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ExtAuthValidation", func() {
	DescribeTable("validate extAuth", func(cm *corev1.ConfigMap, rule gatewayv2alpha1.Rule, expectedFailures []string, expectedSuccess bool) {
		// given
		ctx := context.Background()

		k8sClient := fake.NewClientBuilder().WithObjects(cm).Build()

		//when
		problems, err := validateExtAuthProviders(ctx, k8sClient, "parentAttributePath", rule)

		//then
		Expect(err == nil).To(Equal(expectedSuccess))
		Expect(problems).To(HaveLen(len(expectedFailures)))
		for i, failure := range problems {
			Expect(failure.Message).To(Equal(expectedFailures[i]))
		}
	},
		Entry("should successfully validate the extAuth if an authorizer exists in Istio config map",
			&corev1.ConfigMap{
				Data: map[string]string{
					"mesh": `
extensionProviders:
- name: "my-authorizer"
  envoyExtAuthzHttp:
    service: "my-authorizer-service"
    port: 8080
`,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio",
					Namespace: "istio-system",
				},
			}, gatewayv2alpha1.Rule{
				ExtAuth: &gatewayv2alpha1.ExtAuth{
					ExternalAuthorizers: []string{
						"my-authorizer",
					},
				},
			}, []string{}, true),
		Entry("should return a validation failure if an authorizer does not exist in Istio config map",
			&corev1.ConfigMap{
				Data: map[string]string{
					"mesh": `
extensionProviders:
- name: "my-other-authorizer"
  envoyExtAuthzHttp:
    service: "my-other-authorizer-service"
    port: 8080
`},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio",
					Namespace: "istio-system",
				},
			}, gatewayv2alpha1.Rule{
				ExtAuth: &gatewayv2alpha1.ExtAuth{
					ExternalAuthorizers: []string{
						"my-authorizer",
					},
				},
			}, []string{"Authorizer not found in Istio ConfigMap mesh data"}, true,
		),
		Entry("should return a validation failure if the Istio ConfigMap does not contain the mesh key",
			&corev1.ConfigMap{
				Data: map[string]string{
					"other-key": "other-value",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio",
					Namespace: "istio-system",
				},
			},
			gatewayv2alpha1.Rule{
				ExtAuth: &gatewayv2alpha1.ExtAuth{
					ExternalAuthorizers: []string{
						"my-authorizer",
					},
				},
			}, []string{"Istio ConfigMap does not contain mesh key"}, true),
		Entry("should return a validation failure if the Istio ConfigMap mesh data cannot be unmarshalled",
			&corev1.ConfigMap{
				Data: map[string]string{
					"mesh": "invalid-yaml",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio",
					Namespace: "istio-system",
				},
			},
			gatewayv2alpha1.Rule{
				ExtAuth: &gatewayv2alpha1.ExtAuth{
					ExternalAuthorizers: []string{
						"my-authorizer",
					},
				},
			}, []string{"Failed to unmarshal mesh data"}, true),
		Entry("should return a validation failure if the Istio ConfigMap cannot be found",
			&corev1.ConfigMap{},
			gatewayv2alpha1.Rule{
				ExtAuth: &gatewayv2alpha1.ExtAuth{
					ExternalAuthorizers: []string{
						"my-authorizer",
					},
				},
			}, []string{"Failed to get Istio ConfigMap"}, true),
	)
})
