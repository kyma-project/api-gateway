package controllers_test

import (
	"context"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/kyma-incubator/api-gateway/internal/processing"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/controllers"
	. "github.com/onsi/ginkgo"
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
	ts                                       *testSuite
	serviceName, host, authStrategy, gateway string
	servicePort                              uint32
	isExernal                                bool
)

var _ = Describe("Controller", func() {
	Describe("Reconcile", func() {
		Context("APIRule", func() {
			It("should update status", func() {
				testAPI := fixAPI()

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)
				ctx := context.Background()

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				res := gatewayv1beta1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &res)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				Expect(res.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
				Expect(res.Status.APIRuleStatus.Code).To(Equal(gatewayv1beta1.StatusOK))
			})
		})
	})
})

func fixAPI() *gatewayv1beta1.APIRule {
	serviceName = "test"
	servicePort = 8000
	host = "foo.bar"
	isExernal = false
	authStrategy = "noop"
	gateway = "some-gateway.some-namespace.foo"

	return &gatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
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
								Config: nil,
							},
						},
					},
				},
			},
		},
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

	return &testSuite{
		mgr: getFakeManager(fake.NewFakeClientWithScheme(scheme.Scheme, objects...), scheme.Scheme),
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

func (f fakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (f fakeManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return nil
}

func (f fakeManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return nil
}

func (fakeManager) Add(manager.Runnable) error {
	return nil
}

func (fakeManager) SetFields(interface{}) error {
	return nil
}

func (fakeManager) Start(ctx context.Context) error {
	return nil
}

func (fakeManager) GetConfig() *rest.Config {
	return &rest.Config{}
}

func (f *fakeManager) GetScheme() *runtime.Scheme {
	// Setup schemes for all resources
	return f.sch
}

func (f *fakeManager) GetClient() client.Client {
	return f.client
}

func (fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (fakeManager) GetCache() cache.Cache {
	return nil
}

func (fakeManager) GetRecorder(name string) record.EventRecorder {
	return nil
}

func (fakeManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (fakeManager) GetAPIReader() client.Reader {
	return nil
}

func (fakeManager) GetWebhookServer() *webhook.Server {
	return nil
}

func (fakeManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (fakeManager) GetLogger() logr.Logger {
	return logr.Logger{}
}

func (fakeManager) Stop() meta.RESTMapper {
	return nil
}

func getFakeManager(cli client.Client, sch *runtime.Scheme) manager.Manager {
	return &fakeManager{
		client: cli,
		sch:    sch,
	}
}
