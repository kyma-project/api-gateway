package api_gateway

import (
	"context"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("API-Gateway reconciliation", func() {
	It("Should add finalizer when API-Gateway CR is not marked for deletion", func() {
		// given
		apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}

		c := createFakeClient(apiGatewayCR)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		status := reconciliation.Reconcile(context.TODO(), apiGatewayCR)

		// then
		Expect(status.IsReady()).Should(BeTrue())
		Expect(c.Get(context.TODO(), client.ObjectKey{Name: apiGatewayCR.GetName(), Namespace: apiGatewayCR.GetNamespace()}, apiGatewayCR)).Should(Succeed())
		Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
	})

	It("Should remove finalizer on API-Gateway CR deletion if there are no blocking resources", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{ApiGatewayFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}
		c := createFakeClient(apiGatewayCR)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		status := reconciliation.Reconcile(context.TODO(), apiGatewayCR)

		// then
		Expect(status.IsReady()).Should(BeTrue())
		Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(BeEmpty())
	})

	It("Should not remove finalizer on API-Gateway CR deletion if there is any APIRule on cluster", func() {
		// given
		now := metav1.NewTime(time.Now())
		apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{ApiGatewayFinalizer},
		},
			Spec: operatorv1alpha1.APIGatewaySpec{},
		}
		apiRule := &gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "default",
			},
		}
		c := createFakeClient(apiGatewayCR, apiRule)
		reconciliation := Reconciliation{
			Client: c,
		}

		// when
		status := reconciliation.Reconcile(context.TODO(), apiGatewayCR)

		// then
		Expect(status.IsWarning()).Should(BeTrue())
		Expect(status.NestedError().Error()).To(Equal("could not delete API-Gateway module instance since there are 1 APIRule(s) present that block its deletion"))
		Expect(status.Description()).To(Equal("There are APIRule(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
		Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
	})
})
