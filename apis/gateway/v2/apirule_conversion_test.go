package v2_test

import (
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("APIRule Conversion", func() {
	host1 := v2.Host("host1")

	Describe("v2 to v2alpha1", func() {
		apiName := "test-apirule"
		namespace := "test-namespace"
		apiVersion := "gateway.kyma-project.io/v2"
		apiKind := "APIRule"

		apiRuleV2 := &v2.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      apiName,
				Namespace: namespace,
			},
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersion,
				Kind:       apiKind,
			},
			Spec:   v2.APIRuleSpec{},
			Status: v2.APIRuleStatus{},
		}

		It("should convert status", func() {
			// given
			apiRuleV2.Status = v2.APIRuleStatus{
				State:       v2.Ready,
				Description: "Reconciled successfully",
			}
			apiRuleV2alpha1 := &v2alpha1.APIRule{}

			// when
			err := apiRuleV2.ConvertTo(apiRuleV2alpha1)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Status.State).To(Equal(v2alpha1.Ready))
			Expect(apiRuleV2alpha1.Status.Description).To(Equal("Reconciled successfully"))
		})
		It("should convert spec without changes", func() {
			// given
			apiRuleV2.Spec = v2.APIRuleSpec{
				Hosts: []*v2.Host{&host1},
				Service: &v2.Service{
					Name:      ptr.To("test-service"),
					Namespace: &namespace,
					Port:      ptr.To(uint32(80)),
				},
				Rules: []v2.Rule{
					{
						Path:    "/test",
						Methods: []v2.HttpMethod{"GET", "POST"},
						Service: &v2.Service{
							Name:      ptr.To("test-service-rule-level"),
							Namespace: &namespace,
							Port:      ptr.To(uint32(80)),
						},
						NoAuth: ptr.To(true),
					},
					{
						Path:    "/*",
						Methods: []v2.HttpMethod{"DELETE"},
						Jwt: &v2.JwtConfig{
							Authentications: []*v2.JwtAuthentication{
								{
									Issuer:  "https://issuer.com",
									JwksUri: "https://issuer.com/certs",
								},
							},
							Authorizations: []*v2.JwtAuthorization{
								{
									RequiredScopes: []string{"scope1", "scope2"},
									Audiences:      []string{"audience1", "audience2"},
								},
							},
						},
					},
				},
			}
			// when
			apiRuleV2alpha1 := &v2alpha1.APIRule{}

			err := apiRuleV2.ConvertTo(apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())

			v2alpha1Host := v2alpha1.Host("host1")
			expectedSpec := v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{&v2alpha1Host},
				Service: &v2alpha1.Service{
					Name:      ptr.To("test-service"),
					Namespace: &namespace,
					Port:      ptr.To(uint32(80)),
				},
				Rules: []v2alpha1.Rule{
					{
						Path:    "/test",
						Methods: []v2alpha1.HttpMethod{"GET", "POST"},
						Service: &v2alpha1.Service{
							Name:      ptr.To("test-service-rule-level"),
							Namespace: &namespace,
							Port:      ptr.To(uint32(80)),
						},
						NoAuth: ptr.To(true),
					},
					{
						Path:    "/*",
						Methods: []v2alpha1.HttpMethod{"DELETE"},
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "https://issuer.com",
									JwksUri: "https://issuer.com/certs",
								},
							},
							Authorizations: []*v2alpha1.JwtAuthorization{
								{
									RequiredScopes: []string{"scope1", "scope2"},
									Audiences:      []string{"audience1", "audience2"},
								},
							},
						},
					},
				},
			}
			Expect(apiRuleV2alpha1.Spec).To(Equal(expectedSpec))

		})
	})
	Describe("v2alpha1 to v2", func() {
		apiName := "test-apirule"
		namespace := "test-namespace"
		apiVersion := "gateway.kyma-project.io/v2alpha1"
		apiKind := "APIRule"

		apiRuleV2alpha1 := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      apiName,
				Namespace: namespace,
			},
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersion,
				Kind:       apiKind,
			},
			Spec:   v2alpha1.APIRuleSpec{},
			Status: v2alpha1.APIRuleStatus{},
		}

		It("should convert status", func() {
			// given
			apiRuleV2alpha1.Status = v2alpha1.APIRuleStatus{
				State:       v2alpha1.Ready,
				Description: "Reconciled successfully",
			}
			apiRuleV2 := &v2.APIRule{}

			// when
			err := apiRuleV2.ConvertFrom(apiRuleV2alpha1)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2.Status.State).To(Equal(v2.Ready))
			Expect(apiRuleV2.Status.Description).To(Equal("Reconciled successfully"))

		})

	})
})
