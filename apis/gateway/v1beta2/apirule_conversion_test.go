package v1beta2_test

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("APIRule Conversion", func() {

	Describe("v1beta2 conversion to v1beta1", func() {
		host1 := "host1"
		host2 := "host2"

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

			rawJWT := dst.Spec.Rules[0].AccessStrategies[0].Config.Raw
			var jsonJWT v1beta1.JwtConfig
			err = json.Unmarshal(rawJWT, &jsonJWT)
			Expect(err).To(BeNil())

			Expect(jsonJWT.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(jsonJWT.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(jsonJWT.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(jsonJWT.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeaders[0].Name))
			Expect(jsonJWT.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeaders[0].Prefix))
			Expect(jsonJWT.Authentications[0].FromParams).To(ContainElements("param1", "param2"))
			Expect(jsonJWT.Authorizations).To(HaveLen(1))
			Expect(jsonJWT.Authorizations[0].RequiredScopes).To(ContainElements("scope1", "scope2"))
			Expect(jsonJWT.Authorizations[0].Audiences).To(ContainElements("audience1", "audience2"))
		})
	})
})
