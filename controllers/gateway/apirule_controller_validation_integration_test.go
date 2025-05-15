package gateway_test

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRawConfig(config any) *runtime.RawExtension {
	b, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: b,
	}
}

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("Apirule controller validation", Serial, Ordered, func() {

	Context("with Ory handler", func() {

		BeforeAll(func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
		})

	})

	Context("with istio handler", func() {

		BeforeAll(func() {
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
		})

		testMutatorConfigError := func(mutator *gatewayv1beta1.Mutator, expectedValidationErrors []string) {
			a := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			mutators := []*gatewayv1beta1.Mutator{mutator}
			testConfig(a, mutators, gatewayv1beta1.StatusError, expectedValidationErrors)
		}

		testJwtHandlerConfig := func(accessStrategies []*gatewayv1beta1.Authenticator, expectedStatusCode gatewayv1beta1.StatusCode, expectedValidationErrors []string) {
			testConfig(accessStrategies, []*gatewayv1beta1.Mutator{}, expectedStatusCode, expectedValidationErrors)
		}

		It("should not allow creation of APIRule without config in jwt handler", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			expectedValidationErrors := []string{
				"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": supplied config cannot be empty",
			}

			testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
		})

		It("should not allow creation of APIRule jwks configuration in jwt handler", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer: "https://example.com/",
									},
								},
							}),
					},
				},
			}

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is empty or not a valid uri err=value is empty",
			}

			testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
		})

		It("should allow creation of APIRule with insecure issuer and jwks in jwt config", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
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

			testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusOK, expectedValidationErrors)
		})

		It("should not allow creation of APIRule with jwt and noop handler", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:  "http://example.com/",
										JwksUri: "http://example.com/.well-known/jwks.json",
									},
								},
							}),
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "noop",
					},
				},
			}

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies.accessStrategies[0].handler\": jwt access strategy is not allowed in combination with other access strategies",
				"Attribute \".spec.rules[0].accessStrategies.accessStrategies[1].handler\": noop access strategy is not allowed in combination with other access strategies",
				"Attribute \".spec.rules[0].accessStrategies\": Secure access strategies cannot be used in combination with unsecure access strategies",
			}

			testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
		})

		It("should not allow creation of APIRule with invalid uri for issuer and jwks in jwt config", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:  "invalid_:example",
										JwksUri: "example.com/.well-known/jwks.json",
									},
								},
							}),
					},
				},
			}

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].issuer\": value is empty or not a valid uri err=parse \"invalid_:example\": invalid URI for request",
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is empty or not a valid uri err=parse \"example.com/.well-known/jwks.json\": invalid URI for request",
			}

			testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
		})

		Context("JWT config authorizations", func() {

			It("should not allow creation of APIRule with empty Audiences in jwt config authorizations", func() {
				accessStrategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "jwt",
							Config: getRawConfig(
								gatewayv1beta1.JwtConfig{
									Authentications: []*gatewayv1beta1.JwtAuthentication{
										{
											Issuer:  "https://example.com/",
											JwksUri: "https://example.com/.well-known/jwks.json",
										},
									},
									Authorizations: []*gatewayv1beta1.JwtAuthorization{
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

				testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with empty RequiredScopes in jwt config authorizations", func() {
				accessStrategies := []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "jwt",
							Config: getRawConfig(
								gatewayv1beta1.JwtConfig{
									Authentications: []*gatewayv1beta1.JwtAuthentication{
										{
											Issuer:  "https://example.com/",
											JwksUri: "https://example.com/.well-known/jwks.json",
										},
									},
									Authorizations: []*gatewayv1beta1.JwtAuthorization{
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

				testJwtHandlerConfig(accessStrategies, gatewayv1beta1.StatusError, expectedValidationErrors)
			})
		})

		Context("mutators", func() {

			It("should not allow creation of APIRule with id_token mutator", func() {
				mutator := &gatewayv1beta1.Mutator{
					Handler: &gatewayv1beta1.Handler{
						Name: "id_token",
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": unsupported mutator: id_token",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with hydrator mutator", func() {
				mutator := &gatewayv1beta1.Mutator{
					Handler: &gatewayv1beta1.Handler{
						Name: "hydrator",
					},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": unsupported mutator: hydrator",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule without handler in mutator", func() {
				mutator := &gatewayv1beta1.Mutator{
					Handler: &gatewayv1beta1.Handler{},
				}

				expectedValidationErrors := []string{
					"Attribute \".spec.rules[0].mutators[0].handler\": mutator handler cannot be empty",
				}

				testMutatorConfigError(mutator, expectedValidationErrors)
			})

			It("should not allow creation of APIRule with header mutator without headers", func() {
				mutator := &gatewayv1beta1.Mutator{
					Handler: &gatewayv1beta1.Handler{
						Name: "header",
						Config: getRawConfig(
							gatewayv1beta1.HeaderMutatorConfig{
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
				mutator := &gatewayv1beta1.Mutator{
					Handler: &gatewayv1beta1.Handler{
						Name: "cookie",
						Config: getRawConfig(
							gatewayv1beta1.CookieMutatorConfig{
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

func testConfig(accessStrategies []*gatewayv1beta1.Authenticator, mutators []*gatewayv1beta1.Mutator, expectedStatusCode gatewayv1beta1.StatusCode, expectedValidationErrors []string) {
	// given
	serviceName := generateTestName("httpbin", 5)
	serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)
	testConfigWithServiceAndHost(serviceName, serviceHost, accessStrategies, mutators, expectedStatusCode, expectedValidationErrors)
}

func testConfigWithServiceAndHost(serviceName string, host string, accessStrategies []*gatewayv1beta1.Authenticator, mutators []*gatewayv1beta1.Mutator, expectedStatusCode gatewayv1beta1.StatusCode, expectedValidationErrors []string) {
	// given
	const (
		testNameBase           = "status-test"
		testIDLength           = 5
		testServicePort uint32 = 443
		testPath               = "/.*"
	)

	apiRuleName := generateTestName(testNameBase, testIDLength)

	rule := gatewayv1beta1.Rule{
		Path:             testPath,
		Methods:          defaultMethods,
		Mutators:         mutators,
		AccessStrategies: accessStrategies,
	}
	instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, host, testServicePort, []gatewayv1beta1.Rule{rule})
	svc := testService(serviceName, testNamespace, testServicePort)

	// when
	Expect(c.Create(context.Background(), svc)).Should(Succeed())
	Expect(c.Create(context.Background(), instance)).Should(Succeed())
	defer func() {
		deleteResource(instance)
		deleteResource(svc)
	}()

	// then
	Eventually(func(g Gomega) {
		created := gatewayv1beta1.APIRule{}
		g.Expect(func() error {
			err := c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			return err
		}()).Should(Succeed())

		g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
		g.Expect(created.Status.APIRuleStatus.Code).To(Equal(expectedStatusCode))
		for _, expected := range expectedValidationErrors {
			g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring(expected))
		}
	}, eventuallyTimeout).Should(Succeed())
}
