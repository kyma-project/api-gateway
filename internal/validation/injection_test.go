package validation_test

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"

	"istio.io/api/type/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Istio injection validation", func() {

	scheme := runtime.NewScheme()
	err := gatewayv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	var k8sFakeClient client.WithWatch

	BeforeEach(func() {
		k8sFakeClient = fake.NewClientBuilder().Build()
	})

	It("Should not fail when the Pod for which the service is specified is in different ns", func() {
		//given
		err := k8sFakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test",
				Labels: map[string]string{
					"app": "test",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := (&validation.InjectionValidator{Ctx: context.Background(), Client: k8sFakeClient}).Validate("some.attribute", &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": "test"}}, "default")
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail when the workload selector is nil", func() {
		//when
		problems, err := (&validation.InjectionValidator{Ctx: context.Background(), Client: k8sFakeClient}).Validate("some.attribute", nil, "default")
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.injection"))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})

	It("Should fail when the Pod for which the service is specified is not istio injected and in the same not default namespace", func() {
		//given
		err := k8sFakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test",
				Labels: map[string]string{
					"app": "test",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := (&validation.InjectionValidator{Ctx: context.Background(), Client: k8sFakeClient}).Validate("some.attribute", &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": "test"}}, "test")
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute"))
		Expect(problems[0].Message).To(Equal("Pod test/test-pod does not have an injected istio sidecar"))
	})

	It("Should fail when the Pod for which the service is specified is not istio injected", func() {
		//given
		err := k8sFakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := (&validation.InjectionValidator{Ctx: context.Background(), Client: k8sFakeClient}).Validate("some.attribute", &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": "test"}}, "default")
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute"))
		Expect(problems[0].Message).To(Equal("Pod default/test-pod does not have an injected istio sidecar"))
	})

	It("Should not fail when the Pod for which the service is specified is istio injected", func() {
		//given
		err := k8sFakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-pod-injected",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-injected",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "istio-proxy"},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := (&validation.InjectionValidator{Ctx: context.Background(), Client: k8sFakeClient}).Validate("some.attribute", &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": "test-injected"}}, "default")
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(0))
	})
})
