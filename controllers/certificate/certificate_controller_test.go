package certificate

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API-Gateway Controller", func() {
	Context("Reconcile", func() {
		It("Should return an error when Secret CR was not found", func() {
			// given
			c := createFakeClient()
			agr := &CertificateReconciler{
				Client:                 c,
				Scheme:                 getTestScheme(),
				log:                    logr.Discard(),
				reconciliationInterval: 1 * time.Second,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: secretName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should return an error when unable to get Secret CR", func() {
			// given
			secretCR := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNamespace,
				},
			}

			c := createFakeClient(secretCR)
			fc := &shouldFailClient{c, true}
			agr := &CertificateReconciler{
				Client:                 fc,
				Scheme:                 getTestScheme(),
				log:                    logr.Discard(),
				reconciliationInterval: 1 * time.Second,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: secretName}})

			// then
			Expect(err).To(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})

		It("Should succeed when Secret CR is present and valid", func() {
			// given
			certificate, key, err := generateCertificate(serviceName, testNamespace)
			Expect(err).ShouldNot(HaveOccurred())

			secretCR := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNamespace,
				},
				Data: map[string][]byte{
					certificateName: certificate,
					keyName:         key,
				},
			}

			c := createFakeClient(secretCR)
			agr := &CertificateReconciler{
				Client:                 c,
				Scheme:                 getTestScheme(),
				log:                    logr.Discard(),
				reconciliationInterval: 1 * time.Second,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: secretName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(1 * time.Second))
		})

		It("Should return error when APIRule CRD do not contain webhook spec but have to create new certificate and reschedule reconcile in a minute", func() {
			// given
			certificate, key, err := generateSelfSignedCertificate(serviceName, nil, []string{}, time.Nanosecond*1)
			Expect(err).ShouldNot(HaveOccurred())

			secretCR := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNamespace,
				},
				Data: map[string][]byte{
					certificateName: certificate,
					keyName:         key,
				},
			}

			crd := apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: APIRuleCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Conversion: &apiextensionsv1.CustomResourceConversion{
						Webhook: &apiextensionsv1.WebhookConversion{
							ClientConfig: nil,
						},
					},
				},
			}

			c := createFakeClient(secretCR)
			Expect(c.Create(context.Background(), &crd)).To(Succeed())

			agr := &CertificateReconciler{
				Client:                 c,
				Scheme:                 getTestScheme(),
				log:                    logr.Discard(),
				reconciliationInterval: 1 * time.Second,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: secretName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(1 * time.Minute))
		})

		It("Should succeed when Secret CR is present and generate new certificate when current is expired", func() {
			// given
			certificate, key, err := generateSelfSignedCertificate(serviceName, nil, []string{}, time.Nanosecond*1)
			Expect(err).ShouldNot(HaveOccurred())

			secretCR := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNamespace,
				},
				Data: map[string][]byte{
					certificateName: certificate,
					keyName:         key,
				},
			}

			crd := apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: APIRuleCRDName,
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

			c := createFakeClient(secretCR)
			Expect(c.Create(context.Background(), &crd)).To(Succeed())

			agr := &CertificateReconciler{
				Client:                 c,
				Scheme:                 getTestScheme(),
				log:                    logr.Discard(),
				reconciliationInterval: 1 * time.Second,
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: secretName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.RequeueAfter).Should(Equal(1 * time.Second))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(secretCR), secretCR)).Should(Succeed())
			Expect(secretCR.Data[certificateName]).ShouldNot(Equal(certificate))

			Expect(c.Get(context.TODO(), types.NamespacedName{Name: APIRuleCRDName}, &crd)).Should(Succeed())
			Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).ShouldNot(Equal(certificate))
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
