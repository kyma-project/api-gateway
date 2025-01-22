package ratelimit_test

import (
	"context"
	"fmt"
	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	ratelimitvalidator "github.com/kyma-project/api-gateway/internal/ratelimit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var sc *runtime.Scheme

func init() {
	sc = runtime.NewScheme()
	_ = scheme.AddToScheme(sc)
	_ = ratelimitv1alpha1.AddToScheme(sc)
}

var _ = Describe("RateLimit CR Validation", func() {
	It("Should pass if there is only one RateLimit CR and matching pod", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}

		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}

		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels:    commonSelectors,
				Annotations: map[string]string{
					"sidecar.istio.io/status": "test",
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

		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
				Annotations: map[string]string{
					"sidecar.istio.io/status": "test",
				},
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		rlCR2 := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl2",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "test2",
				},
			},
		}

		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels:    commonSelectors,
				Annotations: map[string]string{
					"sidecar.istio.io/status": "test",
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &rlCR2, &testPod).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should fail if there is no pods matching for the selectors in RateLimit CR", func() {
		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "test",
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("no pods found with the given selectors: %v in namespace %s", rlCR.Spec.SelectorLabels, rlCR.Namespace)))
	})

	It("Should fail if there are already a RateLimit CRs assigned to the matching pods", func() {
		//given
		testPod1 := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod1",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"rateLimitSelector": "ratelimit",
					"otherSelector":     "other1",
				},
				Annotations: map[string]string{
					"sidecar.istio.io/status": "test",
				},
			},
		}
		testPod2 := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod2",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"rateLimitSelector": "ratelimit",
					"otherSelector2":    "other2",
				},
				Annotations: map[string]string{
					"sidecar.istio.io/status": "test",
				},
			},
		}
		existingRL1 := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existingRL1",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"otherSelector": "other1",
				},
			},
		}
		existingRL2 := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existingRL2",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"otherSelector2": "other2",
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&existingRL1, &existingRL2, &testPod1, &testPod2).Build()

		// when
		newRL := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "new-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"rateLimitSelector": "ratelimit", // should be common for both pods
				},
			},
		}
		err := ratelimitvalidator.Validate(context.Background(), c, newRL)

		//then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("conflicting with the following RateLimit CRs: %s, %s", existingRL1.Name, existingRL2.Name)))
	})

	It("Should fail if the pod is not sidecar-enabled", func() {
		commonSelectors := map[string]string{
			"app": "test",
		}
		testPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels:    commonSelectors,
			},
		}
		testPod2 := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod2",
				Namespace: "test-namespace",
				Labels:    commonSelectors,
			},
		}
		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: commonSelectors,
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR, &testPod, &testPod2).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("sidecar injection is not enabled for the following pods: test-namespace/test-pod, test-namespace/test-pod2"))
	})

	It("Should pass if the selected pod is a part of the istio ingress-gateway and RateLimit CR is in the ingress gateway namespace", func() {
		ingressGatewaySelectors := map[string]string{
			"app":   "istio-ingressgateway",
			"istio": "ingressgateway",
		}

		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "istio-system",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"istio": "ingressgateway",
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

	It("Should fail if the selected pod is a part of the Istio ingress-gateway but the RateLimit CR is in a different namespace", func() {
		ingressGatewaySelectors := map[string]string{
			"app":   "istio-ingressgateway",
			"istio": "ingressgateway",
		}

		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"istio": "ingressgateway",
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

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("no pods found with the given selectors: %v in namespace %s", rlCR.Spec.SelectorLabels, rlCR.Namespace)))
	})
	It("Should fail if default bucket fill interval lower than 50ms", func() {
		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "test",
				},
				Local: ratelimitv1alpha1.LocalConfig{DefaultBucket: ratelimitv1alpha1.BucketSpec{FillInterval: &metav1.Duration{Duration: 0}}},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("default_bucket: fill_interval must be greater or equal 50ms"))
	})
	It("Should fail if exact bucket fill interval lower than 50ms", func() {
		rlCR := ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rl",
				Namespace: "test-namespace",
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				SelectorLabels: map[string]string{
					"app": "test",
				},
				Local: ratelimitv1alpha1.LocalConfig{
					DefaultBucket: ratelimitv1alpha1.BucketSpec{FillInterval: &metav1.Duration{Duration: 60 * time.Minute}},
					Buckets: []ratelimitv1alpha1.BucketConfig{
						{
							Bucket: ratelimitv1alpha1.BucketSpec{FillInterval: &metav1.Duration{Duration: 0}},
						},
					},
				},
			},
		}
		c := fake.NewClientBuilder().WithScheme(sc).WithObjects(&rlCR).Build()

		err := ratelimitvalidator.Validate(context.Background(), c, rlCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("bucket '[0]': fill_interval must be greater or equal 50ms"))
	})
})
