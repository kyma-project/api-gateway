/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package external

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/reconciliations/externalgateway"
)

const (
	externalGatewayFinalizer      = "externalgateways.gateway.kyma-project.io/finalizer"
	defaultReconciliationInterval = 1 * time.Hour // Periodic reconciliation for drift detection and self-healing
)

// ExternalGatewayReconciler reconciles a ExternalGateway object
type ExternalGatewayReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// NewExternalGatewayReconciler creates a new ExternalGatewayReconciler
func NewExternalGatewayReconciler(mgr manager.Manager) *ExternalGatewayReconciler {
	log := mgr.GetLogger().WithName("externalgateway-controller")
	return &ExternalGatewayReconciler{
		Client: mgr.GetClient(),
		Log:    log,
		Scheme: mgr.GetScheme(),
	}
}

// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=externalgateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=externalgateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=externalgateways/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch,resourceNames=shoot-info,namespace=kube-system
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=dns.gardener.cloud,resources=dnsentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert.gardener.cloud,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ExternalGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("externalgateway", req.NamespacedName)

	// Fetch the ExternalGateway instance
	externalGateway := &externalv1alpha1.ExternalGateway{}
	if err := r.Get(ctx, req.NamespacedName, externalGateway); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ExternalGateway resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ExternalGateway")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !externalGateway.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, log, externalGateway)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(externalGateway, externalGatewayFinalizer) {
		controllerutil.AddFinalizer(externalGateway, externalGatewayFinalizer)
		if err := r.Update(ctx, externalGateway); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer")
	}

	// Update status to Processing
	if err := r.updateStatus(ctx, externalGateway, externalv1alpha1.Processing, "Reconciling resources"); err != nil {
		// Log conflict errors but don't fail reconciliation - status will be updated on next reconcile
		if apierrors.IsConflict(err) {
			log.Info("Status update conflict, will retry on next reconciliation")
		} else {
			log.Error(err, "Failed to update status to Processing")
			return ctrl.Result{}, err
		}
	}

	// Reconcile all resources
	if err := r.reconcileResources(ctx, log, externalGateway); err != nil {
		log.Error(err, "Failed to reconcile resources")
		if statusErr := r.updateStatus(ctx, externalGateway, externalv1alpha1.Error, err.Error()); statusErr != nil {
			log.Error(statusErr, "Failed to update error status")
		}
		return ctrl.Result{}, err
	}

	// Update status to Ready
	if err := r.updateStatus(ctx, externalGateway, externalv1alpha1.Ready, "All resources reconciled successfully"); err != nil {
		// Log conflict errors but don't fail reconciliation - status will be updated on next reconcile
		if apierrors.IsConflict(err) {
			log.Info("Status update conflict, will retry on next reconciliation")
		} else {
			log.Error(err, "Failed to update status to Ready")
			return ctrl.Result{}, err
		}
	}

	log.Info("Reconciliation completed successfully")
	return ctrl.Result{RequeueAfter: defaultReconciliationInterval}, nil
}

// reconcileResources orchestrates the creation/update of all required resources
func (r *ExternalGatewayReconciler) reconcileResources(ctx context.Context, log logr.Logger, external *externalv1alpha1.ExternalGateway) error {
	// Warn if multiple regions are specified - only the first will be used
	if len(external.Spec.Regions) > 1 {
		log.Info("WARNING: Multiple regions specified, only the first region will be used",
			"specifiedRegions", external.Spec.Regions,
			"usedRegion", external.Spec.Regions[0])
	}

	// Build internal domain
	internalDomain, err := r.buildInternalDomain(ctx, external.Spec.InternalDomain.KymaSubdomain)
	if err != nil {
		return fmt.Errorf("failed to build internal domain: %w", err)
	}

	// Check if Gardener is available
	_, gardenerErr := dependencies.Gardener().AreAvailable(ctx, r.Client)
	isGardenerAvailable := gardenerErr == nil

	// Reconcile Gardener resources if available
	if isGardenerAvailable {
		if err := externalgateway.ReconcileDNSEntry(ctx, r.Client, external, internalDomain); err != nil {
			return fmt.Errorf("failed to reconcile DNSEntry: %w", err)
		}

		if err := externalgateway.ReconcileCertificate(ctx, r.Client, external, internalDomain); err != nil {
			return fmt.Errorf("failed to reconcile Certificate: %w", err)
		}
	} else {
		log.Info("Gardener not available, skipping DNSEntry and Certificate reconciliation")
	}

	// Reconcile CA Secret (copy from app namespace to istio-system)
	if err := externalgateway.ReconcileCASecret(ctx, r.Client, external); err != nil {
		return fmt.Errorf("failed to reconcile CA Secret: %w", err)
	}

	// Reconcile Istio Gateway
	if err := externalgateway.ReconcileGateway(ctx, r.Client, r.Scheme, external, internalDomain); err != nil {
		return fmt.Errorf("failed to reconcile Gateway: %w", err)
	}

	// Resolve certificate subjects from regions ConfigMap
	certSubjects, err := externalgateway.ResolveRegionCertSubjects(ctx, r.Client, external)
	if err != nil {
		return fmt.Errorf("failed to resolve certificate subjects: %w", err)
	}

	// Reconcile EnvoyFilters
	if err := externalgateway.ReconcileXFCCSanitizationFilter(ctx, r.Client, external); err != nil {
		return fmt.Errorf("failed to reconcile XFCC sanitization filter: %w", err)
	}

	if err := externalgateway.ReconcileCertValidationFilter(ctx, r.Client, external, certSubjects); err != nil {
		return fmt.Errorf("failed to reconcile certificate validation filter: %w", err)
	}

	return nil
}

// handleDeletion cleans up all resources when ExternalGateway is deleted
func (r *ExternalGatewayReconciler) handleDeletion(ctx context.Context, log logr.Logger, external *externalv1alpha1.ExternalGateway) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(external, externalGatewayFinalizer) {
		return ctrl.Result{}, nil
	}

	log.Info("Handling deletion, cleaning up resources")

	gatewayName := external.GatewayName()
	namespace := external.Namespace

	// Check if Gardener is available
	_, gardenerErr := dependencies.Gardener().AreAvailable(ctx, r.Client)
	isGardenerAvailable := gardenerErr == nil

	// Delete Gateway
	if err := externalgateway.DeleteGateway(ctx, r.Client, namespace, gatewayName); err != nil {
		log.Error(err, "Failed to delete Gateway")
		return ctrl.Result{}, err
	}

	// Delete EnvoyFilters
	if err := externalgateway.DeleteXFCCSanitizationFilter(ctx, r.Client, gatewayName); err != nil {
		log.Error(err, "Failed to delete XFCC sanitization EnvoyFilter")
		return ctrl.Result{}, err
	}

	if err := externalgateway.DeleteCertValidationFilter(ctx, r.Client, gatewayName); err != nil {
		log.Error(err, "Failed to delete certificate validation EnvoyFilter")
		return ctrl.Result{}, err
	}

	// Delete CA Secret
	if err := externalgateway.DeleteCASecret(ctx, r.Client, gatewayName); err != nil {
		log.Error(err, "Failed to delete CA Secret")
		return ctrl.Result{}, err
	}

	// Delete Gardener resources if available
	if isGardenerAvailable {
		if err := externalgateway.DeleteDNSEntry(ctx, r.Client, gatewayName); err != nil {
			log.Error(err, "Failed to delete DNSEntry")
			return ctrl.Result{}, err
		}

		if err := externalgateway.DeleteCertificate(ctx, r.Client, gatewayName); err != nil {
			log.Error(err, "Failed to delete Certificate")
			return ctrl.Result{}, err
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(external, externalGatewayFinalizer)
	if err := r.Update(ctx, external); err != nil {
		log.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	log.Info("Successfully deleted all resources and removed finalizer")
	return ctrl.Result{}, nil
}

// updateStatus updates the status of the ExternalGateway CR
func (r *ExternalGatewayReconciler) updateStatus(ctx context.Context, external *externalv1alpha1.ExternalGateway, state externalv1alpha1.State, description string) error {
	// Get fresh copy to avoid conflicts
	freshUgw := &externalv1alpha1.ExternalGateway{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(external), freshUgw); err != nil {
		return err
	}

	// Update status fields
	freshUgw.Status.State = state
	freshUgw.Status.Description = description
	freshUgw.Status.LastProcessedTime = metav1.Now()

	return r.Status().Update(ctx, freshUgw)
}

// SetupWithManager sets up the controller with the Manager
func (r *ExternalGatewayReconciler) SetupWithManager(mgr ctrl.Manager, rateLimiterConfig controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&externalv1alpha1.ExternalGateway{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(rateLimiterConfig),
		}).
		Complete(r)
}
