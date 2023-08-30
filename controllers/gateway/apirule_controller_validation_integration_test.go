package gateway_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("Apirule controller validation", Serial, Ordered, func() {

	Context("with Ory handler", func() {

		BeforeAll(func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
		})

		It("should not allow creation of APIRule with blocklisted subdomain api", func() {
			testHostInBlockList()
		})

	})

	Context("with istio handler", func() {

		BeforeAll(func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
		})

		testMutatorConfigError := func(mutator *v1beta1.Mutator, expectedValidationErrors []string) {
			a := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			mutators := []*v1beta1.Mutator{mutator}
			testConfig(a, mutators, v1beta1.StatusError, expectedValidationErrors)
		}

		testJwtHandlerConfig := func(accessStrategies []*v1beta1.Authenticator, expectedStatusCode v1beta1.StatusCode, expectedValidationErrors []string) {
			testConfig(accessStrategies, []*v1beta1.Mutator{}, expectedStatusCode, expectedValidationErrors)
		}

		It("should not allow creation of APIRule with blocklisted subdomain api", func() {
			testHostInBlockList()
		})

		It("should not allow creation of APIRule without config in jwt handler", func() {
			accessStrategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			expectedValidationErrors := []string{
				"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": supplied config cannot be empty",
			}

			testJwtHandlerConfig(accessStrategies, v1beta1.StatusError, expectedValidationErrors)
		})

		It("should not allow creation of APIRule jwks configuration in jwt handler", func() {
			accessStrategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							v1beta1.JwtConfig{
								Authentications: []*v1beta1.JwtAuthentication{
									{
										Issuer: "https://example.com/",
									},
								},
							}),
					},
				},
			}

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is empty or not a valid url err=value is empty",
			}

			testJwtHandlerConfig(accessStrategies, v1beta1.StatusError, expectedValidationErrors)
		})

		It("should allow creation of APIRule with insecure issuer and jwks in jwt config", func() {
			accessStrategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							v1beta1.JwtConfig{
								Authentications: []*v1beta1.JwtAuthentication{
									{
										Issuer:  "http://example.com/",
										JwksUri: "http://example.com/.well-known/jwks.json",
									},
								},
							}),
					},
				},
			}

			expectedValidationErrors := []string{}

			testJwtHandlerConfig(accessStrategies, v1beta1.StatusOK, expectedValidationErrors)
		})

		It("should not allow creation of APIRule with invalid url issuer and jwks in jwt config", func() {
			accessStrategies := []*v1beta1.Authenticator{
				{
					Handler: &v1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							v1beta1.JwtConfig{
								Authentications: []*v1beta1.JwtAuthentication{
									{
										Issuer:  "example.com/",
										JwksUri: "example.com/.well-known/jwks.json",
									},
								},
							}),
					},
				},
			}

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].issuer\": value is empty or not a valid url err=parse \"example.com/\": invalid URI for request",
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is empty or not a valid url err=parse \"example.com/.well-known/jwks.json\": invalid URI for request",
			}

			testJwtHandlerConfig(accessStrategies, v1beta1.StatusError, expectedValidationErrors)
		})

		Context("JWT config authorizations", func() {

			It("should not allow creation of APIRule with empty Audiences in jwt config authorizations", func() {
				accessStrategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: "jwt",
							Config: getRawConfig(
								v1beta1.JwtConfig{
									Authentications: []*v1beta1.JwtAuthentication{
										{
											Issuer:  "https://example.com/",
											JwksUri: "https://example.com/.well-known/jwks.json",
										},
									},
									Authorizations: []*v1beta1.JwtAuthorization{
										{
											Audiences: []string{},
										},
									},
								}),
						},
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].accessStrategies[0].config.authorizations[0].audiences\": value is empty",
				}

				testJwtHandlerConfig(accessStrategies, v1beta1.StatusError, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with empty RequiredScopes in jwt config authorizations", func() {
				accessStrategies := []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: "jwt",
							Config: getRawConfig(
								v1beta1.JwtConfig{
									Authentications: []*v1beta1.JwtAuthentication{
										{
											Issuer:  "https://example.com/",
											JwksUri: "https://example.com/.well-known/jwks.json",
										},
									},
									Authorizations: []*v1beta1.JwtAuthorization{
										{
											RequiredScopes: []string{},
										},
									},
								}),
						},
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].accessStrategies[0].config.authorizations[0].requiredScopes\": value is empty",
				}

				testJwtHandlerConfig(accessStrategies, v1beta1.StatusError, expectedValidationErrors)
			})
		})

		Context("mutators", func() {

			It("should not allow creation of APIRule with id_token mutator", func() {
				mutator := &v1beta1.Mutator{
					Handler: &v1beta1.Handler{
						Name: "id_token",
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": unsupported mutator: id_token",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with hydrator mutator", func() {
				mutator := &v1beta1.Mutator{
					Handler: &v1beta1.Handler{
						Name: "hydrator",
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": unsupported mutator: hydrator",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule without handler in mutator", func() {
				mutator := &v1beta1.Mutator{
					Handler: &v1beta1.Handler{},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": mutator handler cannot be empty",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with header mutator without headers", func() {
				mutator := &v1beta1.Mutator{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: getRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{},
							},
						),
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler.config\": headers cannot be empty",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with cookie mutator without cookies", func() {
				mutator := &v1beta1.Mutator{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: getRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{},
							},
						),
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler.config\": cookies cannot be empty",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

		})
	})

})

func testConfig(accessStrategies []*v1beta1.Authenticator, mutators []*v1beta1.Mutator, expectedStatusCode v1beta1.StatusCode, expectedValidationErrors []string) {
	// given
	serviceName := generateTestName("httpbin", 5)
	serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)
	testConfigWithServiceAndHost(serviceName, serviceHost, accessStrategies, mutators, expectedStatusCode, expectedValidationErrors)
}

func testConfigWithServiceAndHost(serviceName string, host string, accessStrategies []*v1beta1.Authenticator, mutators []*v1beta1.Mutator, expectedStatusCode v1beta1.StatusCode, expectedValidationErrors []string) {
	// given
	const (
		testNameBase           = "status-test"
		testIDLength           = 5
		testServicePort uint32 = 443
		testPath               = "/.*"
	)

	apiRuleName := generateTestName(testNameBase, testIDLength)

	rule := v1beta1.Rule{
		Path:             testPath,
		Methods:          defaultMethods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
	instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, host, testServicePort, []v1beta1.Rule{rule})
	svc := testService(serviceName, testNamespace, testServicePort)

	// when
	Expect(c.Create(context.TODO(), svc)).Should(Succeed())
	Expect(c.Create(context.TODO(), instance)).Should(Succeed())
	defer func() {
		apiRuleTeardown(instance)
		serviceTeardown(svc)
	}()

	// then
	Eventually(func(g Gomega) {
		created := v1beta1.APIRule{}
		g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())
		g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
		g.Expect(created.Status.APIRuleStatus.Code).To(Equal(expectedStatusCode))
		for _, expected := range expectedValidationErrors {
			g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring(expected))
		}
	}, eventuallyTimeout).Should(Succeed())
}

func testHostInBlockList() {
	accessStrategies := []*v1beta1.Authenticator{
		{
			Handler: &v1beta1.Handler{
				Name: "noop",
			},
		},
	}

	serviceName := generateTestName("httpbin", 5)

	expectedErrors := []string{"Validation error: Attribute \".spec.host\": The subdomain api is blocklisted for kyma.local domain"}
	testConfigWithServiceAndHost(serviceName, "api.kyma.local", accessStrategies, nil, v1beta1.StatusError, expectedErrors)
}
