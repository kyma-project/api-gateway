package v2alpha1_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation/v2alpha1"
)

var _ = Describe("Validate", func() {

	serviceName := "some-service"
	gatewayName := "namespace/gateway"
	host := gatewayv2alpha1.Host(serviceName + ".test.dev")

	It("should validate if spec is not empty", func() {
		//given
		apiRule := &gatewayv2alpha1.APIRule{
			Spec: gatewayv2alpha1.APIRuleSpec{},
		}

		fakeClient := createFakeClient()

		//when
		problems := (&v2alpha1.APIRuleValidator{ApiRule: apiRule}).Validate(
			context.Background(),
			fakeClient,
			networkingv1beta1.VirtualServiceList{},
			networkingv1beta1.GatewayList{},
		)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec"))
		Expect(
			problems[0].Message,
		).To(Equal("APIRule in version v2alpha1 contains an empty spec. To troubleshoot, see https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-70-empty-apirule-spec."))
	})

	It("should invoke rules validation", func() {
		//given
		gatewayList := networkingv1beta1.GatewayList{
			Items: []*networkingv1beta1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway",
						Namespace: "namespace",
					},
				},
			},
		}
		apiRule := &gatewayv2alpha1.APIRule{
			Spec: gatewayv2alpha1.APIRuleSpec{
				Rules: nil,
				Service: &gatewayv2alpha1.Service{
					Name: ptr.To(serviceName),
					Port: ptr.To(uint32(8080)),
				},
				Hosts:   []*gatewayv2alpha1.Host{&host},
				Gateway: ptr.To(gatewayName),
			},
		}

		service := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": serviceName,
				},
			},
		}
		fakeClient := createFakeClient(&service)

		//when
		problems := (&v2alpha1.APIRuleValidator{ApiRule: apiRule}).Validate(context.Background(), fakeClient, networkingv1beta1.VirtualServiceList{}, gatewayList)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

})

func createFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := gatewayv2alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}
