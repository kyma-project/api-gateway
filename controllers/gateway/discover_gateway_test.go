package gateway

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	v1beta12 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("discover gateway", func() {
	const gatewayName = "gateway-name"
	const gatewayNamespace = "gateway-namespace"

	var _ = DescribeTable("should return gateway or error according to the APIRule", func(apiRule *v2alpha1.APIRule, gatewayShouldExist bool, expectErr bool) {
		// given
		scheme := runtime.NewScheme()
		schemeBuilder := runtime.NewSchemeBuilder(v1beta1.AddToScheme)
		Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())

		k8sClientBuilder := fake.NewClientBuilder().WithScheme(scheme)
		if gatewayShouldExist {
			k8sClientBuilder.WithObjects(&v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatewayName,
					Namespace: gatewayNamespace,
				},
				Spec: v1beta12.Gateway{
					Servers: []*v1beta12.Server{
						{
							Hosts: []string{"*.example.com"},
							Port: &v1beta12.Port{
								Number:   80,
								Protocol: "HTTPS",
								Name:     "https",
							},
						},
					},
				},
			})
		}

		// when
		gotGateway, gotErr := discoverGateway(k8sClientBuilder.Build(), context.Background(), logr.Discard(), apiRule)

		// then

		Expect(gotErr != nil).To(Equal(expectErr))
		if expectErr {
			return
		}

		Expect(gotGateway.Name).To(Equal(gatewayName))
		Expect(gotGateway.Namespace).To(Equal(gatewayNamespace))
	},
		Entry("should return gateway when it exists", &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To(fmt.Sprintf("%s/%s", gatewayNamespace, gatewayName)),
			},
		}, true, false),
		Entry("should return error when gateway does not exist", &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: nil,
			},
		}, false, true),
		Entry("should return error when gateway is not in namespacedName format", &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To(fmt.Sprintf("%s.%s.svc.cluster.local", gatewayName, gatewayNamespace)),
			},
		}, false, true),
	)
})
