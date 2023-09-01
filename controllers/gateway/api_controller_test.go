package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	"io"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/config"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/kyma-project/api-gateway/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type FakeConfigMapReader struct {
	Content string
}

func (f FakeConfigMapReader) ReadConfigMap(_ context.Context, _ client.Client) ([]byte, error) {
	return io.ReadAll(bytes.NewBufferString(f.Content))
}

var _ = Describe("Controller", func() {

	var ts *testSuite

	Describe("Reconcile", func() {
		Context("APIRule", func() {
			It("should update status", func() {
				testAPI := getApiRule("noop", nil)
				svc := getService(*testAPI.Spec.Service.Name)
				ts = getTestSuite(testAPI, svc)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				apiRule := v1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
				Expect(err).ToNot(HaveOccurred())
				Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusOK))
				Expect(apiRule.Status.VirtualServiceStatus.Code).To(Equal(v1beta1.StatusOK))
			})

			Context("when the jwt handler is istio", func() {
				It("should update status", func() {
					testAPI := getApiRule("jwt", getJWTIstioConfig())
					svc := getService(*testAPI.Spec.Service.Name)
					ts = getTestSuite(testAPI, svc)
					reconciler := getAPIReconciler(ts.mgr)
					ctx := context.Background()

					fakeReader := FakeConfigMapReader{Content: fmt.Sprintf("jwtHandler: %s", helpers.JWT_HANDLER_ISTIO)}
					helpers.ReadConfigMapHandle = fakeReader.ReadConfigMap

					defer func() {
						helpers.ReadConfigMapHandle = helpers.ReadConfigMap
					}()

					result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
					Expect(err).ToNot(HaveOccurred())
					Expect(result.Requeue).To(BeFalse())

					apiRule := v1beta1.APIRule{}
					err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
					Expect(err).ToNot(HaveOccurred())
					Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(v1beta1.StatusOK))
					Expect(apiRule.Status.VirtualServiceStatus.Code).To(Equal(v1beta1.StatusOK))
					Expect(apiRule.Status.RequestAuthenticationStatus.Code).To(Equal(v1beta1.StatusOK))
					Expect(apiRule.Status.AuthorizationPolicyStatus.Code).To(Equal(v1beta1.StatusOK))
				})
			})
		})
	})
})

func getApiRule(authStrategy string, authConfig *runtime.RawExtension) *v1beta1.APIRule {
	var (
		serviceName        = "test"
		servicePort uint32 = 8000
		host               = "foo.bar"
		isExternal         = false
		gateway            = "some-gateway.some-namespace.foo"
	)

	return &v1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "some-namespace",
			Generation: 1,
		},
		Spec: v1beta1.APIRuleSpec{
			Host: &host,
			Service: &v1beta1.Service{
				Name:       &serviceName,
				Port:       &servicePort,
				IsExternal: &isExternal,
			},
			Gateway: &gateway,
			Rules: []v1beta1.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*v1beta1.Authenticator{
						{
							Handler: &v1beta1.Handler{
								Name:   authStrategy,
								Config: authConfig,
							},
						},
					},
				},
			},
		},
	}
}

func getService(name string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "some-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
		},
	}
}

func getJWTIstioConfig() *runtime.RawExtension {
	return getRawConfig(
		v1beta1.JwtConfig{
			Authentications: []*v1beta1.JwtAuthentication{
				{
					Issuer:  "https://example.com/",
					JwksUri: "https://example.com/.well-known/jwks.json",
				},
			},
		})
}

func getRawConfig(config any) *runtime.RawExtension {
	b, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: b,
	}
}

func getAPIReconciler(mgr manager.Manager) reconcile.Reconciler {

	reconcilerConfig := gateway.ApiRuleReconcilerConfiguration{
		AllowListedDomains: "bar, kyma.local",
		CorsAllowOrigins:   "regex:.*",
		CorsAllowMethods:   "GET,POST,PUT,DELETE",
		CorsAllowHeaders:   "header1,header2",
	}

	apiReconciler, err := gateway.NewApiRuleReconciler(mgr, reconcilerConfig)
	Expect(err).NotTo(HaveOccurred())

	return apiReconciler
}

type testSuite struct {
	mgr manager.Manager
}

func getTestSuite(objects ...client.Object) *testSuite {
	err := v1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = networkingv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	return &testSuite{
		mgr: getFakeManager(fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).WithStatusSubresource(objects...).Build(), scheme.Scheme),
	}
}

type fakeManager struct {
	client client.Client
	sch    *runtime.Scheme
}

func (f fakeManager) GetControllerOptions() config.Controller {
	return config.Controller{}
}

func (f fakeManager) Elected() <-chan struct{} {
	return nil
}

func (f fakeManager) AddMetricsExtraHandler(_ string, _ http.Handler) error {
	return nil
}

func (f fakeManager) AddHealthzCheck(_ string, _ healthz.Checker) error {
	return nil
}

func (f fakeManager) AddReadyzCheck(_ string, _ healthz.Checker) error {
	return nil
}

func (fakeManager) Add(manager.Runnable) error {
	return nil
}

func (f fakeManager) SetFields(interface{}) error {
	return nil
}

func (f fakeManager) Start(_ context.Context) error {
	return nil
}

func (f fakeManager) GetConfig() *rest.Config {
	return &rest.Config{}
}

func (f fakeManager) GetScheme() *runtime.Scheme {
	// Setup schemes for all resources
	return f.sch
}

func (f fakeManager) GetClient() client.Client {
	return f.client
}

func (f fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (f fakeManager) GetCache() cache.Cache {
	return nil
}

func (f fakeManager) GetRecorder(_ string) record.EventRecorder {
	return nil
}

func (f fakeManager) GetEventRecorderFor(_ string) record.EventRecorder {
	return nil
}

func (f fakeManager) GetAPIReader() client.Reader {
	return nil
}

func (f fakeManager) GetWebhookServer() webhook.Server {
	return nil
}

func (f fakeManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (f fakeManager) GetLogger() logr.Logger {
	return logr.Logger{}
}

func (f fakeManager) Stop() meta.RESTMapper {
	return nil
}

func (f fakeManager) GetHTTPClient() *http.Client {
	return nil
}

func getFakeManager(cli client.Client, sch *runtime.Scheme) manager.Manager {
	return &fakeManager{
		client: cli,
		sch:    sch,
	}
}
