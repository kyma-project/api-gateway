package controllers_test

import (
	"context"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	servicePort                              int32
	isExernal                                bool
)

var _ = Describe("Controller", func() {
	Describe("Reconcile", func() {
		Context("Gate", func() {
			It("should update status", func() {
				testAPI := fixAPI()

				ts = getTestSuite(testAPI)
				reconciler := getAPIReconciler(ts.mgr)

				result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: testAPI.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				res := gatewayv2alpha1.Gate{}
				err = ts.mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: testAPI.Namespace, Name: testAPI.Name}, &res)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Status.AccessRuleStatus.Code).To(Equal(gatewayv2alpha1.STATUS_SKIPPED))
				Expect(res.Status.PolicyServiceStatus.Code).To(Equal(gatewayv2alpha1.STATUS_SKIPPED))
				Expect(res.Status.VirtualServiceStatus.Code).To(Equal(gatewayv2alpha1.STATUS_OK))
				Expect(res.Status.GateStatus.Code).To(Equal(gatewayv2alpha1.STATUS_OK))
			})
		})
	})
})

func fixAPI() *gatewayv2alpha1.Gate {
	serviceName = "test"
	servicePort = 8000
	host = "foo.bar"
	isExernal = false
	authStrategy = gatewayv2alpha1.PASSTHROUGH
	gateway = "some-gateway.some-namespace.foo"

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Generation: 1,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Service: &gatewayv2alpha1.Service{
				Name:       &serviceName,
				Port:       &servicePort,
				Host:       &host,
				IsExternal: &isExernal,
			},
			Auth: &gatewayv2alpha1.AuthStrategy{
				Name:   &authStrategy,
				Config: nil,
			},
			Gateway: &gateway,
		},
	}
}

func getAPIReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &controllers.ApiReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Api"),
	}
}

type testSuite struct {
	mgr manager.Manager
}

func getTestSuite(objects ...runtime.Object) *testSuite {
	err := gatewayv2alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = networkingv1alpha3.AddToScheme(scheme.Scheme)
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
