package vpa

import (
	"context"
	"fmt"

	"github.com/kyma-project/api-gateway/internal/processing"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/util/retry"
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

	vpacName = vpaName + "-manager"
)

var vpaKey = types.NamespacedName{Name: vpaName, Namespace: vpaNamespace}
var vpacKey = types.NamespacedName{Name: vpacName, Namespace: vpaNamespace}

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
		if err := r.patchVPACheckpointLabels(ctx); err != nil {
			return err
		}
		return nil
	}

	existing.Spec = desired.Spec
	if err := r.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update VPA: %w", err)
	}
	log.Info("VPA updated", "name", vpaName)
	if err := r.patchVPACheckpointLabels(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) isVPACRDInstalled(ctx context.Context) (bool, error) {
	err := r.Get(ctx, types.NamespacedName{Name: vpaCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func (r *Reconciler) patchVPACheckpointLabels(ctx context.Context) error {
	log := ctrl.Log.WithName("vpa-reconciler")
	// Track whether the final retry attempt found the checkpoint and patched labels
	checkpointFound := false
	labelsPatched := false
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Reset for this attempt so final logs represent the final retry outcome
		checkpointFound = false
		labelsPatched = false

		checkpoint := &vpav1.VerticalPodAutoscalerCheckpoint{}
		if err := r.Get(ctx, vpacKey, checkpoint); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		checkpointFound = true

		labels, labelsChanged := ensureModuleLabels(checkpoint.GetLabels())
		if !labelsChanged {
			return nil
		}

		patch := client.MergeFrom(checkpoint.DeepCopy())
		checkpoint.SetLabels(labels)
		if err := r.Patch(ctx, checkpoint, patch); err != nil {
			return err
		}
		labelsPatched = true
		return nil
	}); err != nil {
		return fmt.Errorf("failed to patch VPA checkpoint labels: %w", err)
	}

	if !checkpointFound {
		log.Info("VPA checkpoint not found yet, skipping label patch", "name", vpacName)
		return nil
	}
	if labelsPatched {
		log.Info("VPA checkpoint labels patched", "name", vpacName)
		return nil
	}
	log.Info("VPA checkpoint labels already up to date", "name", vpacName)
	return nil
}

// Ensures the module labels exist in the given map and returns if any change was needed
func ensureModuleLabels(labels map[string]string) (map[string]string, bool) {
	moduleLabels := getModuleLabels()
	// Copy the labels map so we don't mutate the original object state before patching
	patchedLabels := make(map[string]string, len(labels)+len(moduleLabels))

	for name, value := range labels {
		patchedLabels[name] = value
	}
	labelsChanged := false
	for name, value := range moduleLabels {
		if existing, ok := patchedLabels[name]; !ok || existing != value {
			patchedLabels[name] = value
			labelsChanged = true
		}
	}

	return patchedLabels, labelsChanged
}

func getModuleLabels() map[string]string {
	return map[string]string{
		processing.ModuleLabelKey:       processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sComponentLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:    processing.ApiGatewayLabelValue,
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
