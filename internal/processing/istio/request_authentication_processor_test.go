package istio_test

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Request Authentication Processor", func() {
	createIstioJwtAccessStrategy := func() *gatewayv1beta1.Authenticator {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, JwtIssuer, JwksUri)
		return &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
	}

	It("should produce one RA for a rule with one issuer and two paths", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}

		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		ruleJwt2 := GetRuleWithServiceFor(ImgApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt, ruleJwt2})
		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ra.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(1))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
	})

	It("should produce RA for a Rule without service, but service definition on ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt})
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
	})

	It("should produce RA with service from Rule, when service is configured on Rule and ApiRule level", func() {
		// given
		jwt := createIstioJwtAccessStrategy()
		ruleServiceName := "rule-scope-example-service"
		service := &gatewayv1beta1.Service{
			Name: &ruleServiceName,
			Port: &ServicePort,
		}
		client := GetEmptyFakeClient()
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})

		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)
		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ruleServiceName))
	})

	It("should produce RA from a rule with two issuers and one path", func() {
		jwtConfigJSON := fmt.Sprintf(`{
			"authentications": [{"issuer": "%s", "jwksUri": "%s"}, {"issuer": "%s", "jwksUri": "%s"}]
			}`, JwtIssuer, JwksUri, JwtIssuer2, JwksUri2)
		jwt := &gatewayv1beta1.Authenticator{
			Handler: &gatewayv1beta1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		}
		client := GetEmptyFakeClient()
		service := &gatewayv1beta1.Service{
			Name: &ServiceName,
			Port: &ServicePort,
		}
		ruleJwt := GetRuleWithServiceFor(HeadersApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, []*gatewayv1beta1.Authenticator{jwt}, service)
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{ruleJwt})
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.ObjectMeta.Name).To(BeEmpty())
		Expect(ra.ObjectMeta.GenerateName).To(Equal(ApiName + "-"))
		Expect(ra.ObjectMeta.Namespace).To(Equal(ApiNamespace))
		Expect(ra.ObjectMeta.Labels[TestLabelKey]).To(Equal(TestLabelValue))

		Expect(len(ra.OwnerReferences)).To(Equal(1))
		Expect(ra.OwnerReferences[0].APIVersion).To(Equal(ApiAPIVersion))
		Expect(ra.OwnerReferences[0].Kind).To(Equal(ApiKind))
		Expect(ra.OwnerReferences[0].Name).To(Equal(ApiName))
		Expect(ra.OwnerReferences[0].UID).To(Equal(ApiUID))

		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels[TestSelectorKey]).To(Equal(ServiceName))
		Expect(len(ra.Spec.JwtRules)).To(Equal(2))
		Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(JwtIssuer))
		Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(JwksUri))
		Expect(ra.Spec.JwtRules[1].Issuer).To(Equal(JwtIssuer2))
		Expect(ra.Spec.JwtRules[1].JwksUri).To(Equal(JwksUri2))
	})

	It("should not create RA if handler is allow", func() {
		// given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "allow",
				},
			},
		}

		allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
		rules := []gatewayv1beta1.Rule{allowRule}

		apiRule := GetAPIRuleFor(rules)

		overrideServiceName := "testName"
		overrideServiceNamespace := "testName-namespace"
		overrideServicePort := uint32(8080)

		apiRule.Spec.Service = &gatewayv1beta1.Service{
			Name:      &overrideServiceName,
			Namespace: &overrideServiceNamespace,
			Port:      &overrideServicePort,
		}

		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})

	It("should not create RA if handler is noop", func() {
		// given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			},
		}

		allowRule := GetRuleFor(ApiPath, ApiMethods, []*gatewayv1beta1.Mutator{}, strategies)
		rules := []gatewayv1beta1.Rule{allowRule}

		apiRule := GetAPIRuleFor(rules)

		overrideServiceName := "testName"
		overrideServiceNamespace := "testName-namespace"
		overrideServicePort := uint32(8080)

		apiRule.Spec.Service = &gatewayv1beta1.Service{
			Name:      &overrideServiceName,
			Namespace: &overrideServiceNamespace,
			Port:      &overrideServicePort,
		}

		client := GetEmptyFakeClient()
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), client, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(BeEmpty())
	})

	It("should create RA when no exists", func() {
		// given
		jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "test-service")
		rules := []gatewayv1beta1.Rule{jwtRule}

		apiRule := GetAPIRuleFor(rules)
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), GetFakeClient(), apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})

	It("should delete RA when there is no rule configured in ApiRule", func() {
		// given
		apiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{})

		existingRa := securityv1beta1.RequestAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
				},
			},
			Spec: v1beta1.RequestAuthentication{
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test-service",
					},
				},
				JwtRules: []*v1beta1.JWTRule{
					{
						JwksUri: JwksUri,
						Issuer:  JwtIssuer,
					},
				},
			},
		}

		ctrlClient := GetFakeClient(&existingRa)
		processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("delete"))
	})

	When("RA with JWT config exists", func() {

		It("should update RA when nothing changed", func() {
			// given
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "test-service")
			rules := []gatewayv1beta1.Rule{jwtRule}

			apiRule := GetAPIRuleFor(rules)

			existingRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Action.String()).To(Equal("update"))
		})

		It("should delete and create new RA when only service name in JWT Rule has changed", func() {
			// given
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "updated-service")
			rules := []gatewayv1beta1.Rule{jwtRule}
			apiRule := GetAPIRuleFor(rules)

			existingRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "old-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("old-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			createMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("updated-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(deleteMatcher, createMatcher))
		})

		It("should create new RA when new service with new JWT config is added to ApiRule", func() {
			// given
			existingJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "existing-service")
			newJwtRule := GetJwtRuleWithService("https://new.issuer.com/", JwksUri, "new-service")

			rules := []gatewayv1beta1.Rule{existingJwtRule, newJwtRule}

			apiRule := GetAPIRuleFor(rules)

			existingRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "existing-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			updateResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("existing-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			createResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("new-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal("https://new.issuer.com/"),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(createResultMatcher, updateResultMatcher))
		})

		It("should create new RA and delete old RA when JWT ApiRule has new JWKS URI", func() {
			// given
			jwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri2, "test-service")
			rules := []gatewayv1beta1.Rule{jwtRule}

			apiRule := GetAPIRuleFor(rules)

			existingRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "test-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			createResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri2),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			deleteResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("test-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(createResultMatcher, deleteResultMatcher))
		})
	})

	When("Two RA with same JWT config for different services exist", func() {

		It("should update RAs and create new RA for first-service and delete old RA when JWT issuer in JWT Rule for first-service has changed", func() {
			// given
			firstJwtRule := GetJwtRuleWithService("https://new.issuer.com/", JwksUri, "first-service")
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")

			rules := []gatewayv1beta1.Rule{firstJwtRule, secondJwtRule}

			apiRule := GetAPIRuleFor(rules)

			existingFirstServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "firstRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "first-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			existingSecondServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "secondRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "second-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&existingFirstServiceRa, &existingSecondServiceRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			createFirstServiceRaResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("first-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal("https://new.issuer.com/"),
							})),
						),
					}),
				})),
			}))

			deleteFirstServiceRaResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("first-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			secondRaResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("second-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(createFirstServiceRaResultMatcher, deleteFirstServiceRaResultMatcher, secondRaResultMatcher))
		})

		It("should delete only first-service RA when it was removed from ApiRule", func() {
			// given
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")

			rules := []gatewayv1beta1.Rule{secondJwtRule}

			apiRule := GetAPIRuleFor(rules)

			firstServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "firstRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "first-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			secondServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "secondRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "second-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&firstServiceRa, &secondServiceRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))

			deleteResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("delete")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("first-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			updateResultMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("second-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(deleteResultMatcher, updateResultMatcher))
		})

		It("should create new RA when it has different service", func() {
			// given
			firstJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "first-service")
			secondJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "second-service")
			newJwtRule := GetJwtRuleWithService(JwtIssuer, JwksUri, "new-service")

			rules := []gatewayv1beta1.Rule{firstJwtRule, secondJwtRule, newJwtRule}

			apiRule := GetAPIRuleFor(rules)

			firstServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "firstRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "first-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			secondServiceRa := securityv1beta1.RequestAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name: "secondRa",
					Labels: map[string]string{
						processing.OwnerLabelv1alpha1: fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace),
					},
				},
				Spec: v1beta1.RequestAuthentication{
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "second-service",
						},
					},
					JwtRules: []*v1beta1.JWTRule{
						{
							JwksUri: JwksUri,
							Issuer:  JwtIssuer,
						},
					},
				},
			}

			ctrlClient := GetFakeClient(&firstServiceRa, &secondServiceRa)
			processor := istio.NewRequestAuthenticationProcessor(GetTestConfig())

			// when
			result, err := processor.EvaluateReconciliation(context.TODO(), ctrlClient, apiRule)

			// then
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))

			firstRaMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("first-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			secondRaMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("update")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("second-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			newRaMatcher := PointTo(MatchFields(IgnoreExtras, Fields{
				"Action": WithTransform(ActionToString, Equal("create")),
				"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
					"Spec": MatchFields(IgnoreExtras, Fields{
						"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
							"MatchLabels": ContainElement("new-service"),
						})),
						"JwtRules": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"JwksUri": Equal(JwksUri),
								"Issuer":  Equal(JwtIssuer),
							})),
						),
					}),
				})),
			}))

			Expect(result).To(ContainElements(firstRaMatcher, secondRaMatcher, newRaMatcher))
		})
	})
})
