//go:build dev_features

package ratelimit_test

import (
	"context"
	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers/gateway/ratelimit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

var (
	testEnv *envtest.Environment
	c       ctrlclient.Client
	ctx     context.Context
	cancel  context.CancelFunc
)
var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("Setting up the test environment")
	ctx, cancel = context.WithCancel(context.Background())
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.FromSlash("../../../config/crd/bases"),
			filepath.FromSlash("../../../hack/crds"),
		},
	}
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	s := getTestScheme()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
		Logger: zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)),
	})
	Expect(err).ToNot(HaveOccurred())

	c, err = ctrlclient.New(cfg, ctrlclient.Options{Scheme: s})
	Expect(err).ToNot(HaveOccurred())
	rec := &ratelimit.RateLimitReconciler{
		Client:          c,
		Scheme:          s,
		ReconcilePeriod: time.Second,
	}
	Expect(rec.SetupWithManager(mgr)).Should(Succeed())
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).Should(Succeed())
	}()
})
var _ = AfterSuite(func() {
	By("Tearing down the test environment")
	cancel()
	Expect(testEnv.Stop()).Should(Succeed())

})
var _ = Describe("Rate Limit Controller", func() {
	var ns *corev1.Namespace
	BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "ratelimit-test-",
			},
		}
		Expect(c.Create(ctx, ns)).Should(Succeed())
		Expect(ns.Name).ToNot(BeEmpty())
	})
	It("should reconcile RateLimit Status to Error state when validation fails", func() {
		By("Creating RateLimit resource without matching pods")
		namespace := ns.Name
		rl := &ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "rate-limit-",
				Namespace:    namespace,
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				EnableResponseHeaders: true,
				SelectorLabels:        map[string]string{"app": "test"},
				Local: ratelimitv1alpha1.LocalConfig{
					DefaultBucket: ratelimitv1alpha1.BucketSpec{
						MaxTokens:     20,
						TokensPerFill: 20,
						FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
					},
				},
			}}
		Expect(c.Create(ctx, rl)).Should(Succeed())
		Expect(rl.Name).ShouldNot(BeEmpty())
		Eventually(func() string {
			Expect(c.Get(ctx, ctrlclient.ObjectKeyFromObject(rl), rl)).Should(Succeed())
			return rl.Status.State
		}).Should(Equal("Error"))
	})
	It("should reconcile RateLimit Status to Ready state when resource is created", func() {
		namespace := ns.Name
		By("Creating test pod")
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: namespace,
				Annotations: map[string]string{
					"sidecar.istio.io/status": "",
				},
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "busybox",
					},
				}}}
		Expect(c.Create(ctx, pod)).Should(Succeed())

		By("Creating RateLimit resource")
		rl := &ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "rate-limit-",
				Namespace:    namespace,
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				EnableResponseHeaders: true,
				SelectorLabels:        map[string]string{"app": "test"},
				Local: ratelimitv1alpha1.LocalConfig{
					DefaultBucket: ratelimitv1alpha1.BucketSpec{
						MaxTokens:     20,
						TokensPerFill: 20,
						FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
					},
				},
			}}
		Expect(c.Create(ctx, rl)).Should(Succeed())
		Expect(rl.Name).ShouldNot(BeEmpty())
		Eventually(func() string {
			Expect(c.Get(ctx, ctrlclient.ObjectKeyFromObject(rl), rl)).Should(Succeed())
			return rl.Status.State
		}).Should(Equal("Ready"))

		By("Checking that EnvoyFilter got created")
		ef := &networkingv1alpha3.EnvoyFilter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rl.Name,
				Namespace: rl.Namespace,
			},
		}
		Eventually(c.Get(ctx, ctrlclient.ObjectKeyFromObject(ef), ef)).Should(Succeed())
		By("Checking the OwnerReference of EnvoyFilter")
		expectedOwnerReference := metav1.OwnerReference{
			Kind:               "RateLimit",
			Name:               rl.Name,
			UID:                rl.UID,
			APIVersion:         "gateway.kyma-project.io/v1alpha1",
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		}
		Expect(ef.OwnerReferences).To(ContainElement(expectedOwnerReference))

	})
	It("should reconcile state to Ready when RateLimit is modified with new configuration", func() {
		namespace := ns.Name
		By("Creating test pod")
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: namespace,
				Annotations: map[string]string{
					"sidecar.istio.io/status": "",
				},
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test",
						Image: "busybox",
					},
				}}}
		Expect(c.Create(ctx, pod)).Should(Succeed())
		By("Creating RateLimit resource")
		rl := &ratelimitv1alpha1.RateLimit{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "rate-limit-",
				Namespace:    ns.Name,
			},
			Spec: ratelimitv1alpha1.RateLimitSpec{
				EnableResponseHeaders: false,
				SelectorLabels:        map[string]string{"app": "test"},
				Local: ratelimitv1alpha1.LocalConfig{
					DefaultBucket: ratelimitv1alpha1.BucketSpec{
						MaxTokens:     20,
						TokensPerFill: 20,
						FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
					},
				},
			}}
		Expect(c.Create(ctx, rl)).Should(Succeed())
		Expect(rl.Name).ShouldNot(BeEmpty())
		Eventually(func() string {
			Expect(c.Get(ctx, ctrlclient.ObjectKeyFromObject(rl), rl)).Should(Succeed())
			return rl.Status.State
		}).Should(Equal("Ready"))
		By("Checking that EnvoyFilter got created")
		ef := &networkingv1alpha3.EnvoyFilter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rl.Name,
				Namespace: rl.Namespace,
			},
		}
		Eventually(c.Get(ctx, ctrlclient.ObjectKeyFromObject(ef), ef)).Should(Succeed())
		observedGeneration := ef.Generation
		By("Modifying RateLimit with 'enableResponseHeaders' set to true")
		rl.Spec.EnableResponseHeaders = true
		Expect(c.Update(ctx, rl)).Should(Succeed())
		Eventually(func() string {
			Expect(c.Get(ctx, ctrlclient.ObjectKeyFromObject(rl), rl)).Should(Succeed())
			return rl.Status.State
		}).Should(Equal("Ready"))
		By("Checking that EnvoyFilter got updated")
		Eventually(func() bool {
			Expect(c.Get(ctx, ctrlclient.ObjectKeyFromObject(ef), ef)).Should(Succeed())
			return ef.Generation > observedGeneration
		}).Should(BeTrue())
	})
	Context("RateLimit CRD validation", func() {
		It("should fail if path and headers are empty", func() {
			namespace := ns.Name
			By("Creating test pod")
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/status": "",
					},
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "busybox",
						},
					}}}
			Expect(c.Create(ctx, pod)).Should(Succeed())
			By("Creating RateLimit resource")
			rl := &ratelimitv1alpha1.RateLimit{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "rate-limit-",
					Namespace:    ns.Name,
				},
				Spec: ratelimitv1alpha1.RateLimitSpec{
					EnableResponseHeaders: false,
					SelectorLabels:        map[string]string{"app": "test"},
					Local: ratelimitv1alpha1.LocalConfig{
						DefaultBucket: ratelimitv1alpha1.BucketSpec{
							MaxTokens:     20,
							TokensPerFill: 20,
							FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
						},
						Buckets: []ratelimitv1alpha1.BucketConfig{
							{
								Bucket: ratelimitv1alpha1.BucketSpec{
									MaxTokens:     20,
									TokensPerFill: 20,
									FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
								},
							},
						},
					},
				}}
			err := c.Create(ctx, rl)
			Expect(err).ShouldNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("At least one of 'path' or 'headers' must be set"))
		})
		It("should pass if path and headers are defined", func() {
			namespace := ns.Name
			By("Creating test pod")
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/status": "",
					},
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "busybox",
						},
					}}}
			Expect(c.Create(ctx, pod)).Should(Succeed())
			By("Creating RateLimit resource")
			rl := &ratelimitv1alpha1.RateLimit{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "rate-limit-",
					Namespace:    ns.Name,
				},
				Spec: ratelimitv1alpha1.RateLimitSpec{
					EnableResponseHeaders: false,
					SelectorLabels:        map[string]string{"app": "test"},
					Local: ratelimitv1alpha1.LocalConfig{
						DefaultBucket: ratelimitv1alpha1.BucketSpec{
							MaxTokens:     20,
							TokensPerFill: 20,
							FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
						},
						Buckets: []ratelimitv1alpha1.BucketConfig{
							{
								Path: "/anything",
								Headers: map[string]string{
									"X-Foo": "bar",
								},
								Bucket: ratelimitv1alpha1.BucketSpec{
									MaxTokens:     20,
									TokensPerFill: 20,
									FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
								},
							},
						},
					},
				}}
			Expect(c.Create(ctx, rl)).Should(Succeed())
		})
		It("should pass if only path is defined", func() {
			namespace := ns.Name
			By("Creating test pod")
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/status": "",
					},
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "busybox",
						},
					}}}
			Expect(c.Create(ctx, pod)).Should(Succeed())
			By("Creating RateLimit resource")
			rl := &ratelimitv1alpha1.RateLimit{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "rate-limit-",
					Namespace:    ns.Name,
				},
				Spec: ratelimitv1alpha1.RateLimitSpec{
					EnableResponseHeaders: false,
					SelectorLabels:        map[string]string{"app": "test"},
					Local: ratelimitv1alpha1.LocalConfig{
						DefaultBucket: ratelimitv1alpha1.BucketSpec{
							MaxTokens:     20,
							TokensPerFill: 20,
							FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
						},
						Buckets: []ratelimitv1alpha1.BucketConfig{
							{
								Path: "/anything",
								Bucket: ratelimitv1alpha1.BucketSpec{
									MaxTokens:     20,
									TokensPerFill: 20,
									FillInterval:  &metav1.Duration{Duration: time.Minute * 5},
								},
							},
						},
					},
				}}
			Expect(c.Create(ctx, rl)).Should(Succeed())
		})
	})
})

func getTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	Expect(corev1.AddToScheme(s)).Should(Succeed())
	Expect(apiextensionsv1.AddToScheme(s)).Should(Succeed())
	Expect(networkingv1alpha3.AddToScheme(s)).Should(Succeed())
	Expect(ratelimitv1alpha1.AddToScheme(s)).Should(Succeed())
	return s
}

func TestRateLimitReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RateLimit Controller Suite")
}
