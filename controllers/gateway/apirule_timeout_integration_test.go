package gateway_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("APIRule timeout", Serial, func() {

	const (
		testNameBase               = "test"
		testIDLength               = 5
		testServiceNameBase        = "httpbin"
		testServicePort     uint32 = 443
		testIssuer                 = "https://oauth2.example.com/"
		testJwksUri                = "https://oauth2.example.com/.well-known/jwks.json"
	)

	var methodsGet = []gatewayv1beta1.HttpMethod{http.MethodGet}

	Context("when creating an APIRule", func() {

		Context("without timeout", func() {

			defaultTimeout := time.Second * 180

			testFunction := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
					g.Expect(vsList.Items).To(HaveLen(1))

					vs := vsList.Items[0]
					g.Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(defaultTimeout))
				}, eventuallyTimeout).Should(Succeed())
			}

			Context("with Ory JWT handler", func() {
				It("should create a virtual service with default timeout", func() {
					updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
					jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
					testFunction(jwtHandler)

				})
			})

			Context("with Istio JWT handler", func() {
				It("should create a virtual service with default timeout", func() {
					updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
					jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
					testFunction(jwtHandler)

				})
			})
		})

		Context("with 40s timeout", func() {

			var timeout gatewayv1beta1.Timeout = 40

			testTimeoutOnRootLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
				apiRule.Spec.Timeout = &timeout

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
					g.Expect(vsList.Items).To(HaveLen(1))

					vs := vsList.Items[0]
					g.Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(40 * time.Second))
				}, eventuallyTimeout).Should(Succeed())
			}
			testTimeoutOnRuleLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)
				rule.Timeout = &timeout
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.Background(), &vsList, matchingLabels)).Should(Succeed())
					g.Expect(vsList.Items).To(HaveLen(1))

					vs := vsList.Items[0]
					g.Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(40 * time.Second))
				}, eventuallyTimeout).Should(Succeed())
			}

			Context("with Ory JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should create a virtual service with timeout of 5m", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should create a virtual service with timeout of 5m", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})

			Context("with Istio JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should create a virtual service with timeout of 5m", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should create a virtual service with timeout of 5m", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})
		})

		Context("with 4800s timeout", func() {

			var timeout gatewayv1beta1.Timeout = 4800

			testTimeoutOnRootLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
				apiRule.Spec.Timeout = &timeout

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.timeout: Invalid value: 4800: spec.timeout in body should be less than or equal to 3900"))
			}
			testTimeoutOnRuleLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)
				rule.Timeout = &timeout
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.rules[0].timeout: Invalid value: 4800: spec.rules[0].timeout in body should be less than or equal to 3900"))
			}

			Context("with Ory JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})

			Context("with Istio JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})
		})

		Context("with 0s timeout", func() {

			var timeout gatewayv1beta1.Timeout = 0

			testTimeoutOnRootLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
				apiRule.Spec.Timeout = &timeout

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.timeout: Invalid value: 0: spec.timeout in body should be greater than or equal to 1"))
			}
			testTimeoutOnRuleLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", methodsGet, nil, jwtHandler)
				rule.Timeout = &timeout
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.rules[0].timeout: Invalid value: 0: spec.rules[0].timeout in body should be greater than or equal to 1"))
			}

			Context("with Ory JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)
						jwtHandler := testOryJWTHandler(testIssuer, defaultScopes)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})

			Context("with Istio JWT handler", func() {
				Context("on APIRule root level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRootLevel(jwtHandler)

					})
				})
				Context("on rule level", func() {
					It("should set Status to Error and have validation error in ApiRule", func() {
						updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)
						jwtHandler := testIstioJWTHandler(testIssuer, testJwksUri)
						testTimeoutOnRuleLevel(jwtHandler)
					})

				})
			})
		})
	})
})
