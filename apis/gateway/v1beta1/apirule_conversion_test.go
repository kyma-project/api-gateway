package v1beta1_test

import (
	"encoding/json"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"time"
)

var _ = Describe("APIRule Conversion", func() {
	host1string := "host1"
	testV1StatusOK := v1beta1.APIRuleStatus{
		APIRuleStatus: &v1beta1.APIRuleResourceStatus{
			Code:        v1beta1.StatusOK,
			Description: "test description",
		},
	}
	testObjectMeta := metav1.ObjectMeta{
		Name: "test",
	}

	noAuthRuleWithConvertablePath := v1beta1.Rule{
		Path: "/path1",
		AccessStrategies: []*v1beta1.Authenticator{
			{
				Handler: &v1beta1.Handler{Name: v1beta1.AccessStrategyNoAuth, Config: nil},
			},
		},
	}

	defaultValidSpec := v1beta1.APIRuleSpec{
		Rules: []v1beta1.Rule{noAuthRuleWithConvertablePath},
		Host:  &host1string,
	}

	Describe("v1beta1 to v2alpha1", func() {
		It("should have origin version annotation", func() {
			// given

			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: testObjectMeta,
				Status:     testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v1beta1"))
		})
		It("should have origin version annotation", func() {
			// given
			objectMetadata := testObjectMeta
			objectMetadata.Annotations = map[string]string{
				"gateway.kyma-project.io/original-version": "v2alpha1",
			}
			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: objectMetadata,
				Status:     testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v2alpha1"))
		})

		It("should convert host to array", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: testObjectMeta,
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
				},
				Status: testV1StatusOK,
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Spec.Hosts).To(HaveLen(1))
			Expect(string(*apiRuleV2alpha1.Spec.Hosts[0])).To(Equal(host1string))
		})

		It("should convert no_auth to v2alpha1", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: testObjectMeta,
				Status:     testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					Rules: []v1beta1.Rule{
						{
							Path: "/path1",
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{Name: v1beta1.AccessStrategyNoAuth},
								},
							},
						},
					},
				},
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
			Expect(*apiRuleV2alpha1.Spec.Rules[0].NoAuth).To(BeTrue())
		})

		It("should convert rule with nested data to v2alpha1", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: testObjectMeta,
				Status:     testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta1.Service{Name: ptr.To("rule-service")},
					Host:    &host1string,
					Rules: []v1beta1.Rule{
						{
							Path: "/path1",
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{Name: v1beta1.AccessStrategyNoAuth, Config: nil},
								},
							},
						},
					},
				},
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			//then
			Expect(err).ToNot(HaveOccurred())
			Expect(*apiRuleV2alpha1.Spec.Gateway).To(Equal("gateway"))
			Expect(*apiRuleV2alpha1.Spec.Service.Name).To(Equal("rule-service"))
			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Path).To(Equal("/path1"))
			Expect(apiRuleV2alpha1.Spec.Rules[0].NoAuth).To(Equal(ptr.To(true)))
		})

		It("should convert JWT to v2alpha1 with empty spec (not enough data to make conversion)", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: testObjectMeta,
				Status:     testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v1beta1.Service{Name: ptr.To("rule-service")},
					Host:    &host1string,
					Rules: []v1beta1.Rule{
						{
							Path: "/path1",
							AccessStrategies: []*v1beta1.Authenticator{
								{
									Handler: &v1beta1.Handler{Name: v1beta1.AccessStrategyJwt, Config: nil},
								},
							},
						},
					},
				},
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			//then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Spec).To(Equal(v2alpha1.APIRuleSpec{}))
		})

		It("should convert CORS maxAge from duration to seconds as uint64", func() {
			// given

			apiRuleV1beta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					CorsPolicy: &v1beta1.CorsPolicy{
						MaxAge: &metav1.Duration{Duration: time.Minute},
					},
					Rules: []v1beta1.Rule{noAuthRuleWithConvertablePath},
				},
				Status: testV1StatusOK,
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleV2alpha1.Spec.CorsPolicy.MaxAge).To(Equal(ptr.To(uint64(60))))
		})

		It("should convert CORS Policy when MaxAge is not set and don't set a default", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				Spec: v1beta1.APIRuleSpec{
					Host: &host1string,
					CorsPolicy: &v1beta1.CorsPolicy{
						AllowCredentials: ptr.To(true),
					},
					Rules: []v1beta1.Rule{noAuthRuleWithConvertablePath},
				},
				Status: testV1StatusOK,
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(*apiRuleV2alpha1.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
			Expect(apiRuleV2alpha1.Spec.CorsPolicy.MaxAge).To(BeNil())
		})

		It("should convert v1beta1 OK state to Ready status from APIRuleStatus", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				Status: testV1StatusOK,
				Spec:   defaultValidSpec,
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleV2alpha1.Status.State).To(Equal(v2alpha1.Ready))
			Expect(apiRuleV2alpha1.Status.Description).To(Equal(testV1StatusOK.APIRuleStatus.Description))
		})

		It("should convert v1 Error state to Error status from APIRuleStatus status to APIRuleStatusError ", func() {
			// given
			errorDescription := "error description"

			apiRuleV1beta1 := v1beta1.APIRule{
				Status: v1beta1.APIRuleStatus{
					APIRuleStatus: &v1beta1.APIRuleResourceStatus{
						Description: errorDescription,
						Code:        v1beta1.StatusError,
					},
				},
				Spec: defaultValidSpec,
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleV2alpha1.Status.State).To(Equal(v2alpha1.Error))
			Expect(apiRuleV2alpha1.Status.Description).To(Equal(errorDescription))
		})

		It("should convert rule with empty spec", func() {
			// given
			apiRuleV1beta1 := v1beta1.APIRule{
				Spec:   v1beta1.APIRuleSpec{},
				Status: testV1StatusOK,
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Spec).To(Equal(v2alpha1.APIRuleSpec{}))
		})

		It("should convert mutator to request", func() {
			var headerConfig, cookieConfig runtime.RawExtension
			_ = convertOverJson(map[string]string{
				"header1": "value1",
			}, &headerConfig)

			_ = convertOverJson(map[string]string{
				"cookie1": "value2",
			}, &cookieConfig)

			apiRuleV1beta1 := v1beta1.APIRule{
				Status: testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Rules: []v1beta1.Rule{
						{
							Path: "/path1",
							Mutators: []*v1beta1.Mutator{
								{
									Handler: &v1beta1.Handler{
										Name:   v1beta1.HeaderMutator,
										Config: &headerConfig,
									},
								},
								{
									Handler: &v1beta1.Handler{
										Name:   v1beta1.CookieMutator,
										Config: &cookieConfig,
									},
								},
							},
						},
					},
				},
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Request.Cookies).To(HaveLen(1))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Request.Cookies).To(Equal(map[string]string{
				"cookie1": "value2",
			}))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Request.Headers).To(HaveLen(1))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Request.Headers).To(Equal(map[string]string{
				"header1": "value1",
			}))

		})

		It("should store rules in annotation", func() {
			var headerConfig, cookieConfig runtime.RawExtension
			_ = convertOverJson(map[string]string{
				"header1": "value1",
			}, &headerConfig)

			_ = convertOverJson(map[string]string{
				"cookie1": "value2",
			}, &cookieConfig)

			apiRuleV1beta1 := v1beta1.APIRule{
				Status: testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Rules: []v1beta1.Rule{
						{
							Path: "/path1",
							Service: &v1beta1.Service{
								Name: ptr.To("service"),
							},
							Methods: []v1beta1.HttpMethod{"GET", "POST"},
							AccessStrategies: []*v1beta1.Authenticator{
								{Handler: &v1beta1.Handler{Name: v1beta1.AccessStrategyJwt, Config: nil}},
							},
							Mutators: []*v1beta1.Mutator{
								{
									Handler: &v1beta1.Handler{
										Name:   v1beta1.HeaderMutator,
										Config: &headerConfig,
									},
								},
								{
									Handler: &v1beta1.Handler{
										Name:   v1beta1.CookieMutator,
										Config: &cookieConfig,
									},
								},
							},
						},
					},
				},
			}
			apiRuleV2alpha1 := v2alpha1.APIRule{}

			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Annotations["gateway.kyma-project.io/v1beta1-rules"]).To(BeEquivalentTo(`[{"path":"/path1","service":{"name":"service","port":null},"methods":["GET","POST"],"accessStrategies":[{"handler":"jwt"}],"mutators":[{"handler":"header","config":{"header1":"value1"}},{"handler":"cookie","config":{"cookie1":"value2"}}]}]`))
		})

	})

	Describe("v2alpha1 to v1beta1", func() {
		host1 := v2alpha1.Host("host1")
		host2 := v2alpha1.Host("host2")

		It("should have origin version annotation", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"gateway.kyma-project.io/original-version": "v2alpha1",
					},
					Name: "test",
				},
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v2alpha1"))
		})

		It("should keep the first host", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1, &host2},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(*apiRuleBeta1.Spec.Host).To(Equal(string(host1)))
		})

		It("should convert NoAuth to v1beta1", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							NoAuth: ptr.To(true),
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
		})

		It("should convert rule with nested data to v1beta1", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Gateway: ptr.To("gateway"),
					Service: &v2alpha1.Service{Name: ptr.To("service")},
					Hosts:   []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							Path:    "/path1",
							Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
							NoAuth:  ptr.To(true),
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

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
			jwtHeadersV2Alpha1 := []*v2alpha1.JwtHeader{
				{Name: "header1", Prefix: "prefix1"},
			}

			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							Jwt: &v2alpha1.JwtConfig{
								Authentications: []*v2alpha1.JwtAuthentication{
									{
										Issuer:      "issuer",
										JwksUri:     "jwksUri",
										FromHeaders: jwtHeadersV2Alpha1,
										FromParams:  []string{"param1", "param2"},
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
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())

			jwtConfigBeta1 := apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config.Object.(*v2alpha1.JwtConfig)

			Expect(jwtConfigBeta1.Authentications).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authentications[0].Issuer).To(Equal("issuer"))
			Expect(jwtConfigBeta1.Authentications[0].JwksUri).To(Equal("jwksUri"))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeadersV2Alpha1[0].Name))
			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeadersV2Alpha1[0].Prefix))
			Expect(jwtConfigBeta1.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
			Expect(jwtConfigBeta1.Authorizations).To(HaveLen(1))
			Expect(jwtConfigBeta1.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
			Expect(jwtConfigBeta1.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
		})

		It("should convert single rule with JWT and ignore NoAuth when set to false to v1beta1", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							NoAuth: ptr.To(false),
							Jwt: &v2alpha1.JwtConfig{
								Authentications: []*v2alpha1.JwtAuthentication{},
								Authorizations:  []*v2alpha1.JwtAuthorization{},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
			Expect(apiRuleBeta1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())
		})

		It("should convert two rules with NoAuth and JWT to v1beta1", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							NoAuth: ptr.To(true),
						},
						{
							Jwt: &v2alpha1.JwtConfig{
								Authentications: []*v2alpha1.JwtAuthentication{},
								Authorizations:  []*v2alpha1.JwtAuthorization{},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

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

		It("should convert CORS maxAge from seconds as uint64 to duration", func() {
			// given
			maxAge := uint64(60)
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					CorsPolicy: &v2alpha1.CorsPolicy{
						MaxAge: &maxAge,
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta1.Spec.CorsPolicy.MaxAge).To(Equal(&metav1.Duration{Duration: time.Minute}))
		})

		It("should convert CORS Policy when MaxAge is not set and don't set a default", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					CorsPolicy: &v2alpha1.CorsPolicy{
						AllowCredentials: ptr.To(true),
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(*apiRuleBeta1.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
			Expect(apiRuleBeta1.Spec.CorsPolicy.MaxAge).To(BeNil())
		})

		It("should convert CORS Policy when MaxAge is not set and don't set a default", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					CorsPolicy: &v2alpha1.CorsPolicy{
						AllowCredentials: ptr.To(true),
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(*apiRuleBeta1.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
			Expect(apiRuleBeta1.Spec.CorsPolicy.MaxAge).To(BeNil())
		})

		It("should convert v2alpha1 Ready state to OK status from APIRuleStatus", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
				},
				Status: v2alpha1.APIRuleStatus{
					State:       v2alpha1.Ready,
					Description: "description",
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta1.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusOK))
			Expect(apiRuleBeta1.Status.APIRuleStatus.Description).To(Equal("description"))
		})

		It("should convert v2alpha1 Error state to Error status from APIRuleStatusError status from APIRuleStatus", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
				},
				Status: v2alpha1.APIRuleStatus{
					State:       v2alpha1.Error,
					Description: "description",
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).To(BeNil())
			Expect(apiRuleBeta1.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusError))
			Expect(apiRuleBeta1.Status.APIRuleStatus.Description).To(Equal("description"))
		})

		It("should convert rule with empty spec", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec).To(Equal(v1beta1.APIRuleSpec{}))
		})

		It("should convert request to mutator", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Hosts: []*v2alpha1.Host{&host1},
					Rules: []v2alpha1.Rule{
						{
							Request: &v2alpha1.Request{
								Headers: map[string]string{
									"header1": "value1",
								},

								Cookies: map[string]string{
									"cookie1": "value1",
								},
							},
						},
					},
				},
			}
			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Rules[0].Mutators).To(HaveLen(2))

			var configMap map[string]string
			correctMutators := 0

			for _, mutator := range apiRuleBeta1.Spec.Rules[0].Mutators {
				if mutator.Handler.Name == v1beta1.HeaderMutator {
					correctMutators++
					Expect(mutator.Handler.Config.Raw).ToNot(BeNil())
					err := json.Unmarshal(mutator.Handler.Config.Raw, &configMap)
					Expect(err).ToNot(HaveOccurred())
					Expect(configMap).To(HaveKeyWithValue("header1", "value1"))
				}
				if mutator.Handler.Name == v1beta1.CookieMutator {
					correctMutators++
					Expect(mutator.Handler.Config.Raw).ToNot(BeNil())
					err := json.Unmarshal(mutator.Handler.Config.Raw, &configMap)
					Expect(err).ToNot(HaveOccurred())
					Expect(configMap).To(HaveKeyWithValue("cookie1", "value1"))
				}
			}
			Expect(correctMutators).To(Equal(2))
		})
		It("should convert rules from annotation", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"gateway.kyma-project.io/original-version": "v1beta1",
						"gateway.kyma-project.io/v1beta1-rules":    `[{"path":"/path1","service":{"name":"service","port":null},"methods":["GET","POST"],"accessStrategies":[{"handler":"jwt"}],"mutators":[{"handler":"header","config":{"header1":"value1"}},{"handler":"cookie","config":{"cookie1":"value2"}}]}]`,
					},
				},
			}

			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
		})
	})
})

var _ = Describe("Is regex base path", func() {
	It("path without regex", func() {
		result := v1beta1.IsConvertiblePath("/path/without/regex")
		Expect(result).To(BeTrue())
	})
	It("path with regex", func() {
		result := v1beta1.IsConvertiblePath("/path/with/special/characters/[]{}")
		Expect(result).To(BeFalse())
	})
	It("empty path", func() {
		result := v1beta1.IsConvertiblePath("")
		Expect(result).To(BeFalse())
	})
	It("path with wildcard", func() {
		result := v1beta1.IsConvertiblePath("/path/with/*")
		Expect(result).To(BeFalse())
	})

	It("path with double wildcard", func() {
		result := v1beta1.IsConvertiblePath("/path/with/**")
		Expect(result).To(BeFalse())
	})

	It("path with special characters", func() {
		result := v1beta1.IsConvertiblePath("/path/with/special/characters/!@#$%^&*()")
		Expect(result).To(BeFalse())
	})

	It("path with encoded characters", func() {
		result := v1beta1.IsConvertiblePath("/path/with/encoded/characters/%20%21%40")
		Expect(result).To(BeTrue())
	})
	It("path with // ", func() {
		result := v1beta1.IsConvertiblePath("/path/with/{*}")
		Expect(result).To(BeTrue())
	})

	It("path with // ", func() {
		result := v1beta1.IsConvertiblePath("/path/{*}/with")
		Expect(result).To(BeTrue())
	})

	It("path with double wildcard // ", func() {
		result := v1beta1.IsConvertiblePath("/path/{**}/with")
		Expect(result).To(BeTrue())
	})

	It("path with double wildcard // ", func() {
		result := v1beta1.IsConvertiblePath("/path/{*}/with/{**}")
		Expect(result).To(BeTrue())
	})

	It("path with double wildcard // ", func() {
		result := v1beta1.IsConvertiblePath("/{*}")
		Expect(result).To(BeTrue())
	})

	It("path with double wildcard // ", func() {
		result := v1beta1.IsConvertiblePath("/{**}")
		Expect(result).To(BeTrue())
	})

	It("path with double wildcard // ", func() {
		result := v1beta1.IsConvertiblePath("/*")
		Expect(result).To(BeTrue())
	})

})

func convertOverJson(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, dst)
	if err != nil {
		return err
	}

	return nil
}
