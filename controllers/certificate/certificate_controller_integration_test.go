package certificate

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"

	. "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Certificate Controller", Serial, func() {
	// AfterEach(func() {
	// 	deleteApiRules()
	// 	deleteApiGateways()
	// })

	Context("Secret", func() {
		It("Should generate new valid certificate when current is not valid", func() {
			// given
			secret := getSecret([]byte("abc"), []byte("xyz"))

			// when
			Expect(k8sClient.Create(context.Background(), secret)).Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				secret := corev1.Secret{}
				g.Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: secretName, Namespace: secretNamespace}, &secret)).Should(Succeed())

				_, err := cert.ParseCertsPEM(secret.Data[certificateName])
				g.Expect(err).To(BeNil())

				b, _ := pem.Decode(secret.Data[keyName])
				_, err = x509.ParsePKCS1PrivateKey(b.Bytes)
				g.Expect(err).To(BeNil())
			}, eventuallyTimeout).Should(Succeed())
		})
	})
})

// func deleteApiGateways() {
// 	Eventually(func(g Gomega) {
// 		By("Checking if APIGateway exists as part of teardown")
// 		list := v1alpha1.APIGatewayList{}
// 		Expect(k8sClient.List(context.TODO(), &list)).Should(Succeed())

// 		for _, item := range list.Items {
// 			apiGatewayTeardown(&item)
// 		}
// 	}, eventuallyTimeout).Should(Succeed())
// }

// func apiGatewayTeardown(apiGateway *v1alpha1.APIGateway) {
// 	By(fmt.Sprintf("Deleting APIGateway %s as part of teardown", apiGateway.Name))
// 	Eventually(func(g Gomega) {
// 		err := k8sClient.Delete(context.TODO(), apiGateway)

// 		if err != nil {
// 			Expect(errors.IsNotFound(err)).To(BeTrue())
// 		}

// 		a := v1alpha1.APIGateway{}
// 		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: apiGateway.Name}, &a)
// 		g.Expect(errors.IsNotFound(err)).To(BeTrue())
// 	}, eventuallyTimeout).Should(Succeed())
// }

// func deleteApiRules() {
// 	Eventually(func(g Gomega) {
// 		By("Checking if APIRules exists as part of teardown")
// 		list := v1beta1.APIRuleList{}
// 		Expect(k8sClient.List(context.TODO(), &list)).Should(Succeed())

// 		for _, item := range list.Items {
// 			apiRuleTeardown(&item)
// 		}
// 	}, eventuallyTimeout).Should(Succeed())
// }

// func apiRuleTeardown(apiRule *v1beta1.APIRule) {
// 	By(fmt.Sprintf("Deleting APIRule %s as part of teardown", apiRule.Name))
// 	err := k8sClient.Delete(context.TODO(), apiRule)

// 	if err != nil {
// 		Expect(errors.IsNotFound(err)).To(BeTrue())
// 	}

// 	Eventually(func(g Gomega) {
// 		a := v1beta1.APIRule{}
// 		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: apiRule.Name, Namespace: apiRule.Namespace}, &a)
// 		g.Expect(errors.IsNotFound(err)).To(BeTrue())
// 	}, eventuallyTimeout).Should(Succeed())
// }

// func getApiRuleV1Beta1() gatewayv1beta1.APIRule {
// 	var servicePort uint32 = 8080

// 	return gatewayv1beta1.APIRule{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:       "test-api-rule-v1beta1",
// 			Namespace:  "default",
// 			Generation: 1,
// 		},
// 		Spec: gatewayv1beta1.APIRuleSpec{
// 			Host: ptr.To("test-host"),
// 			Service: &gatewayv1beta1.Service{
// 				Name: ptr.To("test-service"),
// 				Port: &servicePort,
// 			},
// 			Gateway: ptr.To(gateway.KymaGatewayFullName),
// 			Rules: []gatewayv1beta1.Rule{
// 				{
// 					Path:    "/.*",
// 					Methods: []gatewayv1beta1.HttpMethod{"GET"},
// 					AccessStrategies: []*gatewayv1beta1.Authenticator{
// 						{
// 							Handler: &gatewayv1beta1.Handler{
// 								Name: "no_auth",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// func getApiRuleV1Beta2() gatewayv1beta2.APIRule {
// 	var servicePort uint32 = 8080

// 	return gatewayv1beta2.APIRule{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:       "test-api-rule-v1beta2",
// 			Namespace:  "default",
// 			Generation: 1,
// 		},
// 		Spec: gatewayv1beta2.APIRuleSpec{
// 			Hosts: []*string{ptr.To("test-host")},
// 			Service: &gatewayv1beta2.Service{
// 				Name: ptr.To("test-service"),
// 				Port: &servicePort,
// 			},
// 			Gateway: ptr.To(gateway.KymaGatewayFullName),
// 			Rules: []gatewayv1beta2.Rule{
// 				{
// 					Path:    "/.*",
// 					Methods: []gatewayv1beta2.HttpMethod{"GET"},
// 					NoAuth:  ptr.To(true),
// 				},
// 			},
// 		},
// 	}
// }
