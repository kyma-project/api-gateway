package controllers_test

import (
	"context"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	crClients "github.com/kyma-incubator/api-gateway/internal/clients"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
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

				result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				res := gatewayv1alpha1.APIRule{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &res)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Status.AccessRuleStatus.Code).To(Equal(gatewayv1alpha1.StatusOK))
				Expect(res.Status.VirtualServiceStatus.Code).To(Equal(gatewayv1alpha1.StatusOK))
				Expect(res.Status.APIRuleStatus.Code).To(Equal(gatewayv1alpha1.StatusOK))
			})
		})
	})
})

func fixAPI() *gatewayv1alpha1.APIRule {
	serviceName = "test"
	servicePort = 8000
	host = "foo.bar"
	isExernal = false
	authStrategy = "noop"
	gateway = "some-gateway.some-namespace.foo"

	return &gatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Generation: 1,
		},
		Spec: gatewayv1alpha1.APIRuleSpec{
			Service: &gatewayv1alpha1.Service{
				Name:       &serviceName,
				Port:       &servicePort,
				Host:       &host,
				IsExternal: &isExernal,
			},
			Gateway: &gateway,
			Rules: []gatewayv1alpha1.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*rulev1alpha1.Authenticator{
						{
							Handler: &rulev1alpha1.Handler{
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
	return &controllers.APIReconciler{
		Client:       mgr.GetClient(),
		ExtCRClients: crClients.New(mgr.GetClient()),
		Log:          ctrl.Log.WithName("controllers").WithName("Api"),
		Validator: &validation.APIRule{
			DomainWhiteList: []string{"bar", "kyma.local"},
		},
	}
}

type testSuite struct {
	mgr manager.Manager
}

func getTestSuite(objects ...runtime.Object) *testSuite {
	err := gatewayv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = networkingv1alpha3.AddToScheme(scheme.Scheme)
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

func (fakeManager) Add(manager.Runnable) error {
	return nil
}

func (fakeManager) SetFields(interface{}) error {
	return nil
}

func (fakeManager) Start(<-chan struct{}) error {
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

func getFakeManager(cli client.Client, sch *runtime.Scheme) manager.Manager {
	return &fakeManager{
		client: cli,
		sch:    sch,
	}
}
