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

package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"istio.io/api/networking/v1alpha3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/builders/envoyfilter"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/ratelimit"
)

const (
	defaultReconciliationPeriod = 3 * time.Minute
)

// RateLimitReconciler reconciles a RateLimit object.
type RateLimitReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	ReconcilePeriod time.Duration
}

//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=ratelimits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=ratelimits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.istio.io,resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main Kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// In this function, the RateLimit object is fetched and validated.
// If the object is not found, it is ignored. If validation fails, an error is returned.
// Otherwise, the function returns a result with a requeue period.
func (r *RateLimitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Starting reconciliation")

	rl := ratelimitv1alpha1.RateLimit{}
	if err := r.Get(ctx, req.NamespacedName, &rl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if d, err := dependencies.RateLimit().AreAvailable(ctx, r.Client); err != nil {
		rl.Status.Error(fmt.Errorf("dependency missing '%s': %w", d, err))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	existingAPIGateways := &operatorv1alpha1.APIGatewayList{}
	if err := r.List(ctx, existingAPIGateways); err != nil {
		l.Info("Unable to list APIGateway CRs")
		return ctrl.Result{}, err
	}

	if len(existingAPIGateways.Items) < 1 {
		rl.Status.Warning(errors.New("failed to reconcile RateLimit CR because of missing APIGateway CR in the cluster"))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		err := errors.New("no APIGateway CR in the cluster")
		return ctrl.Result{}, err
	}

	latestCr := operatorv1alpha1.GetOldestAPIGatewayCR(existingAPIGateways)
	if latestCr == nil {
		rl.Status.Warning(errors.New("failed to reconcile RateLimit CR because of missing APIGateway CR in the cluster"))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		err := errors.New("no APIGateway CR in the cluster")
		return ctrl.Result{}, err
	}

	if latestCr.Status.State != operatorv1alpha1.Ready {
		rl.Status.Warning(fmt.Errorf("failed to create RateLimit CR because APIGateway CR is in %s state", latestCr.Status.State))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		err := fmt.Errorf("APIGateway CR %s/%s is in %s state", latestCr.Namespace, latestCr.Name, latestCr.Status.State)
		return ctrl.Result{}, err
	}

	l.Info("Validating RateLimit resource")
	err := ratelimit.Validate(ctx, r.Client, rl)
	if err != nil {
		rl.Status.Error(fmt.Errorf("failed to validate RateLimit: %w", err))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	builder := envoyfilter.NewEnvoyFilterBuilder().
		WithName(rl.Name).
		WithNamespace(rl.Namespace)
	ef := builder.Build()

	l.Info("Updating EnvoyFilter resource to desired state", "EnvoyFilter.Name", ef.Name)
	if err := r.createOrUpdate(ctx, ef, func() error {
		if err := controllerutil.SetControllerReference(&rl, ef, r.Scheme); err != nil {
			return err
		}
		// build desired configuration
		ef.Spec.WorkloadSelector = &v1alpha3.WorkloadSelector{Labels: rl.Spec.SelectorLabels}
		defaultBucket := ratelimit.Bucket{
			MaxTokens:     rl.Spec.Local.DefaultBucket.MaxTokens,
			TokensPerFill: rl.Spec.Local.DefaultBucket.TokensPerFill,
			FillInterval:  rl.Spec.Local.DefaultBucket.FillInterval.Duration,
		}
		limit := ratelimit.NewLocalRateLimit().
			WithDefaultBucket(defaultBucket).
			Enforce(rl.Spec.Enforce).
			EnableResponseHeaders(rl.Spec.EnableResponseHeaders)
		for _, b := range rl.Spec.Local.Buckets {
			d := ratelimit.Descriptor{}
			if len(b.Path) > 0 {
				d.Entries = append(d.Entries, ratelimit.DescriptorEntry{Key: "path", Val: b.Path})
			}
			for k, v := range b.Headers {
				d.Entries = append(d.Entries, ratelimit.DescriptorEntry{Key: k, Val: v})
			}
			d.Bucket = ratelimit.Bucket{
				MaxTokens:     b.Bucket.MaxTokens,
				TokensPerFill: b.Bucket.TokensPerFill,
				FillInterval:  b.Bucket.FillInterval.Duration,
			}
			limit.For(d)
		}
		limit.SetConfigPatches(ef)
		return nil
	}); err != nil {
		l.Error(err, "Failed to create EnvoyFilter", "EnvoyFilter.Name", ef.Name)
		rl.Status.Error(fmt.Errorf("failed to create EnvoyFilter: %w", err))
		if err := r.Status().Update(ctx, &rl); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	l.Info("Reconciliation finished")
	rl.Status.Ready()
	if err := r.Status().Update(ctx, &rl); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: r.ReconcilePeriod}, nil
}

func (r *RateLimitReconciler) createOrUpdate(ctx context.Context, obj client.Object, mutate func() error) error {
	key := client.ObjectKeyFromObject(obj)
	if err := r.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err := mutate(); err != nil {
			return err
		}
		if err := r.Create(ctx, obj); err != nil {
			return err
		}
		return nil
	}
	if err := mutate(); err != nil {
		return err
	}
	if err := r.Update(ctx, obj); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.RateLimit{}).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

func NewRateLimitReconciler(mgr manager.Manager) *RateLimitReconciler {
	return &RateLimitReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		ReconcilePeriod: defaultReconciliationPeriod,
	}
}
