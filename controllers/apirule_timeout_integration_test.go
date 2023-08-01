package controllers_test

import (
	"context"
	"fmt"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	Context("when creating an APIRule", func() {

		Context("without timeout", func() {

			defaultTimeout := time.Second * 180

			testFunction := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", []string{"GET"}, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.TODO(), svc)).Should(Succeed())
				Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.TODO(), &vsList, matchingLabels)).Should(Succeed())
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

		Context("with 5m timeout", func() {

			timeout := 5 * time.Minute

			testTimeoutOnRootLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", []string{"GET"}, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
				apiRule.Spec.Timeout = &metav1.Duration{Duration: timeout}

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.TODO(), svc)).Should(Succeed())
				Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.TODO(), &vsList, matchingLabels)).Should(Succeed())
					g.Expect(vsList.Items).To(HaveLen(1))

					vs := vsList.Items[0]
					g.Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(5 * time.Minute))
				}, eventuallyTimeout).Should(Succeed())
			}
			testTimeoutOnRuleLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", []string{"GET"}, nil, jwtHandler)
				rule.Timeout = &metav1.Duration{Duration: timeout}
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.TODO(), svc)).Should(Succeed())
				Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusOK)

				matchingLabels := matchingLabelsFunc(apiRuleName, testNamespace)

				By("Verifying created virtual service")
				vsList := networkingv1beta1.VirtualServiceList{}
				Eventually(func(g Gomega) {
					g.Expect(c.List(context.TODO(), &vsList, matchingLabels)).Should(Succeed())
					g.Expect(vsList.Items).To(HaveLen(1))

					vs := vsList.Items[0]
					g.Expect(vs.Spec.Http[0].Timeout.AsDuration()).To(Equal(timeout))
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

		Context("with 80m timeout", func() {

			timeout := 80 * time.Minute

			testTimeoutOnRootLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", []string{"GET"}, nil, jwtHandler)

				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
				apiRule.Spec.Timeout = &metav1.Duration{Duration: timeout}

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.TODO(), svc)).Should(Succeed())
				Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

				By("Verifying APIRule status description")
				Eventually(func(g Gomega) {
					expectedApiRule := gatewayv1beta1.APIRule{}
					g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
					g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
					g.Expect(expectedApiRule.Status.APIRuleStatus.Description).To(ContainSubstring("Validation error: Attribute \"spec.timeout\": Timeout must not exceed 65m"))
				}, eventuallyTimeout).Should(Succeed())
			}
			testTimeoutOnRuleLevel := func(jwtHandler *gatewayv1beta1.Handler) {
				rule := testRule("/img", []string{"GET"}, nil, jwtHandler)
				rule.Timeout = &metav1.Duration{Duration: timeout}
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := fmt.Sprintf("httpbin-%s.kyma.local", apiRuleName)

				apiRule := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})

				svc := testService(serviceName, testNamespace, testServicePort)

				// when
				Expect(c.Create(context.TODO(), svc)).Should(Succeed())
				Expect(c.Create(context.TODO(), apiRule)).Should(Succeed())
				defer func() {
					apiRuleTeardown(apiRule)
					serviceTeardown(svc)
				}()

				expectApiRuleStatus(apiRuleName, gatewayv1beta1.StatusError)

				By("Verifying APIRule status description")
				Eventually(func(g Gomega) {
					expectedApiRule := gatewayv1beta1.APIRule{}
					g.Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &expectedApiRule)).Should(Succeed())
					g.Expect(expectedApiRule.Status.APIRuleStatus).NotTo(BeNil())
					g.Expect(expectedApiRule.Status.APIRuleStatus.Description).To(ContainSubstring(".spec.rules[0].timeout\": Timeout must not exceed 65m"))
				}, eventuallyTimeout).Should(Succeed())
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
