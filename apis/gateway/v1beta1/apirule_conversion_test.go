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
	namespace := "test-namespace"
	var port uint32 = 8080
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
		It("should have origin version annotation not changed (already v2alpha1)", func() {
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
			_ = convertOverJson(struct {
				Headers map[string]string `json:"headers"`
			}{
				Headers: map[string]string{
					"header1": "value1",
				},
			}, &headerConfig)

			_ = convertOverJson(struct {
				Cookies map[string]string `json:"cookies"`
			}{
				Cookies: map[string]string{
					"cookie1": "value2",
				},
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

		It("should store spec in annotation", func() {
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
					Host:    &host1string,
					Gateway: ptr.To("gateway-test"),
					Service: &v1beta1.Service{
						Name:      ptr.To("service-test"),
						Port:      &port,
						Namespace: &namespace,
					},
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
			Expect(apiRuleV2alpha1.Annotations["gateway.kyma-project.io/v1beta1-spec"]).To(BeEquivalentTo(`{"host":"host1","service":{"name":"service-test","namespace":"test-namespace","port":8080},"gateway":"gateway-test","rules":[{"path":"/path1","service":{"name":"service","port":null},"methods":["GET","POST"],"accessStrategies":[{"handler":"jwt"}],"mutators":[{"handler":"header","config":{"header1":"value1"}},{"handler":"cookie","config":{"cookie1":"value2"}}]}]}`))
		})
		It("should convert spec from annotation for v2alpha1 stored in v1beta1", func() {
			v1alpha1AnnotationRules := `[{"path":"/*","methods":["GET"],"noAuth":true}]`

			namespace := "test-namespace"
			var port uint32 = 80

			apiRuleV1beta1 := v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
					Annotations: map[string]string{
						"gateway.kyma-project.io/original-version": "v2alpha1",
						"gateway.kyma-project.io/v2alpha1-rules":   v1alpha1AnnotationRules,
					},
				},
				Status: testV1StatusOK,
				Spec: v1beta1.APIRuleSpec{
					Host:    &host1string,
					Gateway: ptr.To("gateway-test"),
					Service: &v1beta1.Service{
						Name:      ptr.To("service-test"),
						Port:      &port,
						Namespace: &namespace,
					},
				}}
			apiRuleV2alpha1 := v2alpha1.APIRule{}
			// when
			err := apiRuleV1beta1.ConvertTo(&apiRuleV2alpha1)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleV2alpha1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleV2alpha1.Spec.Rules[0].NoAuth).To(Equal(ptr.To(true)))
			Expect(apiRuleV2alpha1.Spec.Rules[0].Methods).To(Equal([]v2alpha1.HttpMethod{"GET"}))
		})
	})

	Describe("v2alpha1 to v1beta1", func() {
		host1 := v2alpha1.Host("host1")

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
		It("should convert spec from annotation", func() {
			// given
			apiRuleV2alpha1 := v2alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"gateway.kyma-project.io/original-version": "v1beta1",
						"gateway.kyma-project.io/v1beta1-spec":     `{"host":"host1","service":{"name":"service-test","namespace":"test-namespace","port":8080},"gateway":"gateway-test","rules":[{"path":"/path1","service":{"name":"service","port":null},"methods":["GET","POST"],"accessStrategies":[{"handler":"jwt"}],"mutators":[{"handler":"header","config":{"header1":"value1"}},{"handler":"cookie","config":{"cookie1":"value2"}}]}]}`,
					},
				},
			}

			apiRuleBeta1 := v1beta1.APIRule{}

			// when
			err := apiRuleBeta1.ConvertFrom(&apiRuleV2alpha1)
			var port uint32 = 8080
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiRuleBeta1.Spec.Rules).To(HaveLen(1))
			Expect(apiRuleBeta1.Spec.Host).To(Equal(ptr.To(host1string)))
			Expect(apiRuleBeta1.Spec.Service).To(Equal(ptr.To(v1beta1.Service{
				Name:      ptr.To("service-test"),
				Namespace: ptr.To("test-namespace"),
				Port:      &port,
			})))
			Expect(*apiRuleBeta1.Spec.Gateway).To(Equal("gateway-test"))
			Expect(apiRuleBeta1.Spec.Rules[0].Path).To(Equal("/path1"))
		})
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
