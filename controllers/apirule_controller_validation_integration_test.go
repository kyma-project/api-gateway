package controllers_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Apirule controller validation", func() {

	const (
		testNameBase           = "status-test"
		testIDLength           = 5
		testServiceName        = "httpbin"
		testServicePort uint32 = 443
		testPath               = "/.*"
	)

	Context("with istio handler", func() {

		testJwtHandlerConfigError := func(accessStrategies []*gatewayv1beta1.Authenticator, expectedValidationErrors []string) {
			// given
			setHandlerConfigMap(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := gatewayv1beta1.Rule{
				Path:             testPath,
				Methods:          defaultMethods,
				Mutators:         defaultMutators,
				AccessStrategies: accessStrategies,
			}
			instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

			err := c.Create(context.TODO(), instance)
			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := c.Delete(context.TODO(), instance)
				Expect(err).NotTo(HaveOccurred())
			}()

			// when
			expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

			// then
			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))

				for _, expected := range expectedValidationErrors {
					g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring(expected))
				}
			}, eventuallyTimeout).Should(Succeed())
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
	})

})
