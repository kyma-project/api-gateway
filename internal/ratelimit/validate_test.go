package ratelimit_test

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/ratelimit/v1alpha1"
	ratelimitvalidator "github.com/kyma-project/api-gateway/internal/ratelimit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var sc *runtime.Scheme

func init() {
	sc = runtime.NewScheme()
	_ = scheme.AddToScheme(sc)
	_ = v1alpha1.AddToScheme(sc)
}

var _ = Describe("RateLimit CR Validation", func() {
	It("Should pass if there is only one RateLimit CR and matching pod", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}

		rlCR := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}

		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test-pod",
				Labels: commonSelectors,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "istio-proxy",
					},
					{
						Name: "test-app",
					},
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &testPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should pass if there is more than one RateLimit CR and matching pods without conflicts", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}

		rlCR := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		rlCR2 := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rl2",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "test2",
				},
			},
		}

		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test-pod",
				Labels: commonSelectors,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "istio-proxy",
					},
					{
						Name: "test-app",
					},
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &rlCR2, &testPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should fail if there is already a RateLimit CR assigned to the matching pod", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}
		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test-pod",
				Labels: commonSelectors,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "istio-proxy",
					},
					{
						Name: "test-app",
					},
				},
			},
		}
		oldRL := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "old-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		newRL := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "new-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&oldRL, &testPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, newRL)
		Expect(err).To(HaveOccurred())
	})

	It("Should fail if the pod is not sidecar-enabled", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}
		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-ns",
				Labels:    commonSelectors,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "test-app",
					},
				},
			},
		}
		rlCR := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &testPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).To(HaveOccurred())
	})

	It("Should pass if the chosen pod is a part of the istio ingress-gateway", func() {
		ingressGatewaySelectors := map[string]string{
			"app":   "istio-ingressgateway",
			"istio": "ingressgateway",
		}

		rlCR := v1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rl",
			},
			Spec: v1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "istio-ingressgateway",
				},
			},
		}

		ingressGatewayPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-ingressgateway-test",
				Namespace: "istio-system",
				Labels:    ingressGatewaySelectors,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "istio-proxy",
					},
				},
			},
		}

		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &ingressGatewayPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)

		Expect(err).ToNot(HaveOccurred())
	})
})
