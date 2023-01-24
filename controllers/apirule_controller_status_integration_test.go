package controllers_test

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/api-gateway/internal/helpers"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Resource status", func() {

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
			setHandlerConfigMap(helpers.JWT_HANDLER_ORY)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := testRule(testPath, defaultMethods, defaultMutators, noConfigHandler("noop"))
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

			created := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.AuthorizationPolicyStatus).To(BeNil())
			Expect(created.Status.RequestAuthenticationStatus).To(BeNil())
		})

		It("should report validation errors in ApiRule status", func() {
			// given
			setHandlerConfigMap(helpers.JWT_HANDLER_ORY)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			invalidConfig := testOauthHandler(defaultScopes)
			invalidConfig.Name = "noop"

			rule := testRule(testPath, defaultMethods, defaultMutators, invalidConfig)
			instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry

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

			created := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path and method"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

			vsList := networkingv1beta1.VirtualServiceList{}
			err = c.List(context.TODO(), &vsList, matchingLabelsFunc(apiRuleName, testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(vsList.Items).To(HaveLen(0))

			Expect(created.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
			Expect(created.Status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
			Expect(created.Status.AuthorizationPolicyStatus).To(BeNil())
			Expect(created.Status.RequestAuthenticationStatus).To(BeNil())
		})
	})

	Context("with istio handler", func() {

		It("should return nil for resources not supported by the handler ", func() {
			// given
			setHandlerConfigMap(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			rule := testRule(testPath, defaultMethods, defaultMutators, noConfigHandler("noop"))
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

			created := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			Expect(created.Status.AccessRuleStatus).To(BeNil())
		})

		It("should report validation errors in ApiRule status", func() {
			setHandlerConfigMap(helpers.JWT_HANDLER_ISTIO)

			apiRuleName := generateTestName(testNameBase, testIDLength)
			serviceName := generateTestName(testServiceName, testIDLength)
			serviceHost := fmt.Sprintf("%s.kyma.local", serviceName)

			invalidConfig := testOauthHandler(defaultScopes)
			invalidConfig.Name = "noop"

			rule := testRule(testPath, defaultMethods, defaultMutators, invalidConfig)
			instance := testInstance(apiRuleName, testNamespace, serviceName, serviceHost, testServicePort, []gatewayv1beta1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry

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

			expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: apiRuleName, Namespace: testNamespace}}

			Eventually(requests, eventuallyTimeout).Should(Receive(Equal(expectedRequest)))

			created := gatewayv1beta1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: apiRuleName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path and method"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

			vsList := networkingv1beta1.VirtualServiceList{}
			err = c.List(context.TODO(), &vsList, matchingLabelsFunc(apiRuleName, testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(vsList.Items).To(HaveLen(0))

			Expect(created.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
			Expect(created.Status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
			Expect(created.Status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
			Expect(created.Status.AccessRuleStatus).To(BeNil())
		})

	})
})
