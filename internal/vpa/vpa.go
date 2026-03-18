package vpa

import (
	"context"
	"fmt"

	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	vpaCRDName           = "verticalpodautoscalers.autoscaling.k8s.io"
	vpaName              = "api-gateway-controller-manager-vpa"
	vpaNamespace         = "kyma-system"
	targetDeploymentName = "api-gateway-controller-manager"

	minAllowedCPU    = "10m"
	minAllowedMemory = "64Mi"
	maxAllowedCPU    = "10000m"
	maxAllowedMemory = "16Gi"
)

var vpaKey = types.NamespacedName{Name: vpaName, Namespace: vpaNamespace}

type Reconciler struct {
	client.Client
}

func NewReconciler(c client.Client) *Reconciler {
	return &Reconciler{Client: c}
}

// Reconcile creates or updates the VPA when the VPA CRD is installed.
// It deletes the VPA when isInDeletion is true, and skips entirely when the CRD is absent.
func (r *Reconciler) Reconcile(ctx context.Context, isInDeletion bool) error {
	log := ctrl.Log.WithName("vpa-reconciler")

	installed, err := r.isVPACRDInstalled(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if VPA CRD is installed: %w", err)
	}
	if !installed {
		log.Info("VPA CRD not installed, skipping")
		return nil
	}

	if isInDeletion {
		if err := client.IgnoreNotFound(r.Delete(ctx, desiredVPA())); err != nil {
			return fmt.Errorf("failed to delete VPA: %w", err)
		}
		log.Info("VPA deleted", "name", vpaName)
		return nil
	}

	existing := &vpav1.VerticalPodAutoscaler{}
	if err := r.Get(ctx, vpaKey, existing); errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredVPA()); err != nil {
			return fmt.Errorf("failed to create VPA: %w", err)
		}
		log.Info("VPA created", "name", vpaName)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get VPA: %w", err)
	}

	desired := desiredVPA()
	if equality.Semantic.DeepEqual(existing.Spec, desired.Spec) {
		log.Info("VPA already up to date, skipping update", "name", vpaName)
		return nil
	}

	existing.Spec = desired.Spec
	if err := r.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update VPA: %w", err)
	}
	log.Info("VPA updated", "name", vpaName)
	return nil
}

func (r *Reconciler) isVPACRDInstalled(ctx context.Context) (bool, error) {
	err := r.Get(ctx, types.NamespacedName{Name: vpaCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func getModuleLabels() map[string]string {
	return map[string]string{
		"kyma-project.io/module":              "api-gateway",
		"operator.kyma-project.io/managed-by": "kyma",
	}
}

func desiredVPA() *vpav1.VerticalPodAutoscaler {
	controlledResources := []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}

	containerPolicy := func(name string) vpav1.ContainerResourcePolicy {
		return vpav1.ContainerResourcePolicy{
			ContainerName: name,
			MinAllowed: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(minAllowedCPU),
				corev1.ResourceMemory: resource.MustParse(minAllowedMemory),
			},
			MaxAllowed: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(maxAllowedCPU),
				corev1.ResourceMemory: resource.MustParse(maxAllowedMemory),
			},
			ControlledResources: &controlledResources,
			ControlledValues:    ptr.To(vpav1.ContainerControlledValuesRequestsAndLimits),
		}
	}

	return &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vpaName,
			Namespace: vpaNamespace,
			Labels:    getModuleLabels(),
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       targetDeploymentName,
			},
			UpdatePolicy: &vpav1.PodUpdatePolicy{
				UpdateMode:  ptr.To(vpav1.UpdateModeInPlaceOrRecreate),
				MinReplicas: ptr.To(int32(1)),
			},
			ResourcePolicy: &vpav1.PodResourcePolicy{
				ContainerPolicies: []vpav1.ContainerResourcePolicy{
					containerPolicy("manager"),
					containerPolicy("init"),
				},
			},
		},
	}
}
