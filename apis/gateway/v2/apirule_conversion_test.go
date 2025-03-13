package v2_test

//
//import (
//	"encoding/json"
//	"github.com/kyma-project/api-gateway/apis/gateway/v2"
//	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/utils/ptr"
//	"time"
//)
//
//var _ = Describe("APIRule Conversion", func() {
//	host1string := "host1"
//	host2string := "host2"
//	host1 := v2.Host(host1string)
//	host2 := v2.Host(host2string)
//
//	Describe("v2 to v1beta1", func() {
//		It("should have origin version annotation", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test",
//				},
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//				},
//			}
//			apiRulev2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRulev2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRulev2alpha1.Annotations).To(HaveKeyWithValue("gateway.kyma-project.io/original-version", "v2"))
//		})
//
//		It("should convert NoAuth to v1beta1", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							NoAuth: ptr.To(true),
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
//		})
//
//		It("should convert rule with nested data to v1beta1", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Gateway: ptr.To("gateway"),
//					Service: &v2.Service{Name: ptr.To("service")},
//					Hosts:   []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							Path:    "/path1",
//							Service: &v2.Service{Name: ptr.To("rule-service")},
//							NoAuth:  ptr.To(true),
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*apiRuleV2alpha1.Spec.Gateway).To(Equal("gateway"))
//			Expect(*apiRuleV2alpha1.Spec.Service.Name).To(Equal("service"))
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].Path).To(Equal("/path1"))
//			Expect(*apiRuleV2alpha1.Spec.Rules[0].Service.Name).To(Equal("rule-service"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Config).To(BeNil())
//		})
//
//		It("should convert JWT to v1beta1", func() {
//			// given
//			jwtHeadersV2 := []*v2.JwtHeader{
//				{Name: "header1", Prefix: "prefix1"},
//			}
//
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							Jwt: &v2.JwtConfig{
//								Authentications: []*v2.JwtAuthentication{
//									{
//										Issuer:      "issuer",
//										JwksUri:     "jwksUri",
//										FromHeaders: jwtHeadersV2,
//										FromParams:  []string{"param1", "param2"},
//									},
//								},
//								Authorizations: []*v2.JwtAuthorization{
//									{
//										RequiredScopes: []string{"scope1", "scope2"},
//										Audiences:      []string{"audience1", "audience2"},
//									},
//								},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())
//
//			jwtConfigBeta1 := apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Config.Object.(*v2.JwtConfig)
//
//			Expect(jwtConfigBeta1.Authentications).To(HaveLen(1))
//			Expect(jwtConfigBeta1.Authentications[0].Issuer).To(Equal("issuer"))
//			Expect(jwtConfigBeta1.Authentications[0].JwksUri).To(Equal("jwksUri"))
//			Expect(jwtConfigBeta1.Authentications[0].FromHeaders).To(HaveLen(1))
//			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeadersV2[0].Name))
//			Expect(jwtConfigBeta1.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeadersV2[0].Prefix))
//			Expect(jwtConfigBeta1.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
//			Expect(jwtConfigBeta1.Authorizations).To(HaveLen(1))
//			Expect(jwtConfigBeta1.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
//			Expect(jwtConfigBeta1.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
//		})
//
//		It("should convert single rule with JWT and ignore NoAuth when set to false to v1beta1", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							NoAuth: ptr.To(false),
//							Jwt: &v2.JwtConfig{
//								Authentications: []*v2.JwtAuthentication{},
//								Authorizations:  []*v2.JwtAuthorization{},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Config).ToNot(BeNil())
//		})
//
//		It("should convert two rules with NoAuth and JWT to v1beta1", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							NoAuth: ptr.To(true),
//						},
//						{
//							Jwt: &v2.JwtConfig{
//								Authentications: []*v2.JwtAuthentication{},
//								Authorizations:  []*v2.JwtAuthorization{},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(2))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(Equal("no_auth"))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].AccessStrategies[0].Config).To(BeNil())
//			Expect(apiRuleV2alpha1.Spec.Rules[1].AccessStrategies).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[1].AccessStrategies[0].Handler.Name).To(Equal("jwt"))
//			Expect(apiRuleV2alpha1.Spec.Rules[1].AccessStrategies[0].Config).ToNot(BeNil())
//		})
//
//		It("should convert CORS maxAge from seconds as uint64 to duration", func() {
//			// given
//			maxAge := uint64(60)
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					CorsPolicy: &v2.CorsPolicy{
//						MaxAge: &maxAge,
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2alpha1.Spec.CorsPolicy.MaxAge).To(Equal(&metav1.Duration{Duration: time.Minute}))
//		})
//
//		It("should convert CORS Policy when MaxAge is not set and don't set a default", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					CorsPolicy: &v2.CorsPolicy{
//						AllowCredentials: ptr.To(true),
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(*apiRuleV2alpha1.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
//			Expect(apiRuleV2alpha1.Spec.CorsPolicy.MaxAge).To(BeNil())
//		})
//
//		It("should convert CORS Policy when MaxAge is not set and don't set a default", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					CorsPolicy: &v2.CorsPolicy{
//						AllowCredentials: ptr.To(true),
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(*apiRuleV2alpha1.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
//			Expect(apiRuleV2alpha1.Spec.CorsPolicy.MaxAge).To(BeNil())
//		})
//
//		It("should convert v2 Ready state to OK status from APIRuleStatus", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//				},
//				Status: v2.APIRuleStatus{
//					State:       v2.Ready,
//					Description: "description",
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2alpha1.Status.APIRuleStatus.Code).To(Equal(v2alpha1.StatusOK))
//			Expect(apiRuleV2alpha1.Status.APIRuleStatus.Description).To(Equal("description"))
//		})
//
//		It("should convert v2 Error state to Error status from APIRuleStatusError status from APIRuleStatus", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//				},
//				Status: v2.APIRuleStatus{
//					State:       v2.Error,
//					Description: "description",
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2alpha1.Status.APIRuleStatus.Code).To(Equal(v2alpha1.StatusError))
//			Expect(apiRuleV2alpha1.Status.APIRuleStatus.Description).To(Equal("description"))
//		})
//
//		It("should convert rule with empty spec", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2alpha1.Spec).To(Equal(v2alpha1.APIRuleSpec{}))
//		})
//
//		It("should convert request to mutator", func() {
//			// given
//			apiRuleV2 := v2.APIRule{
//				Spec: v2.APIRuleSpec{
//					Hosts: []*v2.Host{&host1},
//					Rules: []v2.Rule{
//						{
//							Request: &v2.Request{
//								Headers: map[string]string{
//									"header1": "value1",
//								},
//
//								Cookies: map[string]string{
//									"cookie1": "value1",
//								},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2alpha1 := v2alpha1.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertTo(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//
//			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2alpha1.Spec.Rules[0].Mutators).To(HaveLen(2))
//
//			var configMap map[string]string
//			correctMutators := 0
//
//			for _, mutator := range apiRuleV2alpha1.Spec.Rules[0].Mutators {
//				if mutator.Handler.Name == v2alpha1.HeaderMutator {
//					correctMutators++
//					Expect(mutator.Handler.Config.Raw).ToNot(BeNil())
//					err := json.Unmarshal(mutator.Handler.Config.Raw, &configMap)
//					Expect(err).ToNot(HaveOccurred())
//					Expect(configMap).To(HaveKeyWithValue("header1", "value1"))
//				}
//				if mutator.Handler.Name == v2alpha1.CookieMutator {
//					correctMutators++
//					Expect(mutator.Handler.Config.Raw).ToNot(BeNil())
//					err := json.Unmarshal(mutator.Handler.Config.Raw, &configMap)
//					Expect(err).ToNot(HaveOccurred())
//					Expect(configMap).To(HaveKeyWithValue("cookie1", "value1"))
//				}
//			}
//			Expect(correctMutators).To(Equal(2))
//		})
//	})
//
//	Describe("v1beta1 to v2", func() {
//		It("should have the host in array", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2.Spec.Hosts).To(HaveExactElements(&host1))
//		})
//
//		It("should convert NoAuth to v2", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//					Rules: []v2alpha1.Rule{
//						{
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "no_auth",
//									},
//								},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//			Expect(*apiRuleV2.Spec.Rules[0].NoAuth).To(BeTrue())
//		})
//
//		It("should convert rule with nested data to v2", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Gateway: ptr.To("gateway"),
//					Service: &v2alpha1.Service{Name: ptr.To("service")},
//					Host:    &host1string,
//					Rules: []v2alpha1.Rule{
//						{
//							Path:    "/path1",
//							Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "no_auth",
//									},
//								},
//							},
//						},
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*apiRuleV2.Spec.Gateway).To(Equal("gateway"))
//			Expect(*apiRuleV2.Spec.Service.Name).To(Equal("service"))
//			Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//			Expect(apiRuleV2.Spec.Rules[0].Path).To(Equal("/path1"))
//			Expect(*apiRuleV2.Spec.Rules[0].Service.Name).To(Equal("rule-service"))
//			Expect(*apiRuleV2.Spec.Rules[0].NoAuth).To(BeTrue())
//		})
//
//		It("should convert two rules with NoAuth and JWT to v2", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//					Rules: []v2alpha1.Rule{
//						{
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "no_auth",
//									},
//								},
//							},
//						},
//						{
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "jwt",
//										Config: &runtime.RawExtension{
//											Object: &v2alpha1.JwtConfig{
//												Authentications: []*v2alpha1.JwtAuthentication{
//													{
//														Issuer:  "issuer",
//														JwksUri: "jwksUri",
//													},
//												},
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//			}
//			v2 := v2.APIRule{}
//
//			// when
//			err := v2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(v2.Spec.Rules).To(HaveLen(2))
//			Expect(*v2.Spec.Rules[0].NoAuth).To(BeTrue())
//			Expect(v2.Spec.Rules[0].Jwt).To(BeNil())
//			Expect(v2.Spec.Rules[1].NoAuth).To(BeNil())
//			Expect(v2.Spec.Rules[1].Jwt).ToNot(BeNil())
//		})
//
//		Context("with JWT", func() {
//			It("should convert JWT", func() {
//				// given
//				jwtHeadersBeta1 := []*v2alpha1.JwtHeader{
//					{Name: "header1", Prefix: "prefix1"},
//				}
//
//				jwtConfigBeta1 := v2alpha1.JwtConfig{
//					Authentications: []*v2alpha1.JwtAuthentication{
//						{
//							Issuer:      "issuer",
//							JwksUri:     "jwksUri",
//							FromHeaders: jwtHeadersBeta1,
//							FromParams:  []string{"param1", "param2"},
//						},
//					},
//					Authorizations: []*v2alpha1.JwtAuthorization{
//						{
//							RequiredScopes: []string{"scope1", "scope2"},
//							Audiences:      []string{"audience1", "audience2"},
//						},
//					},
//				}
//
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					Spec: v2alpha1.APIRuleSpec{
//						Host: &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name:   "jwt",
//											Config: &runtime.RawExtension{Object: &jwtConfigBeta1},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Name).To(Equal(jwtHeadersBeta1[0].Name))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].FromHeaders[0].Prefix).To(Equal(jwtHeadersBeta1[0].Prefix))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].FromParams).To(HaveExactElements("param1", "param2"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
//			})
//
//			It("should convert JWT without Authorizations", func() {
//				// given
//				jwtConfigBeta1 := v2alpha1.JwtConfig{
//					Authentications: []*v2alpha1.JwtAuthentication{
//						{
//							Issuer:  "issuer",
//							JwksUri: "jwksUri",
//						},
//					},
//				}
//
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					Spec: v2alpha1.APIRuleSpec{
//						Host: &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name:   "jwt",
//											Config: &runtime.RawExtension{Object: &jwtConfigBeta1},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(BeEmpty())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
//			})
//
//			It("should convert JWT without Authentications", func() {
//				// given
//				jwtConfigBeta1 := v2alpha1.JwtConfig{
//					Authorizations: []*v2alpha1.JwtAuthorization{
//						{
//							RequiredScopes: []string{"scope1", "scope2"},
//							Audiences:      []string{"aud1", "aud2"},
//						},
//					},
//				}
//
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					Spec: v2alpha1.APIRuleSpec{
//						Host: &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name:   "jwt",
//											Config: &runtime.RawExtension{Object: &jwtConfigBeta1},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(BeEmpty())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].RequiredScopes).To(ContainElements("scope1", "scope2"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].Audiences).To(ContainElements("aud1", "aud2"))
//			})
//
//			It("should convert JWT to v2 when config stored as raw", func() {
//				// given
//				jwtHeadersBeta1 := []*v2alpha1.JwtHeader{
//					{Name: "header1", Prefix: "prefix1"},
//				}
//
//				jwtConfigBeta1 := v2alpha1.JwtConfig{
//					Authentications: []*v2alpha1.JwtAuthentication{
//						{
//							Issuer:      "issuer",
//							JwksUri:     "jwksUri",
//							FromHeaders: jwtHeadersBeta1,
//							FromParams:  []string{"param1", "param2"},
//						},
//					},
//					Authorizations: []*v2alpha1.JwtAuthorization{
//						{
//							RequiredScopes: []string{"scope1", "scope2"},
//							Audiences:      []string{"audience1", "audience2"},
//						},
//					},
//				}
//				jsonConfig, err := json.Marshal(jwtConfigBeta1)
//				Expect(err).ToNot(HaveOccurred())
//
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					Spec: v2alpha1.APIRuleSpec{
//						Host: &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name:   "jwt",
//											Config: &runtime.RawExtension{Raw: jsonConfig},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err = apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
//			})
//
//			It("should convert rule with ory jwt to v2", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "jwt",
//											Config: &runtime.RawExtension{
//												Raw: []byte(`{
//												"trusted_issuers": ["issuer"],
//												"jwks_urls": ["jwksUri"],
//												"required_scope": ["scope1", "scope2"],
//												"target_audience": ["audience1", "audience2"]
//											}`),
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].RequiredScopes).To(HaveExactElements("scope1", "scope2"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations[0].Audiences).To(HaveExactElements("audience1", "audience2"))
//			})
//
//			It("should convert rule with ory jwt with trusted_issuers and jwks_urls only", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "jwt",
//											Config: &runtime.RawExtension{
//												Raw: []byte(`{
//												"trusted_issuers": ["issuer"],
//												"jwks_urls": ["jwksUri"]
//											}`),
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec.Rules).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt).ToNot(BeNil())
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications).To(HaveLen(1))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].Issuer).To(Equal("issuer"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authentications[0].JwksUri).To(Equal("jwksUri"))
//				Expect(apiRuleV2.Spec.Rules[0].Jwt.Authorizations).To(HaveLen(1))
//			})
//
//			It("should convert rule with ory jwt handler with multiple trusted_issuers to v2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "jwt",
//											Config: &runtime.RawExtension{
//												Raw: []byte(`{
//												"trusted_issuers": ["issuer", "issuer2"],
//												"jwks_urls": ["jwksUri"]
//											}`),
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//
//			It("should convert rule with ory jwt handler with multiple jwks_urls to apiRuleV2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "jwt",
//											Config: &runtime.RawExtension{
//												Raw: []byte(`{
//												"trusted_issuers": ["issuer"],
//												"jwks_urls": ["https://jwksUri.com", "https://jwksUriTwo.com"]
//											}`),
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//		})
//
//		Context("with unsupported handler", func() {
//			It("should set object meta data and status when converting handler that does not support full conversion to v2", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "allow",
//										},
//									},
//								},
//							},
//						},
//					},
//					Status: v2alpha1.APIRuleStatus{
//						APIRuleStatus: &v2alpha1.APIRuleResourceStatus{
//							Code:        v2alpha1.StatusOK,
//							Description: "description",
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Name).To(Equal("test-name"))
//				Expect(apiRuleV2.Namespace).To(Equal("test-ns"))
//				Expect(apiRuleV2.Status).ToNot(BeNil())
//				Expect(apiRuleV2.Status.State).To(Equal(v2.Ready))
//				Expect(apiRuleV2.Status.Description).To(Equal("description"))
//
//			})
//
//			It("should convert rule with allow handler to v2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "allow",
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//
//			It("should convert rule with oauth2_introspection handler to v2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "oauth2_introspection",
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//
//			It("should convert rule with noop handler to v2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Gateway: ptr.To("gateway"),
//						Service: &v2alpha1.Service{Name: ptr.To("service")},
//						Host:    &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								Path:    "/path1",
//								Service: &v2alpha1.Service{Name: ptr.To("rule-service")},
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name:   "noop",
//											Config: &runtime.RawExtension{},
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//
//			It("should convert two rules with JWT and allow to v2 with empty spec", func() {
//				// given
//				apiRuleV2alpha1 := v2alpha1.APIRule{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "test-ns",
//						Name:      "test-name",
//					},
//					Spec: v2alpha1.APIRuleSpec{
//						Host: &host1string,
//						Rules: []v2alpha1.Rule{
//							{
//								AccessStrategies: []*v2alpha1.Authenticator{
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "jwt",
//											Config: &runtime.RawExtension{
//												Object: &v2alpha1.JwtConfig{
//													Authentications: []*v2alpha1.JwtAuthentication{
//														{
//															Issuer:  "issuer",
//															JwksUri: "jwksUri",
//														},
//													},
//												},
//											},
//										},
//									},
//									{
//										Handler: &v2alpha1.Handler{
//											Name: "allow",
//										},
//									},
//								},
//							},
//						},
//					},
//				}
//				apiRuleV2 := v2.APIRule{}
//
//				// when
//				err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//				// then
//				Expect(err).ToNot(HaveOccurred())
//				Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//			})
//		})
//
//		It("should convert OK status from APIRuleStatus to v2 Ready state", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//				},
//				Status: v2alpha1.APIRuleStatus{
//					APIRuleStatus: &v2alpha1.APIRuleResourceStatus{
//						Code:        v2alpha1.StatusOK,
//						Description: "description",
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2.Status.State).To(Equal(v2.Ready))
//			Expect(apiRuleV2.Status.Description).To(Equal("description"))
//		})
//
//		It("should convert Error status from APIRuleStatus to v2 Error state", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//				},
//				Status: v2alpha1.APIRuleStatus{
//					APIRuleStatus: &v2alpha1.APIRuleResourceStatus{
//						Code:        v2alpha1.StatusError,
//						Description: "description",
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2.Status.State).To(Equal(v2.Error))
//			Expect(apiRuleV2.Status.Description).To(Equal("description"))
//		})
//
//		It("should convert CORS maxAge from duration to seconds as uint64, ignoring values less than 1 second", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Host: &host1string,
//					CorsPolicy: &v2alpha1.CorsPolicy{
//						MaxAge: &metav1.Duration{Duration: time.Minute + time.Millisecond},
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(*apiRuleV2.Spec.CorsPolicy.MaxAge).To(Equal(uint64(60)))
//		})
//
//		It("should convert CORS policy when MaxAge is not set", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Hosts: &host1string,
//					CorsPolicy: &v2alpha1.CorsPolicy{
//						AllowCredentials: ptr.To(true),
//					},
//				},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(*apiRuleV2.Spec.CorsPolicy.AllowCredentials).To(BeTrue())
//		})
//
//		It("should convert rule with empty spec", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{},
//			}
//			apiRuleV2 := v2.APIRule{}
//
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).ToNot(HaveOccurred())
//			Expect(apiRuleV2.Spec).To(Equal(v2.APIRuleSpec{}))
//		})
//
//		It("should convert mutators", func() {
//			// given
//			apiRuleV2alpha1 := v2alpha1.APIRule{
//				Spec: v2alpha1.APIRuleSpec{
//					Hosts: []v2alpha1.Host{v2alpha1.Host(host1string)},
//					Rules: []v2alpha1.Rule{
//						{
//							Path: "/path1",
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "no_auth",
//									},
//								},
//							},
//							Mutators: []*v2alpha1.Mutator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "cookie",
//										Config: getRawConfig(
//											map[string]string{
//												"cookie1": "value1",
//											},
//										),
//									},
//								},
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "header",
//										Config: getRawConfig(
//											map[string]string{
//												"header1": "value1",
//											},
//										),
//									},
//								},
//							},
//						},
//						{
//							Path: "/path2",
//							AccessStrategies: []*v2alpha1.Authenticator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "no_auth",
//									},
//								},
//							},
//							Mutators: []*v2alpha1.Mutator{
//								{
//									Handler: &v2alpha1.Handler{
//										Name: "cookie",
//										Config: getRawConfig(
//											map[string]string{
//												"cookie2": "value2",
//											},
//										),
//									},
//								},
//							},
//						},
//					},
//				},
//			}
//
//			var apiRuleV2 v2.APIRule
//			// when
//			err := apiRuleV2.ConvertFrom(&apiRuleV2alpha1)
//
//			// then
//			Expect(err).To(BeNil())
//			Expect(apiRuleV2.Spec.Rules[0].Request.Cookies).ToNot(BeNil())
//			Expect(apiRuleV2.Spec.Rules[0].Request.Cookies).To(HaveLen(1))
//			Expect(apiRuleV2.Spec.Rules[0].Request.Cookies["cookie1"]).To(Equal("value1"))
//
//			Expect(apiRuleV2.Spec.Rules[0].Request.Headers).ToNot(BeNil())
//			Expect(apiRuleV2.Spec.Rules[0].Request.Headers).To(HaveLen(1))
//			Expect(apiRuleV2.Spec.Rules[0].Request.Headers["header1"]).To(Equal("value1"))
//
//			Expect(apiRuleV2.Spec.Rules[1].Request.Cookies).ToNot(BeNil())
//			Expect(apiRuleV2.Spec.Rules[1].Request.Cookies).To(HaveLen(1))
//			Expect(apiRuleV2.Spec.Rules[1].Request.Cookies["cookie2"]).To(Equal("value2"))
//		})
//	})
//})
//
//func getRawConfig(config any) *runtime.RawExtension {
//	b, err := json.Marshal(config)
//	Expect(err).To(BeNil())
//	return &runtime.RawExtension{
//		Raw: b,
//	}
//}
