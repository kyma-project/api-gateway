package controller

import (
	"context"
	"fmt"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("status", func() {

	Context("UpdateApiGatewayStatus", func() {

		It("Should Update APIGateway CR state and set description", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}

			k8sClient := createFakeClient(&cr)
			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ErrorStatus(fmt.Errorf("test error"), "test description"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(cr.Status.Description).To(Equal("test description"))
		})

		It("Should return error if update status fails", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}

			k8sClient := fake.NewClientBuilder().Build()
			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ReadyStatus(operatorv1alpha1.ReadyCondition()))

			// then
			Expect(err).To(HaveOccurred())
		})

		It("Should contain condition that is not nil and with expected value", func() {
			// given
			condition := metav1.Condition{Type: operatorv1alpha1.ConditionTypeReady, Status: metav1.ConditionFalse, Reason: "test"}

			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ErrorStatus(fmt.Errorf(""), "", condition))

			// then
			Expect(err).To(BeNil())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(readyCond.Status).To(Equal(metav1.ConditionFalse))
		})

		It("Should preserve ObservedGeneration while Processing when condition type already exists", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 5},
				Status: operatorv1alpha1.APIGatewayStatus{
					Conditions: []metav1.Condition{{Type: operatorv1alpha1.ConditionTypeReady, Status: metav1.ConditionTrue, ObservedGeneration: 3}},
				},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ProcessingStatus(operatorv1alpha1.ProcessingCondition()))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Processing))
			Expect(readyCond.Status).To(Equal(metav1.ConditionUnknown))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ProcessingCondition().Reason))
			Expect(readyCond.ObservedGeneration).To(Equal(int64(3)))
		})

		It("Should keep default ObservedGeneration while Processing when previous condition type does not exist", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 5},
				Status: operatorv1alpha1.APIGatewayStatus{
					Conditions: []metav1.Condition{{Type: "Other", Status: metav1.ConditionTrue, ObservedGeneration: 9}},
				},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ProcessingStatus(operatorv1alpha1.ProcessingCondition()))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Processing))
			Expect(readyCond.Status).To(Equal(metav1.ConditionUnknown))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ProcessingCondition().Reason))
			Expect(readyCond.ObservedGeneration).To(Equal(int64(0)))
		})

		It("Should set ObservedGeneration to current generation for non-Processing status", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 8},
				Status: operatorv1alpha1.APIGatewayStatus{
					Conditions: []metav1.Condition{{Type: operatorv1alpha1.ConditionTypeReady, Status: metav1.ConditionUnknown, ObservedGeneration: 2}},
				},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ReadyStatus(operatorv1alpha1.ReadyCondition()))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Ready))
			Expect(readyCond.Status).To(Equal(metav1.ConditionTrue))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ReadyCondition().Reason))
			Expect(readyCond.ObservedGeneration).To(Equal(int64(8)))
		})

		It("Should set Ready condition status and reason for Warning state", func() {
			// given
			cr := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 4}}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr,
				WarningStatus(fmt.Errorf("older CR exists"), "older CR exists", operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonOlderCRExists, "older CR exists")))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Warning))
			Expect(readyCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ReasonOlderCRExists))
		})

		It("Should set Ready condition status and reason for Error state", func() {
			// given
			cr := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 4}}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr,
				ErrorStatus(fmt.Errorf("boom"), "boom", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "boom")))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(cr.Status.State).To(Equal(operatorv1alpha1.Error))
			Expect(readyCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ReasonReconcileFailed))
		})

		It("Should replace an existing Ready condition instead of duplicating it", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 9},
				Status: operatorv1alpha1.APIGatewayStatus{
					Conditions: []metav1.Condition{
						{Type: operatorv1alpha1.ConditionTypeReady, Status: metav1.ConditionUnknown, Reason: "OldReason", ObservedGeneration: 1},
						{Type: "Other", Status: metav1.ConditionTrue},
					},
				},
			}
			k8sClient := createFakeClient(&cr)

			// when
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr, ReadyStatus(operatorv1alpha1.ReadyCondition()))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())
			var readyConditions []metav1.Condition
			for _, condition := range cr.Status.Conditions {
				if condition.Type == operatorv1alpha1.ConditionTypeReady {
					readyConditions = append(readyConditions, condition)
				}
			}
			Expect(readyConditions).To(HaveLen(1))
			Expect(readyConditions[0].Reason).To(Equal(operatorv1alpha1.ReadyCondition().Reason))
			Expect(readyConditions[0].Status).To(Equal(metav1.ConditionTrue))
		})

		It("Should preserve unrelated conditions and write passed conditions", func() {
			// given
			cr := operatorv1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Generation: 7},
				Status: operatorv1alpha1.APIGatewayStatus{
					Conditions: []metav1.Condition{
						{Type: operatorv1alpha1.ConditionTypeNetworkPolicy, Status: metav1.ConditionTrue, Reason: operatorv1alpha1.ReasonNetworkPolicyReconcileSucceeded, ObservedGeneration: 3},
						{Type: operatorv1alpha1.ConditionTypeReady, Status: metav1.ConditionTrue, Reason: operatorv1alpha1.ReasonReconcileSucceeded, ObservedGeneration: 3},
					},
				},
			}
			k8sClient := createFakeClient(&cr)

			// when
			msg := "Kyma Gateway deletion blocked because of the existing custom resources: blocking-api-rule"
			err := UpdateApiGatewayStatus(context.Background(), k8sClient, &cr,
				WarningStatus(fmt.Errorf("blocked"), "blocked",
					operatorv1alpha1.DeletionBlockedExistingResourcesCondition(msg),
					operatorv1alpha1.KymaGatewayDeletionBlockedCondition(msg)))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: "test"}, &cr)).To(Succeed())

			readyCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeReady)
			Expect(readyCond).ToNot(BeNil())
			Expect(readyCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(readyCond.Reason).To(Equal(operatorv1alpha1.ReasonDeletionBlockedExistingResources))

			gatewayCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeKymaGateway)
			Expect(gatewayCond).ToNot(BeNil())
			Expect(gatewayCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(gatewayCond.Reason).To(Equal(operatorv1alpha1.ReasonKymaGatewayDeletionBlocked))
			Expect(gatewayCond.Message).To(Equal(msg))
			Expect(gatewayCond.ObservedGeneration).To(Equal(int64(7)))

			networkCond := meta.FindStatusCondition(cr.Status.Conditions, operatorv1alpha1.ConditionTypeNetworkPolicy)
			Expect(networkCond).ToNot(BeNil())
			Expect(networkCond.Status).To(Equal(metav1.ConditionTrue))
			Expect(networkCond.ObservedGeneration).To(Equal(int64(3)))
		})
	})

	Context("ToAPIGatewayStatus", func() {

		It("Should return Error with description set", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description")

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Error))
			Expect(apiGatewayStatus.Description).To(Equal("test description"))
		})

		It("Should return Warning with description set", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description")

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Warning))
			Expect(apiGatewayStatus.Description).To(Equal("test description"))
		})

		It("Should return Ready with default description", func() {
			// given
			status := ReadyStatus(metav1.Condition{})

			// when
			apiGatewayStatus, err := status.ToAPIGatewayStatus()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(apiGatewayStatus.State).To(Equal(operatorv1alpha1.Ready))
			Expect(apiGatewayStatus.Description).To(Equal("Successfully reconciled"))
		})

	})

	Context("IsError", func() {
		It("Should return true if status is Error", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description")

			// when
			isError := status.IsError()

			// then
			Expect(isError).To(BeTrue())
		})
		It("Should return false if status is not Error", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description")

			// when
			isError := status.IsError()

			// then
			Expect(isError).To(BeFalse())
		})
	})

	Context("IsWarning", func() {
		It("Should return true if status is Warning", func() {
			// given
			status := WarningStatus(fmt.Errorf("test error"), "test description")

			// when
			isWarning := status.IsWarning()

			// then
			Expect(isWarning).To(BeTrue())
		})
		It("Should return false if status is not Warning", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description")

			// when
			isWarning := status.IsWarning()

			// then
			Expect(isWarning).To(BeFalse())
		})
	})

	Context("IsReady", func() {
		It("Should return true if status is Ready", func() {
			// given
			status := ReadyStatus(metav1.Condition{})

			// when
			result := status.IsReady()

			// then
			Expect(result).To(BeTrue())
		})
		It("Should return false if status is not Ready", func() {
			// given
			status := ErrorStatus(fmt.Errorf("test error"), "test description")

			// when
			result := status.IsReady()

			// then
			Expect(result).To(BeFalse())
		})
	})
})
