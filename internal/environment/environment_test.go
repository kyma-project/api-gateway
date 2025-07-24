package environment_test

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/environment"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Loader", func() {
	var loader *environment.Loader

	configMapWithProjectName := func(projectName string) *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "shoot-info",
			},
			Data: map[string]string{
				"projectName": projectName,
			},
		}
	}

	configMapWithoutProjectName := func() *corev1.ConfigMap {
		return &corev1.ConfigMap{
			Data: map[string]string{},
		}
	}

	loaderWithFakeClient := func(objects ...client.Object) *environment.Loader {
		scheme := runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		k8sClient := createFakeClient(objects...)
		return &environment.Loader{
			K8sClient: k8sClient,
			Config: &environment.Config{
				RunsOnStage: false,
				Loaded:      &atomic.Bool{},
			},
			Log: logr.Discard(),
		}
	}

	Context("when projectName is 'kyma-stage'", func() {
		BeforeEach(func() {
			loader = loaderWithFakeClient(configMapWithProjectName("kyma-stage"))
		})

		It("should set RunsOnStage to true", func() {
			err := loader.Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(loader.Config.Loaded.Load()).To(BeTrue())
			Expect(loader.Config.RunsOnStage).To(BeTrue())
		})
	})

	Context("when projectName is not 'kyma-stage'", func() {
		BeforeEach(func() {
			loader = loaderWithFakeClient(configMapWithProjectName("other-project"))
		})

		It("should set RunsOnStage to false", func() {
			err := loader.Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(loader.Config.Loaded.Load()).To(BeTrue())
			Expect(loader.Config.RunsOnStage).To(BeFalse())
		})
	})

	Context("when ConfigMap has no projectName", func() {
		BeforeEach(func() {
			loader = loaderWithFakeClient(configMapWithoutProjectName())
		})

		It("should set RunsOnStage to false", func() {
			err := loader.Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(loader.Config.Loaded.Load()).To(BeTrue())
			Expect(loader.Config.RunsOnStage).To(BeFalse())
		})
	})

	Context("when ConfigMap is missing", func() {
		BeforeEach(func() {
			loader = loaderWithFakeClient()
		})

		It("should set RunsOnStage to false", func() {
			err := loader.Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(loader.Config.Loaded.Load()).To(BeTrue())
			Expect(loader.Config.RunsOnStage).To(BeFalse())
		})
	})
})
