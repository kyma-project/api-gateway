package v1beta2_test

import (
	"encoding/json"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("APIRule Conversion", func() {
	host1 := "host1"
	host2 := "host2"

	Describe("v1beta2 to v1beta1", func() {
		It("should have origin version annotation", func() {
			src := v1beta2.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*string{&host1},
				},
			}
			dst := v1beta1.APIRule{}

			err := src.ConvertTo(&dst)

			Expect(err).To(BeNil())
			Expect(dst.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/origin-version", "v1beta2"))
		})

		It("should keep the first host", func() {
			src := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*string{&host1, &host2},
				},
			}
			dst := v1beta1.APIRule{}

			err := src.ConvertTo(&dst)

			Expect(err).To(BeNil())
			Expect(*dst.Spec.Host).To(Equal(host1))
		})

		It("should convert NoAuth to v1beta1", func() {
			noAuthTrue := true

			src := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*string{&host1},
					Rules: []v1beta2.Rule{
						{
							NoAuth: &noAuthTrue,
						},
					},
				},
			}
			dst := v1beta1.APIRule{}

			err := src.ConvertTo(&dst)

			Expect(err).To(BeNil())
			Expect(dst.Spec.Rules).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
			Expect(dst.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
		})

		It("should convert JWT to v1beta1", func() {
			jwtHeaders := []*v1beta2.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			srcJwt := v1beta2.JwtConfig{
				Authentications: []*v1beta2.JwtAuthentication{
					{
						Issuer:      "issuer",
						JwksUri:     "jwksUri",
						FromHeaders: jwtHeaders,
						FromParams:  []string{"param1", "param2"},
					},
				},
				Authorizations: []*v1beta2.JwtAuthorization{
					{
						RequiredScopes: []string{"scope1", "scope2"},
						Audiences:      []string{"audience1", "audience2"},
					},
				},
			}

			src := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*string{&host1},
					Rules: []v1beta2.Rule{
						{
							Jwt: &srcJwt,
						},
					},
				},
			}
			dst := v1beta1.APIRule{}

			err := src.ConvertTo(&dst)

			Expect(err).To(BeNil())
			Expect(dst.Spec.Rules).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(dst.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())

			var v1beta1JWT v1beta1.JwtConfig
			err = json.Unmarshal(dst.Spec.Rules[0].AccessStrategies[0].Config.Raw, &v1beta1JWT)
			Expect(err).To(BeNil())

			Expect(v1beta1JWT.Authentications).To(HaveLen(1))
			Expect(v1beta1JWT.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(v1beta1JWT.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(v1beta1JWT.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(v1beta1JWT.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeaders[0].Name))
			Expect(v1beta1JWT.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeaders[0].Prefix))
			Expect(v1beta1JWT.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
			Expect(v1beta1JWT.Authorizations).To(HaveLen(1))
			Expect(v1beta1JWT.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
			Expect(v1beta1JWT.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
		})
	})

	Describe("v1beta1 to v1beta2", func() {
		It("should have the host in array", func() {
			src := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1,
				},
			}
			dst := v1beta2.APIRule{}

			err := dst.ConvertFrom(&src)

			Expect(err).To(BeNil())
			Expect(dst.Spec.Hosts).To(HaveExactElements(&host1))
		})

		It("should convert NoAuth to v1beta2", func() {
			accessStrategiesNoAuth := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "no_auth",
					},
				},
			}

			src := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: accessStrategiesNoAuth,
						},
					},
				},
			}
			dst := v1beta2.APIRule{}

			err := dst.ConvertFrom(&src)

			Expect(err).To(BeNil())
			Expect(dst.Spec.Rules).To(HaveLen(1))
			Expect(*dst.Spec.Rules[0].NoAuth).To(BeTrue())
		})

		It("should convert JWT to v1beta2", func() {
			jwtHeaders := []*v1beta1.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			srcJwt := v1beta1.JwtConfig{
				Authentications: []*v1beta1.JwtAuthentication{
					{
						Issuer:      "issuer",
						JwksUri:     "jwksUri",
						FromHeaders: jwtHeaders,
						FromParams:  []string{"param1", "param2"},
					},
				},
				Authorizations: []*v1beta1.JwtAuthorization{
					{
						RequiredScopes: []string{"scope1", "scope2"},
						Audiences:      []string{"audience1", "audience2"},
					},
				},
			}
			jwtBytes, err := json.Marshal(srcJwt)
			Expect(err).To(BeNil())

			accessStrategiesJwt := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name:   "jwt",
						Config: &runtime.RawExtension{Raw: jwtBytes},
					},
				},
			}

			src := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: accessStrategiesJwt,
						},
					},
				},
			}
			dst := v1beta2.APIRule{}

			err = dst.ConvertFrom(&src)

			Expect(err).To(BeNil())
			Expect(dst.Spec.Rules).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].Jwt).ToNot(BeNil())
			Expect(dst.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeaders[0].Name))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeaders[0].Prefix))
			Expect(dst.Spec.Rules[0].Jwt.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
			Expect(dst.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
			Expect(dst.Spec.Rules[0].Jwt.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
			Expect(dst.Spec.Rules[0].Jwt.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
		})
	})
})
