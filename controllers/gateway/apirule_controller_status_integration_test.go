package gateway_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("Resource status", Serial, func() {

	const (
		testNameBase           = "status-test"
		testIDLength           = 5
		testServiceName        = "httpbin"
		testServicePort uint32 = 443
		testPath               = "/.*"
	)

	Context("with ory handler", func() {

		It("should return nil for resources not supported by the handler ", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := testRule(testPath, defaultMethods, defaultMutators, noConfigHandler("noop"))
			instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(instance)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), instance)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				g.Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())
				g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusWarning))
			}, eventuallyTimeout).Should(Succeed())

		})

		It("should report validation errors in ApiRule status", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ORY)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			invalidConfig := testOauthHandler(defaultScopes)
			invalidConfig.Name = "noop"

			rule := testRule(testPath, defaultMethods, defaultMutators, invalidConfig)
			instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(instance)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), instance)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())

				g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path and method"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

				shouldHaveVirtualServices(g, apiRuleName, testNamespace, 0)
			}, eventuallyTimeout).Should(Succeed())

		})
	})

	Context("with istio handler", func() {

		It("should return nil for resources not supported by the handler ", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := testRule(testPath, defaultMethods, []*gatewayv1beta1.Mutator{}, noConfigHandler("noop"))
			instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(instance)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), instance)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())
				g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusWarning))
			}, eventuallyTimeout).Should(Succeed())
		})

		It("should report validation errors in ApiRule status", func() {
			// given
			updateJwtHandlerTo(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			invalidConfig := testOauthHandler(defaultScopes)
			invalidConfig.Name = "noop"

			rule := testRule(testPath, defaultMethods, []*gatewayv1beta1.Mutator{}, invalidConfig)
			instance := testApiRule(apiRuleName, testNamespace, serviceName, testNamespace, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			svc := testService(serviceName, testNamespace, testServicePort)
			defer func() {
				deleteResource(instance)
				deleteResource(svc)
			}()

			// when
			Expect(c.Create(context.Background(), svc)).Should(Succeed())
			Expect(c.Create(context.Background(), instance)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				created := gatewayv1beta1.APIRule{}
				Expect(c.Get(context.Background(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)).Should(Succeed())
				g.Expect(created.Status.APIRuleStatus).NotTo(BeNil())
				g.Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path and method"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
				g.Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

				shouldHaveVirtualServices(g, apiRuleName, testNamespace, 0)
			}, eventuallyTimeout).Should(Succeed())
		})

	})
})
