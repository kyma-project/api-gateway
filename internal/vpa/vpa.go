package vpa

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	vpaCRDName           = "verticalpodautoscalers.autoscaling.k8s.io"
	vpaName              = "api-gateway-controller-manager-vpa"
	vpaNamespace         = "kyma-system"
	targetDeploymentName = "api-gateway-controller-manager"
)

type Reconciler struct {
	client.Client
}

func NewReconciler(c client.Client) *Reconciler {
	return &Reconciler{Client: c}
}

// Reconcile creates the VPA if the VPA CRD is installed, or does nothing if it is not.
// On deletion of the APIGateway CR, it deletes the VPA.
func (r *Reconciler) Reconcile(ctx context.Context, isInDeletion bool) error {
	log := ctrl.Log.WithName("vpa-reconciler")

	crdInstalled, err := r.isVPACRDInstalled(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if VPA CRD is installed: %w", err)
	}

	if !crdInstalled {
		log.Info("VPA CRD is not installed on the cluster, skipping VPA reconciliation")
		return nil
	}

	if isInDeletion {
		return r.deleteVPA(ctx)
	}

	return r.createOrUpdateVPA(ctx)
}

func (r *Reconciler) isVPACRDInstalled(ctx context.Context) (bool, error) {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Get(ctx, types.NamespacedName{Name: vpaCRDName}, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Reconciler) deleteVPA(ctx context.Context) error {
	log := ctrl.Log.WithName("vpa-reconciler")

	vpa := r.buildVPAObject()
	if err := r.Delete(ctx, vpa); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete VPA: %w", err)
	}

	log.Info("Successfully deleted VPA", "name", vpaName, "namespace", vpaNamespace)
	return nil
}

func (r *Reconciler) createOrUpdateVPA(ctx context.Context) error {
	log := ctrl.Log.WithName("vpa-reconciler")

	vpa := r.buildVPAObject()
	existing := r.buildVPAObject()

	err := r.Get(ctx, types.NamespacedName{Name: vpaName, Namespace: vpaNamespace}, existing)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get existing VPA: %w", err)
		}
		r.setVPASpec(vpa)
		if createErr := r.Create(ctx, vpa); createErr != nil {
			return fmt.Errorf("failed to create VPA: %w", createErr)
		}
		log.Info("Successfully created VPA", "name", vpaName, "namespace", vpaNamespace)
		return nil
	}

	r.setVPASpec(existing)
	if updateErr := r.Update(ctx, existing); updateErr != nil {
		return fmt.Errorf("failed to update VPA: %w", updateErr)
	}
	log.Info("Successfully updated VPA", "name", vpaName, "namespace", vpaNamespace)
	return nil
}

func (r *Reconciler) buildVPAObject() *unstructured.Unstructured {
	vpa := &unstructured.Unstructured{}
	vpa.SetGroupVersionKind(vpaGVK())
	vpa.SetName(vpaName)
	vpa.SetNamespace(vpaNamespace)
	return vpa
}

func (r *Reconciler) setVPASpec(vpa *unstructured.Unstructured) {
	vpa.Object["spec"] = map[string]interface{}{
		"targetRef": map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"name":       targetDeploymentName,
		},
		"updatePolicy": map[string]interface{}{
			"updateMode":  "InPlaceOrRecreate",
			"minReplicas": int64(1),
		},
		"resourcePolicy": map[string]interface{}{
			"containerPolicies": []interface{}{
				map[string]interface{}{
					"containerName": "manager",
					"minAllowed": map[string]interface{}{
						"cpu":    "10m",
						"memory": "64Mi",
					},
					"maxAllowed": map[string]interface{}{
						"cpu":    "10000m",
						"memory": "16Gi",
					},
					"controlledResources": []interface{}{"cpu", "memory"},
					"controlledValues":    "RequestsAndLimits",
				},
				map[string]interface{}{
					"containerName": "init",
					"minAllowed": map[string]interface{}{
						"cpu":    "10m",
						"memory": "64Mi",
					},
					"maxAllowed": map[string]interface{}{
						"cpu":    "10000m",
						"memory": "16Gi",
					},
					"controlledResources": []interface{}{"cpu", "memory"},
					"controlledValues":    "RequestsAndLimits",
				},
			},
		},
	}
}

func vpaGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "autoscaling.k8s.io",
		Version: "v1",
		Kind:    "VerticalPodAutoscaler",
	}
}
