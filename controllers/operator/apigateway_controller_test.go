package operator

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	"github.com/kyma-project/api-gateway/internal/conditions"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("API-Gateway Controller", func() {
	Context("Reconcile", func() {
		It("Should requeue if the reconciliation was successful", func() {
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
			Expect(result).Should(Equal(reconcile.Result{
				RequeueAfter: defaultApiGatewayReconciliationInterval,
			}))
		})

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
			Expect(result).Should(Equal(reconcile.Result{
				RequeueAfter: defaultApiGatewayReconciliationInterval,
			}))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))
		})

		It("Should set status to Ready and add condition when API-Gateway reconciliation succeed", func() {
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
			Expect(result).Should(Equal(reconcile.Result{
				RequeueAfter: defaultApiGatewayReconciliationInterval,
			}))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).Should(Equal(operatorv1alpha1.Ready))
			Expect(apiGatewayCR.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(conditions.ReconcileSucceeded.Condition().Type),
				"Status": Equal(metav1.ConditionTrue),
			})))
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
			Expect(result.RequeueAfter).To(Equal(time.Hour * 1))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Ready))
		})

		It("Should set an error status with condition and do not requeue an APIGateway CR when an older APIGateway CR is present", func() {
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
			Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(secondApiGatewayCR), secondApiGatewayCR)).Should(Succeed())
			Expect(secondApiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(secondApiGatewayCR.Status.Description).To(Equal(fmt.Sprintf("stopped APIGateway CR reconciliation: only APIGateway CR %s reconciles the module", apiGatewayCRName)))
			Expect(secondApiGatewayCR.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(conditions.OlderCRExists.Condition().Type),
				"Status": Equal(metav1.ConditionFalse),
			})))

		})

		It("Should set an error status with condition and requeue an APIGateway CR when is unable to list Istio CRs", func() {
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
			Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(apiGatewayCR.Status.Description).To(Equal("Unable to list APIGateway CRs"))
			Expect(apiGatewayCR.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(conditions.ReconcileFailed.Condition().Type),
				"Status": Equal(metav1.ConditionFalse),
			})))
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

			Expect(errors.IsNotFound(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR))).To(BeTrue())
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

			var rules []client.Object
			// we initialize more than 5 objects, so we validate if we show only 5 in a condition
			for i := 0; i < 6; i++ {
				apiRule := &gatewayv1beta1.APIRule{
					TypeMeta: metav1.TypeMeta{Kind: "APIRule"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("api-rule-%d", i),
						Namespace: "default",
					},
				}
				rules = append(rules, apiRule)
			}

			c := createFakeClient(append(rules, apiGatewayCR)...)
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

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))

			Expect(apiGatewayCR.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":    Equal(conditions.DeletionBlockedExistingResources.Condition().Type),
				"Message": Equal("API Gateway deletion blocked because of the existing custom resources: default/api-rule-0, default/api-rule-1, default/api-rule-2, default/api-rule-3, default/api-rule-4"),
				"Status":  Equal(metav1.ConditionFalse),
			})))
		})

		It("Should not delete API-Gateway CR if there are any ORY Oathkeeper Rules on cluster", func() {
			// given
			now := metav1.NewTime(time.Now())
			apiGatewayCR := &operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:              apiGatewayCRName,
					Namespace:         testNamespace,
					DeletionTimestamp: &now,
					Finalizers:        []string{ApiGatewayFinalizer},
				},
			}
			var rules []client.Object
			for i := 0; i < 6; i++ {
				// we initialize more than 5 objects, so we validate if we show only 5 in a condition
				r := &oryv1alpha1.Rule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("ory-rule-%d", i),
						Namespace: "default",
					},
				}

				rules = append(rules, r)
			}
			c := createFakeClient(append(rules, apiGatewayCR)...)
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

			Expect(c.Get(context.Background(), client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR)).Should(Succeed())
			Expect(apiGatewayCR.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayCR.Status.Description).To(Equal("There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"))
			Expect(apiGatewayCR.GetObjectMeta().GetFinalizers()).To(ContainElement(ApiGatewayFinalizer))

			Expect(apiGatewayCR.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":    Equal(conditions.DeletionBlockedExistingResources.Condition().Type),
				"Message": Equal("API Gateway deletion blocked because of the existing custom resources: default/ory-rule-0, default/ory-rule-1, default/ory-rule-2, default/ory-rule-3, default/ory-rule-4"),
				"Status":  Equal(metav1.ConditionFalse),
			})))
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
