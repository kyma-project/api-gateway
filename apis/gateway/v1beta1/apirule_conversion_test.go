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
	})

	Describe("Convert from v2alpha1 to v1beta1", func() {
		It("should convert", func() {})
	})

})

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

//var _ = Describe("Is regex base path", func() {
//	It("path without regex", func() {
//		result := v1beta1.isConvertablePath("/path/without/regex")
//		Expect(result).To(BeFalse())
//	})
//	It("path with regex", func() {
//		result := v1beta1.isConvertablePath("/path/with/special/characters/[]{}")
//		Expect(result).To(BeTrue())
//	})
//	It("empty path", func() {
//		result := v1beta1.isConvertablePath("")
//		Expect(result).To(BeTrue())
//	})
//	It("path with wildcard", func() {
//		result := v1beta1.isConvertablePath("/path/with/*")
//		Expect(result).To(BeFalse())
//	})
//
//	It("path with double wildcard", func() {
//		result := v1beta1.isConvertablePath("/path/with/**")
//		Expect(result).To(BeFalse())
//	})
//
//	It("path with special characters", func() {
//		result := v1beta1.isConvertablePath("/path/with/special/characters/!@#$%^&*()")
//		Expect(result).To(BeTrue())
//	})
//
//	It("path with encoded characters", func() {
//		result := v1beta1.isConvertablePath("/path/with/encoded/characters/%20%21%40")
//		Expect(result).To(BeFalse())
//	})
//
//})

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
