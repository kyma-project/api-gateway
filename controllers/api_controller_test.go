package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	istioint "github.com/kyma-incubator/api-gateway/internal/types/istio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	ts                         *testSuite
	serviceName, host, gateway string
	servicePort                uint32
	isExernal                  bool
)

type FakeConfigMapReader struct {
	Content string
}

func (f FakeConfigMapReader) ReadConfigMap(_ context.Context, _ client.Client) ([]byte, error) {
	return io.ReadAll(bytes.NewBufferString(f.Content))
}

var _ = Describe("Controller", func() {
	Describe("Reconcile", func() {
		Context("APIRule", func() {
			It("should update status", func() {
				testAPI := getApiRule("noop", nil)

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				fakeReader := FakeConfigMapReader{Content: fmt.Sprintf("jwtHandler: %s", helpers.JWT_HANDLER_ORY)}
				helpers.ReadConfigMapHandle = fakeReader.ReadConfigMap

				defer func() {
					helpers.ReadConfigMapHandle = helpers.ReadConfigMap
				}()

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				apiRule := gatewayv1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
				Expect(err).ToNot(HaveOccurred())
				Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				Expect(apiRule.Status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				Expect(apiRule.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			})

			It("should fail if config is empty", func() {
				testAPI := getApiRule("noop", nil)

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				errorReader := FakeConfigMapReader{}
				helpers.ReadConfigMapHandle = errorReader.ReadConfigMap

				defer func() {
					helpers.ReadConfigMapHandle = helpers.ReadConfigMap
				}()

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())

				apiRule := gatewayv1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
				Expect(err).ToNot(HaveOccurred())
				Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			})

			It("should fail if config is in wrong format", func() {
				testAPI := getApiRule("noop", nil)

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				fakeReader := FakeConfigMapReader{Content: "<xml/>"}
				helpers.ReadConfigMapHandle = fakeReader.ReadConfigMap

				defer func() {
					helpers.ReadConfigMapHandle = helpers.ReadConfigMap
				}()

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())

				apiRule := gatewayv1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
				Expect(err).ToNot(HaveOccurred())
				Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
			})

			It("should fail if config is unsupported", func() {
				testAPI := getApiRule("noop", nil)

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				fakeReader := FakeConfigMapReader{Content: "jwtHandler: foo"}
				helpers.ReadConfigMapHandle = fakeReader.ReadConfigMap

				defer func() {
					helpers.ReadConfigMapHandle = helpers.ReadConfigMap
				}()

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())

				apiRule := gatewayv1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
				Expect(err).ToNot(HaveOccurred())
				Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))

				Expect(apiRule.Status.APIRuleStatus.Description).To(Equal(`Validation error: Attribute "": Unsupported JWT Handler: foo`))
			})

			Context("when the jwt handler is istio", func() {
				It("should update status", func() {
					testAPI := getApiRule("jwt", getJWTIstioConfig())

					ts = getTestSuite(testAPI)
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

					apiRule := gatewayv1beta1.APIRule{}
					err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &apiRule)
					Expect(err).ToNot(HaveOccurred())
					Expect(apiRule.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
					Expect(apiRule.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
					Expect(apiRule.Status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
					Expect(apiRule.Status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				})
			})
		})
	})
})

func getApiRule(authStrategy string, authConfig *runtime.RawExtension) *gatewayv1beta1.APIRule {
	serviceName = "test"
	servicePort = 8000
	host = "foo.bar"
	isExernal = false
	gateway = "some-gateway.some-namespace.foo"

	return &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "some-namespace",
			Generation: 1,
		},
		Spec: gatewayv1beta1.APIRuleSpec{
			Host: &host,
			Service: &gatewayv1beta1.Service{
				Name:       &serviceName,
				Port:       &servicePort,
				IsExternal: &isExernal,
			},
			Gateway: &gateway,
			Rules: []gatewayv1beta1.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*gatewayv1beta1.Authenticator{
						{
							Handler: &gatewayv1beta1.Handler{
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

func getJWTIstioConfig() *runtime.RawExtension {
	return getRawConfig(
		istioint.JwtConfig{
			Authentications: []istioint.JwtAuth{
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
	return &controllers.APIRuleReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Api"),
		DomainAllowList: []string{"bar", "kyma.local"},
		CorsConfig: &processing.CorsConfig{
			AllowOrigins: TestAllowOrigins,
			AllowMethods: TestAllowMethods,
			AllowHeaders: TestAllowHeaders,
		},
		GeneratedObjectsLabels: map[string]string{},
		Config:                 &helpers.Config{},
	}
}

type testSuite struct {
	mgr manager.Manager
}

func getTestSuite(objects ...runtime.Object) *testSuite {
	err := gatewayv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = networkingv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	return &testSuite{
		mgr: getFakeManager(fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build(), scheme.Scheme),
	}
}

type fakeManager struct {
	client client.Client
	sch    *runtime.Scheme
}

func (f fakeManager) GetControllerOptions() v1alpha1.ControllerConfigurationSpec {
	return v1alpha1.ControllerConfigurationSpec{}
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

func (f fakeManager) GetWebhookServer() *webhook.Server {
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

func getFakeManager(cli client.Client, sch *runtime.Scheme) manager.Manager {
	return &fakeManager{
		client: cli,
		sch:    sch,
	}
}
