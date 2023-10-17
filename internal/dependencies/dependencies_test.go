package dependencies_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Check", func() {
	Context("APIRule dependencies", func() {
		It("Should fail if required CRDs are missing", func() {
			k8sClient := createFakeClient()
			status := dependencies.NewAPIRule().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeFalse())
			Expect(status.Description()).To(Equal("CRD virtualservices.networking.istio.io is not present. Make sure to install required dependencies for the component"))
		})

		It("Should not fail if required CRDs are present", func() {

			crds := []v1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "virtualservices.networking.istio.io",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rules.oathkeeper.ory.sh",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "authorizationpolicies.security.istio.io",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "requestauthentications.security.istio.io",
					},
				},
			}

			k8sClient := createFakeClient()

			for _, crd := range crds {
				Expect(k8sClient.Create(context.Background(), &crd)).To(Succeed())
			}
			status := dependencies.NewAPIRule().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeTrue())
		})
	})
	Context("APIGateway dependencies", func() {
		It("Should fail if required CRDs are missing", func() {
			k8sClient := createFakeClient()
			status := dependencies.NewAPIGateway().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeFalse())
			Expect(status.Description()).To(Equal("CRD gateways.networking.istio.io is not present. Make sure to install required dependencies for the component"))
		})

		It("Should not fail if required CRDs are present", func() {

			crds := []v1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.networking.istio.io",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "virtualservices.networking.istio.io",
					},
				},
			}

			k8sClient := createFakeClient()

			for _, crd := range crds {
				Expect(k8sClient.Create(context.Background(), &crd)).To(Succeed())
			}
			status := dependencies.NewAPIGateway().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeTrue())
		})
	})
	Context("APIGateway Gardener dependencies", func() {
		It("Should fail if required CRDs are missing", func() {
			k8sClient := createFakeClient()
			status := dependencies.NewGardenerAPIGateway().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeFalse())
			Expect(status.Description()).To(Equal("CRD gateways.networking.istio.io is not present. Make sure to install required dependencies for the component"))
		})

		It("Should not fail if required CRDs are present", func() {

			crds := []v1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.networking.istio.io",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "virtualservices.networking.istio.io",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dnsentries.dns.gardener.cloud",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "certificates.cert.gardener.cloud",
					},
				},
			}

			k8sClient := createFakeClient()

			for _, crd := range crds {
				Expect(k8sClient.Create(context.Background(), &crd)).To(Succeed())
			}
			status := dependencies.NewGardenerAPIGateway().Check(context.Background(), k8sClient)
			Expect(status.IsReady()).To(BeTrue())
		})
	})
})
