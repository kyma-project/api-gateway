package gateway

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1beta12 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("discover external gateway", func() {
	const externalGatewayName = "external-gateway"
	const externalGatewayNamespace = "test-namespace"
	const generatedGatewayName = externalGatewayName + "-gateway"

	var _ = DescribeTable("should return gateway for ExternalGateway", func(
		apiRule *v2alpha1.APIRule,
		createExternalGateway bool,
		createGeneratedGateway bool,
		expectNilGateway bool,
		expectError bool,
	) {
		// given
		scheme := runtime.NewScheme()
		schemeBuilder := runtime.NewSchemeBuilder(
			v1beta1.AddToScheme,
			v2alpha1.AddToScheme,
			externalv1alpha1.AddToScheme,
		)
		Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())

		k8sClientBuilder := fake.NewClientBuilder().WithScheme(scheme)

		if createExternalGateway {
			externalGw := &externalv1alpha1.ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      externalGatewayName,
					Namespace: externalGatewayNamespace,
				},
				Spec: externalv1alpha1.ExternalGatewaySpec{
					ExternalDomain: "*.example.com",
					InternalDomain: externalv1alpha1.InternalDomainConfig{
						KymaSubdomain: "external",
					},
					BTPRegion: "eu10",
					CASecretRef: &corev1.SecretReference{
						Name: "ca-secret",
					},
				},
			}
			k8sClientBuilder.WithObjects(externalGw)
		}

		if createGeneratedGateway {
			generatedGw := &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generatedGatewayName,
					Namespace: externalGatewayNamespace,
				},
				Spec: v1beta12.Gateway{
					Servers: []*v1beta12.Server{
						{
							Hosts: []string{"*.example.com"},
							Port: &v1beta12.Port{
								Number:   443,
								Protocol: "HTTPS",
								Name:     "https",
							},
						},
					},
				},
			}
			k8sClientBuilder.WithObjects(generatedGw)
		}

		// when
		gotGateway, gotErr := discoverGateway(k8sClientBuilder.Build(), context.Background(), logr.Discard(), apiRule)

		// then
		if expectError {
			Expect(gotErr).ToNot(BeNil())
			return
		}

		Expect(gotErr).To(BeNil())

		if expectNilGateway {
			Expect(gotGateway).To(BeNil())
			Expect(apiRule.Status.State).To(Equal(v2alpha1.Error))
			return
		}

		Expect(gotGateway).ToNot(BeNil())
		Expect(gotGateway.Name).To(Equal(generatedGatewayName))
		Expect(gotGateway.Namespace).To(Equal(externalGatewayNamespace))
	},
		Entry("should return generated gateway when ExternalGateway exists",
			&v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					ExternalGateway: ptr.To(fmt.Sprintf("%s/%s", externalGatewayNamespace, externalGatewayName)),
				},
			},
			true,  // createExternalGateway
			true,  // createGeneratedGateway
			false, // expectNilGateway
			false, // expectError
		),
		Entry("should set error status when ExternalGateway does not exist",
			&v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					ExternalGateway: ptr.To(fmt.Sprintf("%s/%s", externalGatewayNamespace, externalGatewayName)),
				},
			},
			false, // createExternalGateway
			false, // createGeneratedGateway
			true,  // expectNilGateway
			false, // expectError
		),
		Entry("should set error status when ExternalGateway exists but generated Gateway does not",
			&v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					ExternalGateway: ptr.To(fmt.Sprintf("%s/%s", externalGatewayNamespace, externalGatewayName)),
				},
			},
			true,  // createExternalGateway
			false, // createGeneratedGateway
			true,  // expectNilGateway
			false, // expectError
		),
	)
})
