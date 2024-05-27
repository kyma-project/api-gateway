package v1beta2_test

import (
	"encoding/json"
	"time"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var _ = Describe("APIRule Conversion", func() {
	host1string := "host1"
	host2string := "host2"
	host1 := v1beta2.Host(host1string)
	host2 := v1beta2.Host(host2string)

	Describe("v1beta2 to v1beta1", func() {
		It("should have origin version annotation", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v1beta2"))
		})

		It("should keep the first host", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1, &host2},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(*apiRuleBeta1.Spec.Host).To(Equal(string(host1)))
		})

		It("should convert NoAuth to v1beta1", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							NoAuth: ptr.To(true),
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
		})

		It("should convert rule with nested data to v1beta1", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta2.Service{Name: ptr.To("service")},
					Hosts:   []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							Path:    "/path1",
							Service: &v1beta2.Service{Name: ptr.To("rule-service")},
							NoAuth:  ptr.To(true),
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(*apiRuleBeta1.Spec.Gateway).To(Equal("gateway"))
			Expect(*apiRuleBeta1.Spec.Service.Name).To(Equal("service"))
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].Path).To(Equal("/path1"))
			Expect(*apiRuleBeta1.Spec.Rules[0].Service.Name).To(Equal("rule-service"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
		})

		It("should convert JWT to v1beta1", func() {
			// given
			jwtHeadersBeta2 := []*v1beta2.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							Jwt: &v1beta2.JwtConfig{
								Authentications: []*v1beta2.JwtAuthentication{
									{
										Issuer:      "issuer",
										JwksUri:     "jwksUri",
										FromHeaders: jwtHeadersBeta2,
										FromParams:  []string{"param1", "param2"},
									},
								},
								Authorizations: []*v1beta2.JwtAuthorization{
									{
										RequiredScopes: []string{"scope1", "scope2"},
										Audiences:      []string{"audience1", "audience2"},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())

			jwtConfigBeta1 := apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config.Object.(*v1beta2.JwtConfig)

			Expect(jwtConfigBeta1.Authentications).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(jwtConfigBeta1.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeadersBeta2[0].Name))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeadersBeta2[0].Prefix))
			Expect(jwtConfigBeta1.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
			Expect(jwtConfigBeta1.Authorizations).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
			Expect(jwtConfigBeta1.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
		})

		It("should convert single rule with JWT and ignore NoAuth when set to false to v1beta1", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							NoAuth: ptr.To(false),
							Jwt: &v1beta2.JwtConfig{
								Authentications: []*v1beta2.JwtAuthentication{},
								Authorizations:  []*v1beta2.JwtAuthorization{},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())
		})

		It("should convert two rules with NoAuth and JWT to v1beta1", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							NoAuth: ptr.To(true),
						},
						{
							Jwt: &v1beta2.JwtConfig{
								Authentications: []*v1beta2.JwtAuthentication{},
								Authorizations:  []*v1beta2.JwtAuthorization{},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(2))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config).To(BeNil())
			Expect(apiRuleBeta1.Spec.Rules[1].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[1].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(apiRuleBeta1.Spec.Rules[1].AccessStrategies[0].Config).ToNot(BeNil())
		})

		It("should fail when jwt is not configured and no_auth is set to false", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
					Rules: []v1beta2.Rule{
						{
							NoAuth: ptr.To(false),
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("either jwt is configured or noAuth must be set to true in a rule"))
		})

 		It("should convert CORS maxAge from seconds as uint64 to duration", func() {
 			// given
 			apiRuleBeta2 := v1beta2.APIRule{
 				Spec: v1beta2.APIRuleSpec{
 					Hosts: []*v1beta2.Host{&host1},
 					CorsPolicy: &v1beta2.CorsPolicy{
 						MaxAge: 60,
 					},
 				},
 			}
 			apiRuleBeta1 := v1beta1.APIRule{}

 			// when
 			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

 			// then
 			Expect(err).To(BeNil())
 			Expect(apiRuleBeta1.Spec.CorsPolicy.MaxAge).To(Equal(&metav1.Duration{Duration: time.Minute}))
 		})

		It("should convert v1beta2 Ready state to OK status from APIRuleStatus", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
				},
				Status: v1beta2.APIRuleStatus{
					State:       v1beta2.Ready,
					Description: "description",
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta1.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusOK))
			Expect(apiRuleBeta1.Status.APIRuleStatus.Description).To(Equal("description"))
		})

		It("should convert v1beta2 Error state to Error status from APIRuleStatusError status from APIRuleStatus", func() {
			// given
			apiRuleBeta2 := v1beta2.APIRule{
				Spec: v1beta2.APIRuleSpec{
					Hosts: []*v1beta2.Host{&host1},
				},
				Status: v1beta2.APIRuleStatus{
					State:       v1beta2.Error,
					Description: "description",
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta2.ConvertTo(&apiRuleBeta1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta1.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusError))
			Expect(apiRuleBeta1.Status.APIRuleStatus.Description).To(Equal("description"))
		})
	})

	Describe("v1beta1 to v1beta2", func() {
		It("should have the host in array", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta2.Spec.Hosts).To(HaveExactElements(&host1))
		})

		It("should convert NoAuth to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "no_auth",
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta2.Spec.Rules).To(HaveLen(1))
			Expect(*apiRuleBeta2.Spec.Rules[0].NoAuth).To(BeTrue())
		})

		It("should convert rule with nested data to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta1.Service{Name: ptr.To("service")},
					Host:    &host1string,
					Rules: []v1beta1.Rule{
						{
							Path:    "/path1",
							Service: &v1beta1.Service{Name: ptr.To("rule-service")},
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "no_auth",
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(*apiRuleBeta2.Spec.Gateway).To(Equal("gateway"))
			Expect(*apiRuleBeta2.Spec.Service.Name).To(Equal("service"))
			Expect(apiRuleBeta2.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Path).To(Equal("/path1"))
			Expect(*apiRuleBeta2.Spec.Rules[0].Service.Name).To(Equal("rule-service"))
			Expect(*apiRuleBeta2.Spec.Rules[0].NoAuth).To(BeTrue())
		})

		It("should convert JWT to v1beta2", func() {
			// given
			jwtHeadersBeta1 := []*v1beta1.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			jwtConfigBeta1 := v1beta1.JwtConfig{
				Authentications: []*v1beta1.JwtAuthentication{
					{
						Issuer:      "issuer",
						JwksUri:     "jwksUri",
						FromHeaders: jwtHeadersBeta1,
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

			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name:   "jwt",
										Config: &runtime.RawExtension{Object: &jwtConfigBeta1},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta2.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt).ToNot(BeNil())
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeadersBeta1[0].Name))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeadersBeta1[0].Prefix))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
		})

		It("should convert JWT to v1beta2 when config stored as raw", func() {
			// given
			jwtHeadersBeta1 := []*v1beta1.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			jwtConfigBeta1 := v1beta1.JwtConfig{
				Authentications: []*v1beta1.JwtAuthentication{
					{
						Issuer:      "issuer",
						JwksUri:     "jwksUri",
						FromHeaders: jwtHeadersBeta1,
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
			jsonConfig, err := json.Marshal(jwtConfigBeta1)
			Expect(err).ToNot(HaveOccurred())

			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name:   "jwt",
										Config: &runtime.RawExtension{Raw: jsonConfig},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err = apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta2.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt).ToNot(BeNil())
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
		})

		It("should convert two rules with NoAuth and JWT to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "no_auth",
									},
								},
							},
						},
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "jwt",
										Config: &runtime.RawExtension{
											Object: &v1beta1.JwtConfig{},
										},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta2.Spec.Rules).To(HaveLen(2))
			Expect(*apiRuleBeta2.Spec.Rules[0].NoAuth).To(BeTrue())
			Expect(apiRuleBeta2.Spec.Rules[0].Jwt).To(BeNil())
			Expect(apiRuleBeta2.Spec.Rules[1].NoAuth).To(BeNil())
			Expect(apiRuleBeta2.Spec.Rules[1].Jwt).ToNot(BeNil())
		})

		It("should fail to convert rule with ory jwt to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-name",
				},
				Spec: v1beta1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta1.Service{Name: ptr.To("service")},
					Host:    &host1string,
					Rules: []v1beta1.Rule{
						{
							Path:    "/path1",
							Service: &v1beta1.Service{Name: ptr.To("rule-service")},
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "jwt",
										Config: &runtime.RawExtension{
											Raw: []byte(`{
												"trusted_issuers": ["issuer"],
												"jwks_urls": ["jwksUri"],
												"required_scope": ["scope1", "scope2"],
												"target_audience": ["audience1", "audience2"]
											}`),
										},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("APIRule in version v1beta1 has been deprecated. To request APIRule v1beta1, use the command 'kubectl get -n test-ns apirules.v1beta1.gateway.kyma-project.io test-name'. See APIRule v1beta2 documentation and consider migrating to the newer version."))
		})

		It("should fail to convert rule with handler different to no_auth or jwt to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-name",
				},
				Spec: v1beta1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta1.Service{Name: ptr.To("service")},
					Host:    &host1string,
					Rules: []v1beta1.Rule{
						{
							Path:    "/path1",
							Service: &v1beta1.Service{Name: ptr.To("rule-service")},
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name:   "any_handler",
										Config: &runtime.RawExtension{},
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("APIRule in version v1beta1 has been deprecated. To request APIRule v1beta1, use the command 'kubectl get -n test-ns apirules.v1beta1.gateway.kyma-project.io test-name'. See APIRule v1beta2 documentation and consider migrating to the newer version."))
		})

		It("should fail to convert two rules with JWT and allow to v1beta2", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-name",
				},
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{
										Name: "jwt",
										Config: &runtime.RawExtension{
											Object: &v1beta1.JwtConfig{},
										},
									},
								},
								{
									Handler: &v1beta1.Handler{
										Name: "allow",
									},
								},
							},
						},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("APIRule in version v1beta1 has been deprecated. To request APIRule v1beta1, use the command 'kubectl get -n test-ns apirules.v1beta1.gateway.kyma-project.io test-name'. See APIRule v1beta2 documentation and consider migrating to the newer version."))
		})

		It("should convert OK status from APIRuleStatus to v1beta2 Ready state", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
				Status: v1beta1.APIRuleStatus{
					APIRuleStatus: &v1beta1.APIRuleResourceStatus{
						Code:        v1beta1.StatusOK,
						Description: "description",
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta2.Status.State).To(Equal(v1beta2.Ready))
			Expect(apiRuleBeta2.Status.Description).To(Equal("description"))
		})

		It("should convert Error status from APIRuleStatus to v1beta2 Error state", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
				Status: v1beta1.APIRuleStatus{
					APIRuleStatus: &v1beta1.APIRuleResourceStatus{
						Code:        v1beta1.StatusError,
						Description: "description",
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta2.Status.State).To(Equal(v1beta2.Error))
			Expect(apiRuleBeta2.Status.Description).To(Equal("description"))
		})

		It("should convert CORS maxAge from duration to seconds as uint64, ignoring values less than 1 second", func() {
			// given
			apiRuleBeta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					CorsPolicy: &v1beta1.CorsPolicy{
						MaxAge: &metav1.Duration{Duration: time.Minute + time.Millisecond},
					},
				},
			}
			apiRuleBeta2 := v1beta2.APIRule{}

			// when
			err := apiRuleBeta2.ConvertFrom(&apiRuleBeta1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta2.Spec.CorsPolicy.MaxAge).To(Equal(uint64(60)))
		})
	})
})
