package certificate_test

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/controllers/certificate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Certificate Controller", func() {
	Context("Reconcile", func() {
		It("Should return an error when Secret was not found", func() {
			// given
			c := createFakeClient()
			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).Should(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should return an error when unable to get Secret", func() {
			// given
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-gateway-webhook-certificate",
					Namespace: testNamespace,
				},
			}

			c := createFakeClient(secret)
			fc := &shouldFailClient{c, true}
			agr := getReconciler(fc, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).To(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})

		It("Should succeed when Secret is present and valid", func() {
			// given
			certificate, key, err := certificate.GenerateSelfSignedCertificate("api-gateway-webhook-service", nil, []string{}, time.Hour*24*365)
			Expect(err).ShouldNot(HaveOccurred())

			secret := getSecret(certificate, key)

			c := createFakeClient(secret)
			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(time.Hour * 1))
		})

		It("Should return error when APIRule CRD do not contain webhook spec but have to create new certificate and reschedule reconcile in a minute", func() {
			// given
			certificate, key, err := certificate.GenerateSelfSignedCertificate("api-gateway-webhook-service", nil, []string{}, time.Nanosecond*1)
			Expect(err).ShouldNot(HaveOccurred())

			secret := getSecret(certificate, key)

			crd := getCRD(certificate)
			crd.Spec.Conversion.Webhook.ClientConfig = nil

			c := createFakeClient(secret)
			Expect(c.Create(context.Background(), crd)).To(Succeed())

			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			_, err = agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to update certificate into CRD: can not add certificate into CRD: client config for conversion webhook not found in APIRule CRD"))
		})

		It("Should succeed when Secret is present and generate new certificate when current is expired", func() {
			// given
			certificate, key, err := certificate.GenerateSelfSignedCertificate("api-gateway-webhook-service", nil, []string{}, time.Nanosecond*1)
			Expect(err).ShouldNot(HaveOccurred())

			secret := getSecret(certificate, key)
			crd := getCRD(certificate)

			c := createFakeClient(secret)
			Expect(c.Create(context.Background(), crd)).To(Succeed())

			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(time.Hour * 1))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(secret), secret)).Should(Succeed())
			Expect(secret.Data["tls.crt"]).ShouldNot(Equal(certificate))

			Expect(c.Get(context.Background(), types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())
			Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).ShouldNot(Equal(certificate))

			Expect(secret.Data["tls.crt"]).To(Equal(crd.Spec.Conversion.Webhook.ClientConfig.CABundle))
		})

		It("Should succeed when Secret is present and generate new certificate when current has missing required keys", func() {
			// given
			secret := getSecret([]byte{}, []byte{})
			secret.Data = make(map[string][]byte)

			crd := getCRD([]byte{})

			c := createFakeClient(secret)
			Expect(c.Create(context.Background(), crd)).To(Succeed())

			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(time.Hour * 1))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(secret), secret)).Should(Succeed())
			Expect(secret.Data["tls.crt"]).ShouldNot(Equal([]byte{}))

			Expect(c.Get(context.Background(), types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())
			Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).ShouldNot(Equal([]byte{}))

			Expect(secret.Data["tls.crt"]).To(Equal(crd.Spec.Conversion.Webhook.ClientConfig.CABundle))
		})

		It("Should succeed when Secret is present and generate new certificate when current has incorrect certificate", func() {
			// given
			secret := getSecret([]byte{1, 2, 3}, []byte{3, 2, 1})
			crd := getCRD([]byte{})

			c := createFakeClient(secret)
			Expect(c.Create(context.Background(), crd)).To(Succeed())

			agr := getReconciler(c, getTestScheme(), logr.Discard())

			// when
			result, err := agr.Reconcile(
				context.Background(),
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "api-gateway-webhook-certificate"}},
			)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(time.Hour * 1))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(secret), secret)).Should(Succeed())
			Expect(secret.Data["tls.crt"]).ShouldNot(Equal([]byte{}))

			Expect(c.Get(context.Background(), types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())
			Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).ShouldNot(Equal([]byte{}))

			Expect(secret.Data["tls.crt"]).To(Equal(crd.Spec.Conversion.Webhook.ClientConfig.CABundle))
		})
	})
})

type shouldFailClient struct {
	client.Client
	FailOnGet bool
}

func (p *shouldFailClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if p.FailOnGet {
		return goerrors.New("fail on purpose")
	}
	return p.Client.Get(ctx, key, obj, opts...)
}

func getReconciler(c client.Client, scheme *runtime.Scheme, log logr.Logger) *certificate.Reconciler {
	return &certificate.Reconciler{
		Client: c,
		Scheme: scheme,
		Log:    log,
	}
}

func getSecret(certificate, key []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-webhook-certificate",
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"tls.crt": certificate,
			"tls.key": key,
		},
	}
}

func getCRD(certificate []byte) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "apirules.gateway.kyma-project.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Conversion: &apiextensionsv1.CustomResourceConversion{
				Webhook: &apiextensionsv1.WebhookConversion{
					ClientConfig: &apiextensionsv1.WebhookClientConfig{
						CABundle: certificate,
					},
				},
			},
		},
	}
}
