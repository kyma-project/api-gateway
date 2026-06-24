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
	"errors"
	"fmt"
	"time"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
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
	defaultReconciliationInterval = 1 * time.Hour
	pendingRequeueInterval        = 10 * time.Second
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

	externalGateway := &externalv1alpha1.ExternalGateway{}
	if err := r.Get(ctx, req.NamespacedName, externalGateway); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ExternalGateway resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ExternalGateway")
		return ctrl.Result{}, err
	}

	if !externalGateway.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, log, externalGateway)
	}

	if !controllerutil.ContainsFinalizer(externalGateway, externalGatewayFinalizer) {
		controllerutil.AddFinalizer(externalGateway, externalGatewayFinalizer)
		if err := r.Update(ctx, externalGateway); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer")
	}

	if err := r.updateStatus(ctx, externalGateway, externalv1alpha1.Processing, "Reconciling resources",
		externalv1alpha1.ProcessingConditions(externalGateway.Generation)); err != nil {
		if apierrors.IsConflict(err) {
			log.Info("Status update conflict on Processing, will retry on next reconciliation")
		} else {
			log.Error(err, "Failed to update status to Processing")
			return ctrl.Result{}, err
		}
	}

	conditions, requeue, err := r.reconcileResources(ctx, log, externalGateway)
	if err != nil {
		log.Error(err, "Failed to reconcile resources")
		conditions = append(conditions, externalv1alpha1.ErrorCondition(externalGateway.Generation, err.Error()))
		if statusErr := r.updateStatus(ctx, externalGateway, externalv1alpha1.Error, err.Error(), conditions); statusErr != nil {
			log.Error(statusErr, "Failed to update error status")
		}
		return ctrl.Result{}, err
	}

	if requeue {
		conditions = append(conditions, externalv1alpha1.WaitingCondition(externalGateway.Generation))
		if err := r.updateStatus(ctx, externalGateway, externalv1alpha1.Processing, "Waiting for sub-resources to become ready", conditions); err != nil {
			log.Error(err, "Failed to update Processing status while waiting for sub-resources")
		}
		return ctrl.Result{RequeueAfter: pendingRequeueInterval}, nil
	}

	conditions = append(conditions, externalv1alpha1.ReadyCondition(externalGateway.Generation))
	if err := r.updateStatus(ctx, externalGateway, externalv1alpha1.Ready, "All resources reconciled successfully", conditions); err != nil {
		if apierrors.IsConflict(err) {
			log.Info("Status update conflict on Ready, will retry on next reconciliation")
			return ctrl.Result{RequeueAfter: pendingRequeueInterval}, nil
		}
		log.Error(err, "Failed to update status to Ready")
		return ctrl.Result{}, err
	}

	log.Info("Reconciliation completed successfully")
	return ctrl.Result{RequeueAfter: defaultReconciliationInterval}, nil
}

// reconcileResources orchestrates all sub-resource reconciliations and returns per-component conditions.
// requeue is true when Gardener sub-resources are applied but not yet Ready (normal async behaviour).
func (r *ExternalGatewayReconciler) reconcileResources(ctx context.Context, log logr.Logger, external *externalv1alpha1.ExternalGateway) (conditions []metav1.Condition, requeue bool, err error) {
	log.Info("Reconciling ExternalGateway resources", "region", external.Spec.Region)

	internalDomain, err := r.buildInternalDomain(ctx, external.Spec.InternalDomain.KymaSubdomain)
	if err != nil {
		return nil, false, fmt.Errorf("failed to build internal domain: %w", err)
	}

	_, gardenerErr := dependencies.Gardener().AreAvailable(ctx, r.Client)
	isGardenerAvailable := gardenerErr == nil

	if isGardenerAvailable {
		dnsCond, dnsPending, dnsErr := r.reconcileDNSEntry(ctx, external, internalDomain)
		conditions = append(conditions, dnsCond)
		requeue = requeue || dnsPending

		certCond, certPending, certErr := r.reconcileCertificate(ctx, external, internalDomain)
		conditions = append(conditions, certCond)
		requeue = requeue || certPending

		if err := errors.Join(dnsErr, certErr); err != nil {
			return conditions, false, err
		}
	} else {
		log.Info("Gardener not available, skipping DNSEntry and Certificate reconciliation")
		conditions = append(conditions,
			metav1.Condition{
				Type:               externalv1alpha1.ConditionTypeDNSEntryReady,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: external.Generation,
				Reason:             externalv1alpha1.ReasonGardenerCRDUnavailable,
				Message:            "Gardener CRDs are not available; DNSEntry management skipped",
			},
			metav1.Condition{
				Type:               externalv1alpha1.ConditionTypeCertificateReady,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: external.Generation,
				Reason:             externalv1alpha1.ReasonGardenerCRDUnavailable,
				Message:            "Gardener CRDs are not available; Certificate management skipped",
			},
		)
	}

	if err := externalgateway.ReconcileCASecret(ctx, r.Client, external); err != nil {
		conditions = append(conditions, metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		})
		return conditions, false, fmt.Errorf("failed to reconcile CA Secret: %w", err)
	}

	if err := externalgateway.ReconcileGateway(ctx, r.Client, r.Scheme, external, internalDomain); err != nil {
		conditions = append(conditions, metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		})
		return conditions, false, fmt.Errorf("failed to reconcile Gateway: %w", err)
	}

	certSubjects, err := externalgateway.ResolveRegionCertSubjects(ctx, r.Client, external)
	if err != nil {
		conditions = append(conditions, metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		})
		return conditions, false, fmt.Errorf("failed to resolve certificate subjects: %w", err)
	}

	if err := externalgateway.ReconcileXFCCSanitizationFilter(ctx, r.Client, external); err != nil {
		conditions = append(conditions, metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		})
		return conditions, false, fmt.Errorf("failed to reconcile XFCC sanitization filter: %w", err)
	}

	if err := externalgateway.ReconcileCertValidationFilter(ctx, r.Client, external, certSubjects); err != nil {
		conditions = append(conditions, metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		})
		return conditions, false, fmt.Errorf("failed to reconcile certificate validation filter: %w", err)
	}

	conditions = append(conditions, metav1.Condition{
		Type:               externalv1alpha1.ConditionTypeGatewayConfigured,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: external.Generation,
		Reason:             externalv1alpha1.ReasonReady,
		Message:            "Istio Gateway and EnvoyFilters configured successfully",
	})

	return conditions, requeue, nil
}

// handleDeletion cleans up all resources when ExternalGateway is deleted
func (r *ExternalGatewayReconciler) handleDeletion(ctx context.Context, log logr.Logger, external *externalv1alpha1.ExternalGateway) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(external, externalGatewayFinalizer) {
		return ctrl.Result{}, nil
	}

	log.Info("Handling deletion, cleaning up resources")

	_, gardenerErr := dependencies.Gardener().AreAvailable(ctx, r.Client)
	isGardenerAvailable := gardenerErr == nil

	if err := externalgateway.DeleteGateway(ctx, r.Client, external.Namespace, external.GatewayName()); err != nil {
		log.Error(err, "Failed to delete Gateway")
		return ctrl.Result{}, err
	}

	if err := externalgateway.DeleteXFCCSanitizationFilter(ctx, r.Client, external.XFCCFilterName()); err != nil {
		log.Error(err, "Failed to delete XFCC sanitization EnvoyFilter")
		return ctrl.Result{}, err
	}

	if err := externalgateway.DeleteCertValidationFilter(ctx, r.Client, external.CertValidationFilterName()); err != nil {
		log.Error(err, "Failed to delete certificate validation EnvoyFilter")
		return ctrl.Result{}, err
	}

	if err := externalgateway.DeleteCASecret(ctx, r.Client, external.CASecretName()); err != nil {
		log.Error(err, "Failed to delete CA Secret")
		return ctrl.Result{}, err
	}

	if isGardenerAvailable {
		if err := externalgateway.DeleteDNSEntry(ctx, r.Client, external.DNSEntryName()); err != nil {
			log.Error(err, "Failed to delete DNSEntry")
			return ctrl.Result{}, err
		}

		if err := externalgateway.DeleteCertificate(ctx, r.Client, external.CertificateName(), external.TLSSecretName()); err != nil {
			log.Error(err, "Failed to delete Certificate")
			return ctrl.Result{}, err
		}
	}

	controllerutil.RemoveFinalizer(external, externalGatewayFinalizer)
	if err := r.Update(ctx, external); err != nil {
		log.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	log.Info("Successfully deleted all resources and removed finalizer")
	return ctrl.Result{}, nil
}

// updateStatus applies the given state, description, and conditions to the CR status using conflict-safe retry.
func (r *ExternalGatewayReconciler) updateStatus(ctx context.Context, external *externalv1alpha1.ExternalGateway,
	state externalv1alpha1.State, description string, conditions []metav1.Condition) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		fresh := &externalv1alpha1.ExternalGateway{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(external), fresh); err != nil {
			return err
		}
		fresh.Status.State = state
		fresh.Status.Description = description
		// Only advance ObservedGeneration once reconciliation has completed (success or terminal error).
		// During Processing, the spec has not yet been fully reconciled, so leaving ObservedGeneration
		// at its previous value correctly signals to consumers that the status is not yet up-to-date.
		if state != externalv1alpha1.Processing {
			fresh.Status.ObservedGeneration = fresh.Generation
		}
		fresh.Status.LastProcessedTime = metav1.Now()
		for _, c := range conditions {
			apimeta.SetStatusCondition(&fresh.Status.Conditions, c)
		}
		return r.Status().Update(ctx, fresh)
	})
}

// dnsEntryCondition maps a Gardener DNSEntry state string to a metav1.Condition.
func dnsEntryCondition(generation int64, state, message string) metav1.Condition {
	switch state {
	case dnsv1alpha1.STATE_READY:
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeDNSEntryReady,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonReady,
			Message:            "DNSEntry is ready",
		}
	case dnsv1alpha1.STATE_ERROR, dnsv1alpha1.STATE_INVALID, dnsv1alpha1.STATE_STALE:
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeDNSEntryReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonDNSEntryError,
			Message:            message,
		}
	default:
		// Pending, empty (just created), or any other transient state
		msg := "DNSEntry is being provisioned"
		if message != "" {
			msg = message
		}
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeDNSEntryReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonDNSEntryPending,
			Message:            msg,
		}
	}
}

// certificateCondition maps a Gardener Certificate state string to a metav1.Condition.
func certificateCondition(generation int64, state, message string) metav1.Condition {
	switch state {
	case certv1alpha1.StateReady:
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeCertificateReady,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonReady,
			Message:            "Certificate is ready",
		}
	case certv1alpha1.StateError, certv1alpha1.StateRevoked:
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeCertificateReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonCertificateError,
			Message:            message,
		}
	default:
		// Pending, empty (just created), or any other transient state
		msg := "Certificate is being issued"
		if message != "" {
			msg = message
		}
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeCertificateReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			Reason:             externalv1alpha1.ReasonCertificatePending,
			Message:            msg,
		}
	}
}

// reconcileDNSEntry reconciles the DNSEntry sub-resource and returns its condition, whether it is
// still pending, and any error that should block the overall reconciliation result.
// A hard k8s API error is returned directly; a Gardener terminal state is returned as an error
// with the condition already set so the caller always has both pieces of information.
func (r *ExternalGatewayReconciler) reconcileDNSEntry(ctx context.Context, external *externalv1alpha1.ExternalGateway, internalDomain string) (metav1.Condition, bool, error) {
	if err := externalgateway.ReconcileDNSEntry(ctx, r.Client, external, internalDomain); err != nil {
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeDNSEntryReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		}, false, fmt.Errorf("failed to reconcile DNSEntry: %w", err)
	}

	state, msg, err := externalgateway.GetDNSEntryStatus(ctx, r.Client, external.DNSEntryName())
	if err != nil {
		return metav1.Condition{}, false, fmt.Errorf("failed to get DNSEntry status: %w", err)
	}

	cond := dnsEntryCondition(external.Generation, state, msg)
	switch state {
	case dnsv1alpha1.STATE_READY:
		return cond, false, nil
	case dnsv1alpha1.STATE_ERROR, dnsv1alpha1.STATE_INVALID, dnsv1alpha1.STATE_STALE:
		return cond, false, fmt.Errorf("DNSEntry is in a terminal error state %q: %s", state, msg)
	default:
		return cond, true, nil
	}
}

// reconcileCertificate reconciles the Certificate sub-resource and returns its condition, whether it
// is still pending, and any error that should block the overall reconciliation result.
func (r *ExternalGatewayReconciler) reconcileCertificate(ctx context.Context, external *externalv1alpha1.ExternalGateway, internalDomain string) (metav1.Condition, bool, error) {
	if err := externalgateway.ReconcileCertificate(ctx, r.Client, external, internalDomain); err != nil {
		return metav1.Condition{
			Type:               externalv1alpha1.ConditionTypeCertificateReady,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: external.Generation,
			Reason:             externalv1alpha1.ReasonFailed,
			Message:            err.Error(),
		}, false, fmt.Errorf("failed to reconcile Certificate: %w", err)
	}

	state, msg, err := externalgateway.GetCertificateStatus(ctx, r.Client, external.CertificateName())
	if err != nil {
		return metav1.Condition{}, false, fmt.Errorf("failed to get Certificate status: %w", err)
	}

	cond := certificateCondition(external.Generation, state, msg)
	switch state {
	case certv1alpha1.StateReady:
		return cond, false, nil
	case certv1alpha1.StateError, certv1alpha1.StateRevoked:
		return cond, false, fmt.Errorf("certificate is in a terminal error state %q: %s", state, msg)
	default:
		return cond, true, nil
	}
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
