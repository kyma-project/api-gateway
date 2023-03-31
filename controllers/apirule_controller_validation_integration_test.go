package controllers_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("Apirule controller validation", Serial, func() {

	const (
		testNameBase           = "status-test"
		testIDLength           = 5
		testServiceName        = "httpbin"
		testServicePort uint32 = 443
		testPath               = "/.*"
	)

	Context("with istio handler", func() {

		testConfigError := func(accessStrategies []*gatewayv1beta1.Authenticator, mutators []*gatewayv1beta1.Mutator, expectedValidationErrors []string) {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := gatewayv1beta1.Rule{
				Path:             testPath,
				Methods:          defaultMethods,
				Mutators:         mutators,
				AccessStrategies: accessStrategies,
			}
			instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			// when
			Expect(c.Create(context.TODO(), instance)).Should(Succeed())
			defer func() {
				deleteApiRule(instance)
			}()

			// then
			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())
				g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
				for _, expected := range expectedValidationErrors {
					g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring(expected))
				}
			}, eventuallyTimeout).Should(Succeed())
		}

		testMutatorConfigError := func(mutator *gatewayv1beta1.Mutator, expectedValidationErrors []string) {
			a := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
					},
				},
			}

			mutators := []*gatewayv1beta1.Mutator{mutator}
			testConfigError(a, mutators, expectedValidationErrors)
		}

		testJwtHandlerConfigError := func(accessStrategies []*gatewayv1beta1.Authenticator, expectedValidationErrors []string) {
			testConfigError(accessStrategies, []*gatewayv1beta1.Mutator{}, expectedValidationErrors)
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

			testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
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
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is empty or not a valid url err=value is empty",
			}

			testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
		})

		It("should not allow creation of APIRule with insecure issuer and jwks in jwt config", func() {
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

			expectedValidationErrors := []string{
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].issuer\": value is not a secured url err=value is unsecure",
				"Attribute \".spec.rules[0].accessStrategies[0].config.authentications[0].jwksUri\": value is not a secured url err=value is unsecure",
			}

			testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
		})

		It("should not allow creation of APIRule with invalid url issuer and jwks in jwt config", func() {
			accessStrategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
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

			testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
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

				testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
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

				testJwtHandlerConfigError(accessStrategies, expectedValidationErrors)
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
