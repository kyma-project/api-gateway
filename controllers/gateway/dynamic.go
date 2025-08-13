package gateway

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/certificate"
	apiGatewayMetrics "github.com/kyma-project/api-gateway/internal/metrics"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sync"
	"sync/atomic"
)

type APIRuleReconcilerStarter struct {
	isStarted            *atomic.Bool
	metricsRegistered    *atomic.Bool
	managerContextCancel context.CancelFunc
	options              StarterOptions
	creationLock         *sync.Mutex
	metrics              *apiGatewayMetrics.ApiGatewayMetrics
}

type StarterOptions struct {
	scheme         *runtime.Scheme
	rateLimiterCfg controllers.RateLimiterConfig
	setupLog       logr.Logger
	reconcileCfg   ApiRuleReconcilerConfiguration
	flagVar        *flag.FlagSet
}

func NewAPIRuleReconcilerStarter(
	scheme *runtime.Scheme,
	rateLimiterCfg controllers.RateLimiterConfig,
	setupLog logr.Logger,
	reconcileCfg ApiRuleReconcilerConfiguration,
) *APIRuleReconcilerStarter {
	return &APIRuleReconcilerStarter{
		isStarted:    &atomic.Bool{},
		metricsRegistered: &atomic.Bool{},
		creationLock: &sync.Mutex{},
		options: StarterOptions{
			scheme:         scheme,
			rateLimiterCfg: rateLimiterCfg,
			setupLog:       setupLog,
			reconcileCfg:   reconcileCfg,
		},
	}
}

func (r *APIRuleReconcilerStarter) SetupAndStartManager() error {
	r.creationLock.Lock()
	defer r.creationLock.Unlock()

	if r.isStarted != nil && r.isStarted.Load() {
		r.options.setupLog.Info("APIRule reconciler is already started, skipping setup")
		return nil
	}

	options := ctrl.Options{
		Scheme: r.options.scheme,
		Metrics: metricsserver.Options{
			BindAddress: ":8082",
		},
		LeaderElection:   true,
		LeaderElectionID: "apirules.kyma-project.io",
		WebhookServer: webhook.NewServer(webhook.Options{
			TLSOpts: []func(*tls.Config){
				func(cfg *tls.Config) {
					cfg.GetCertificate = certificate.GetCertificate
				},
			},
		}),
		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			opts.ByObject = map[client.Object]cache.ByObject{
				&corev1.Secret{}: {
					Namespaces: map[string]cache.Config{
						"kyma-system": {},
					},
				},
			}
			return cache.New(config, opts)
		},
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&rulev1alpha1.Rule{},
					/*
						Reading v1beta1 and v2alpha1 APIRules during reconciliation led to an issue that the APIRule could not be read in v2alpha1 after it was deleted.
						This would self-heal in the next reconciliation loop.To avoid this confusion with this issue, we disable the cache for v2alpha1 APIRules.
						This can probably be enabled again when reconciliation only uses v2alpha1.
					*/
					&gatewayv1beta1.APIRule{},
					&gatewayv2alpha1.APIRule{},
					&gatewayv2.APIRule{},
					&corev1.Secret{},
				},
			},
		},
	}

	restCfg, err := ctrl.GetConfig()
	if err != nil {
		r.options.setupLog.Error(err, "Unable to get REST config for APIRule reconciler")
		return err
	}

	mgr, err := ctrl.NewManager(restCfg, options)
	if err != nil {
		r.options.setupLog.Error(err, "Unable to create manager for APIRule reconciler")
		return err
	}

	if r.metricsRegistered != nil && !r.metricsRegistered.Load() {
		r.metrics = apiGatewayMetrics.NewApiGatewayMetrics()
		r.metricsRegistered.Store(true)
	}

	if err := NewApiRuleReconciler(mgr, r.options.reconcileCfg, r.metrics).
		SetupWithManager(mgr, r.options.rateLimiterCfg); err != nil {
		r.options.setupLog.Error(err, "Unable to create controller", "controller", "APIGateway")
		return err
	}

	if err = certificate.NewCertificateReconciler(mgr).SetupWithManager(mgr, r.options.rateLimiterCfg); err != nil {
		r.options.setupLog.Error(err, "Unable to create controller", "controller", "certificate")
		os.Exit(1)
	}

	if err := (&gatewayv2alpha1.APIRule{}).SetupWebhookWithManager(mgr); err != nil {
		r.options.setupLog.Error(err, "Unable to create webhook", "webhook", "APIRule")
		os.Exit(1)
	}

	k8sClient, err := client.New(restCfg, client.Options{Scheme: r.options.scheme})
	if err != nil {
		r.options.setupLog.Error(err, "Unable to create client")
		os.Exit(1)
	}

	if err = certificate.ReadCertificateSecret(context.Background(), k8sClient, r.options.setupLog); err != nil {
		r.options.setupLog.Error(err, "Unable to read certificate secret", "webhook", "certificate")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.managerContextCancel = cancel

	r.isStarted.Store(true)
	go func() {
		err := mgr.Start(ctx)
		if err != nil {
			r.options.setupLog.Error(err, "Problem running APIRule reconciler manager")
		}
	}()

	return nil
}

func (r *APIRuleReconcilerStarter) StopManager() error {
	if r.isStarted == nil || !r.isStarted.Load() {
		r.options.setupLog.Info("APIRule reconciler is not started, nothing to stop")
		return nil
	}

	r.options.setupLog.Info("Stopping APIRule reconciler manager")
	r.managerContextCancel()

	r.isStarted.Store(false)
	r.options.setupLog.Info("Manager stopped successfully")
	return nil
}
