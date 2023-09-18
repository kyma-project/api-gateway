package gateway

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Kyma gateway", func() {

	It("should not create gateway when Spec doesn't contain EnableKymaGateway flag", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		k8sClient := createFakeClient(&apiGateway)

		// when
		err := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(err).ShouldNot(HaveOccurred())

		created := v1alpha3.Gateway{}
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should not create gateway when EnableKymaGateway is false", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(false),
			},
		}

		k8sClient := createFakeClient(&apiGateway)

		// when
		err := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(err).ShouldNot(HaveOccurred())

		created := v1alpha3.Gateway{}
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("should create gateway with *.local.kyma.dev hosts when EnableKymaGateway is true and no Gardener shoot-info exists", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(true),
			},
		}

		k8sClient := createFakeClient(&apiGateway)

		// when
		err := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(err).ShouldNot(HaveOccurred())

		created := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)).Should(Succeed())

		for _, server := range created.Spec.GetServers() {
			Expect(server.Hosts).To(ContainElement("*.local.kyma.dev"))
		}
	})

	It("should create gateway with hosts from shoot-info domain when EnableKymaGateway is true and Gardener shoot-info exists", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(true),
			},
		}

		cm := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shoot-info",
				Namespace: "kube-system",
			},
			Data: map[string]string{
				"domain": "some.gardener.domain",
			},
		}

		k8sClient := createFakeClient(&apiGateway, &cm)

		// when
		err := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(err).ShouldNot(HaveOccurred())

		created := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)).Should(Succeed())

		for _, server := range created.Spec.GetServers() {
			Expect(server.Hosts).To(ContainElement("*.some.gardener.domain"))
		}
	})

	It("should apply disclaimer annotation on Kyma gateway when it was removed", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(true),
			},
		}

		k8sClient := createFakeClient(&apiGateway)
		Expect(Reconcile(context.TODO(), k8sClient, apiGateway)).Should(Succeed())

		By("removing disclaimer annotation from Kyma gateway")
		kymaGateway := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())
		kymaGateway.Annotations = nil
		Expect(k8sClient.Update(context.TODO(), &kymaGateway)).Should(Succeed())

		// when
		err := Reconcile(context.TODO(), k8sClient, apiGateway)

		// then
		Expect(err).ShouldNot(HaveOccurred())

		created := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &created)).Should(Succeed())

		Expect(created.Annotations).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
			"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))
	})

	It("should delete Kyma gateway when EnableKymaGateway is updated to false", func() {
		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(false),
			},
		}
		testShouldDeleteKymaGateway(updatedApiGateway)
	})

	It("should delete Kyma gateway when EnableKymaGateway is removed in updated APIGateway", func() {
		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		testShouldDeleteKymaGateway(updatedApiGateway)
	})

	It("should not delete Kyma Gateway when EnableKymaGateway is updated to false, but any APIRule exists", func() {
		// given
		apiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.APIGatewaySpec{
				EnableKymaGateway: ptr.To(true),
			},
		}

		apiRule := v1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		k8sClient := createFakeClient(&apiGateway, &apiRule)
		err := Reconcile(context.TODO(), k8sClient, apiGateway)
		kymaGateway := v1alpha3.Gateway{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

		updatedApiGateway := v1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		// when
		err = Reconcile(context.TODO(), k8sClient, updatedApiGateway)

		// then
		err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
		Expect(err).ShouldNot(HaveOccurred())
	})

})

func testShouldDeleteKymaGateway(updatedApiGateway v1alpha1.APIGateway) {
	// given
	apiGateway := v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: ptr.To(true),
		},
	}

	k8sClient := createFakeClient(&apiGateway)
	err := Reconcile(context.TODO(), k8sClient, apiGateway)
	kymaGateway := v1alpha3.Gateway{}
	Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)).Should(Succeed())

	// when
	err = Reconcile(context.TODO(), k8sClient, updatedApiGateway)

	// then
	err = k8sClient.Get(context.TODO(), client.ObjectKey{Name: kymaGatewayName, Namespace: kymaGatewayNamespace}, &kymaGateway)
	Expect(errors.IsNotFound(err)).To(BeTrue())
}
