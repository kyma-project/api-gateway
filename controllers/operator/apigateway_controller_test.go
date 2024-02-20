package operator

import (
	"context"
	goerrors "errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API-Gateway Controller", func() {
	Context("Reconcile", func() {
		It("Should not return an error when CR was not found", func() {
			// given
			c := createFakeClient()
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("Should add finalizer when API-Gateway CR is not marked for deletion", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})

		It("Should set status to Ready when API-Gateway reconciliation succeed", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:       apiGatewayCRName,
					Namespace:  testNamespace,
					Finalizers: []string{ApiGatewayFinalizer},
				},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).Should(Equal(operatorv1alpha1.Ready))
		})

		It("Should set status Ready on the older APIGateway CR when there are two in the cluster", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					UID:       "1",
					// 11 May 2017
					CreationTimestamp: metav1.Unix(1494505756, 0),
				},
			}

			secondApiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "default2",
					Namespace:         testNamespace,
					UID:               "2",
					CreationTimestamp: metav1.Now(),
				},
			}

			c := createFakeClient(apiGatewayCR, secondApiGatewayCR)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Ready))
		})

		It("Should set an error status and do not requeue an APIGateway CR when an older APIGateway CR is present", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					UID:       "1",
					// 11 May 2017
					CreationTimestamp: metav1.Unix(1494505756, 0),
				},
			}

			secondApiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:              fmt.Sprintf("%s-2", apiGatewayCRName),
					Namespace:         testNamespace,
					UID:               "2",
					CreationTimestamp: metav1.Now(),
				},
			}

			c := createFakeClient(apiGatewayCR, secondApiGatewayCR)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: fmt.Sprintf("%s-2", apiGatewayCRName)}})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(secondApiGatewayCR), secondApiGatewayCR)).Should(Succeed())
			Expect(secondApiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(secondApiGatewayCR.Status.Description).To(Equal(fmt.Sprintf("stopped APIGateway CR reconciliation: only APIGateway CR %s reconciles the module", apiGatewayCRName)))

		})

		It("Should set an error status and requeue an APIGateway CR when is unable to list Istio CRs", func() {
			// given
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiGatewayCRName,
					Namespace: testNamespace,
					UID:       "1",
					// 11 May 2017
					CreationTimestamp: metav1.Unix(1494505756, 0),
				},
			}

			c := createFakeClient(apiGatewayCR)
			fc := &shouldFailClient{c, true}
			agr := &APIGatewayReconciler{
				Client:               fc,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).To(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(apiGatewayCR.Status.Description).To(Equal("Unable to list APIGateway CRs"))
		})

		It("Should delete API-Gateway CR if there are no blocking resources", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
				Name:              apiGatewayCRName,
				Namespace:         testNamespace,
				DeletionTimestamp: &now,
				Finalizers:        []string{ApiGatewayFinalizer},
			},
			}

			c := createFakeClient(apiGatewayCR)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(errors.IsNotFound(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR))).To(BeTrue())
		})

		It("Should not delete API-Gateway CR if there are any APIRules on cluster", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
				Name:              apiGatewayCRName,
				Namespace:         testNamespace,
				DeletionTimestamp: &now,
				Finalizers:        []string{ApiGatewayFinalizer},
			},
			}
			apiRule := &gatewayv1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-rule",
					Namespace: "default",
				},
			}
			oryRule := &oryv1alpha1.Rule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ory-rule",
					Namespace: "default",
				},
			}

			c := createFakeClient(apiGatewayCR, apiRule, oryRule)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("could not delete API-Gateway CR since there are APIRule(s) that block its deletion"))
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})

		It("Should not delete API-Gateway CR if there are any ORY Oathkeeper Rules on cluster", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{
				Name:              apiGatewayCRName,
				Namespace:         testNamespace,
				DeletionTimestamp: &now,
				Finalizers:        []string{ApiGatewayFinalizer},
			},
			}
			oryRule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&oryv1alpha1.Rule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ory-rule",
					Namespace: "default",
				},
			})

			r := &unstructured.Unstructured{Object: oryRule}
			r.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "oathkeeper.ory.sh",
				Version: "v1alpha1",
				Kind:    "rule",
			})

			Expect(err).ToNot(HaveOccurred())

			c := createFakeClient(apiGatewayCR, r)
			agr := &APIGatewayReconciler{
				Client:               c,
				Scheme:               getTestScheme(),
				log:                  logr.Discard(),
				oathkeeperReconciler: oathkeeperReconcilerWithoutVerification{},
			}

			// when
			result, err := agr.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: apiGatewayCRName}})

			// then
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("could not delete API-Gateway CR since there are ORY Oathkeeper Rule(s) that block its deletion"))
			Expect(result).Should(Equal(reconcile.Result{}))

			Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})
	})
})

type shouldFailClient struct {
	client.Client
	FailOnList bool
}

func (p *shouldFailClient) List(ctx context.Context, list client.ObjectList, _ ...client.ListOption) error {
	if p.FailOnList {
		return goerrors.New("fail on purpose")
	}
	return p.Client.List(ctx, list)
}
