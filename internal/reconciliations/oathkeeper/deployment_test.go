package oathkeeper_test

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Oathkeeper reconciliation", func() {

	It("Successfully reconcile Oathkeeper", func() {
		apiGateway := createApiGateway(nil)
		k8sClient := createFakeClient(apiGateway)
		status := oathkeeper.ReconcileOathkeeper(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue())
	})

})

func createApiGateway(enableKymaGateway *bool) *v1alpha1.APIGateway {
	return &v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.APIGatewaySpec{
			EnableKymaGateway: enableKymaGateway,
		},
	}

}
