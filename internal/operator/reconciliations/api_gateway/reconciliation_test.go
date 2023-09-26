package api_gateway

import (
	"context"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

const (
	resourceListPath string = "test/test_controlled_resource_list.yaml"
	testKey          string = "key"
	testValue        string = "value"
)

var _ = Describe("API-Gateway reconciliation", func() {
	It("should add finalizer when API-Gateway CR is not marked for deletion", func() {
		// given
		apiGatewayCR := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			ResourceVersion: "1",
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}

		c := createFakeClient(&apiGatewayCR)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		reconciledCR, err := reconciliation.Reconcile(context.TODO(), apiGatewayCR, resourceListPath)

		// then
		Expect(err).ShouldNot(HaveOccurred())
		Expect(reconciledCR.GetObjectMeta().GetFinalizers()).To(ContainElement(reconciliationFinalizer))
	})

	It("should remove finalizer on API-Gateway CR deletion if there are no user created resources", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			ResourceVersion:   "1",
			DeletionTimestamp: &now,
			Finalizers:        []string{reconciliationFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}

		c := createFakeClient(&apiGatewayCR)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		reconciledCR, err := reconciliation.Reconcile(context.TODO(), apiGatewayCR, resourceListPath)

		// then
		Expect(err).ShouldNot(HaveOccurred())
		Expect(reconciledCR.GetObjectMeta().GetFinalizers()).To(BeEmpty())
	})

	It("should remove finalizer on API-Gateway CR deletion if there are only user resources not referring Kyma default gateway", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			ResourceVersion:   "1",
			DeletionTimestamp: &now,
			Finalizers:        []string{reconciliationFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}
		unmanagedGateway := "unmanaged-namespace/gateway"
		apiRule := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "default",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Gateway: &unmanagedGateway,
			},
		}
		virtualService := &networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "virtual-service",
				Namespace: "default",
			},
			Spec: istioapiv1beta1.VirtualService{
				Gateways: []string{unmanagedGateway},
			},
		}
		c := createFakeClient(&apiGatewayCR, apiRule, virtualService)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		reconciledCR, err := reconciliation.Reconcile(context.TODO(), apiGatewayCR, resourceListPath)

		// then
		Expect(err).ShouldNot(HaveOccurred())
		Expect(reconciledCR.GetObjectMeta().GetFinalizers()).To(BeEmpty())
	})

	It("should not remove finalizer and block deletion of API-Gateway CR if there is APIRule referring Kyma default gateway", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			ResourceVersion:   "1",
			DeletionTimestamp: &now,
			Finalizers:        []string{reconciliationFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}
		managedGateway := dafaultGateway
		apiRule := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "default",
			},
			Spec: gatewayv1beta1.APIRuleSpec{
				Gateway: &managedGateway,
			},
		}
		c := createFakeClient(&apiGatewayCR, apiRule)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		reconciledCR, err := reconciliation.Reconcile(context.TODO(), apiGatewayCR, resourceListPath)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("could not delete API-Gateway module instance since there are 1 custom resource(s) present that block its deletion"))
		Expect(err.Description()).To(Equal("There are custom resource(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
		Expect(reconciledCR.GetObjectMeta().GetFinalizers()).ToNot(BeEmpty())
	})

	It("should not remove finalizer and block deletion of API-Gateway CR if there is VirtualService referring Kyma default gateway", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			ResourceVersion:   "1",
			DeletionTimestamp: &now,
			Finalizers:        []string{reconciliationFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}
		managedGateway := dafaultGateway
		virtualService := &networkingv1beta1.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "virtual-service",
				Namespace: "default",
			},
			Spec: istioapiv1beta1.VirtualService{
				Gateways: []string{managedGateway},
			},
		}
		c := createFakeClient(&apiGatewayCR, virtualService)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		reconciledCR, err := reconciliation.Reconcile(context.TODO(), apiGatewayCR, resourceListPath)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("could not delete API-Gateway module instance since there are 1 custom resource(s) present that block its deletion"))
		Expect(err.Description()).To(Equal("There are custom resource(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
		Expect(reconciledCR.GetObjectMeta().GetFinalizers()).ToNot(BeEmpty())
	})
})

func createFakeClient(objects ...client.Object) client.Client {
	err := operatorv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())
	err = gatewayv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())
	err = networkingv1alpha3.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())
	err = networkingv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).WithStatusSubresource(objects...).Build()
}
