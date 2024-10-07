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

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"time"

	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	"github.com/kyma-project/api-gateway/internal/version"

	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/certificate"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	"github.com/kyma-project/api-gateway/controllers/operator"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

type FlagVar struct {
	initOnly                    bool
	metricsAddr                 string
	enableLeaderElection        bool
	probeAddr                   string
	rateLimiterFailureBaseDelay time.Duration
	rateLimiterFailureMaxDelay  time.Duration
	rateLimiterFrequency        int
	rateLimiterBurst            int
	reconciliationInterval      time.Duration
	migrationInterval           time.Duration
	logJson                     bool
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
	utilruntime.Must(gatewayv2alpha1.AddToScheme(scheme))
	utilruntime.Must(dnsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(certv1alpha1.AddToScheme(scheme))

	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(rulev1alpha1.AddToScheme(scheme))
	utilruntime.Must(securityv1beta1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func defineFlagVar() *FlagVar {
	flagVar := new(FlagVar)
	flag.BoolVar(&flagVar.initOnly, "init-only", false,
		"Should only initialise operator prerequisites.")
	flag.StringVar(&flagVar.metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&flagVar.probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&flagVar.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&flagVar.rateLimiterBurst, "rate-limiter-burst", controllers.RateLimiterBurst,
		"Indicates the burst value for the bucket rate limiter.")
	flag.IntVar(&flagVar.rateLimiterFrequency, "rate-limiter-frequency", controllers.RateLimiterFrequency,
		"Indicates the bucket rate limiter frequency, signifying no. of events per second.")
	flag.DurationVar(&flagVar.rateLimiterFailureBaseDelay, "failure-base-delay", controllers.RateLimiterFailureBaseDelay,
		"Indicates the failure base delay for rate limiter.")
	flag.DurationVar(&flagVar.rateLimiterFailureMaxDelay, "failure-max-delay", controllers.RateLimiterFailureMaxDelay,
		"Indicates the failure max delay for rate limiter. .")
	flag.DurationVar(&flagVar.reconciliationInterval, "reconciliation-interval", 30*time.Minute,
		"Indicates the time based reconciliation interval of APIRule.")
	flag.DurationVar(&flagVar.migrationInterval, "migration-interval", 1*time.Minute,
		"Indicates the time taken between steps of APIRule version migration.")
	flag.BoolVar(&flagVar.logJson, "log-json", true,
		"Project program logs as JSON.")

	return flagVar
}

func main() {
	flagVar := defineFlagVar()
	flag.Parse()
	opts := zap.Options{
		Development: flagVar.logJson,
	}
	opts.BindFlags(flag.CommandLine)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	config := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Unable to create client")
		os.Exit(1)
	}

	if flagVar.initOnly {
		setupLog.Info("Initialisation only mode")
		utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
		err = certificate.InitialiseCertificateSecret(context.Background(), k8sClient, setupLog)
		if err != nil {
			setupLog.Error(err, "Unable to initialise certificate secret")
			os.Exit(1)
		}
		os.Exit(0)
	}

	options := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: flagVar.metricsAddr,
		},
		HealthProbeBindAddress: flagVar.probeAddr,
		LeaderElection:         flagVar.enableLeaderElection,
		LeaderElectionID:       "69358922.kyma-project.io",
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
					&gatewayv2alpha1.APIRule{},
					&corev1.Secret{},
				},
			},
		},
	}

	mgr, err := ctrl.NewManager(config, options)
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	reconcileConfig := gateway.ApiRuleReconcilerConfiguration{
		OathkeeperSvcAddr:             "ory-oathkeeper-proxy.kyma-system.svc.cluster.local",
		OathkeeperSvcPort:             4455,
		CorsAllowOrigins:              "regex:.*",
		CorsAllowMethods:              "GET,POST,PUT,DELETE,PATCH",
		CorsAllowHeaders:              "Authorization,Content-Type,*",
		ReconciliationPeriod:          uint(flagVar.reconciliationInterval.Seconds()),
		ErrorReconciliationPeriod:     60,
		MigrationReconciliationPeriod: uint(flagVar.migrationInterval.Seconds()),
	}

	rateLimiterCfg := controllers.RateLimiterConfig{
		Burst:            flagVar.rateLimiterBurst,
		Frequency:        flagVar.rateLimiterFrequency,
		FailureBaseDelay: flagVar.rateLimiterFailureBaseDelay,
		FailureMaxDelay:  flagVar.rateLimiterFailureMaxDelay,
	}

	if err := gateway.NewApiRuleReconciler(mgr, reconcileConfig).SetupWithManager(mgr, rateLimiterCfg); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "APIRule")
		os.Exit(1)
	}

	if err = operator.NewAPIGatewayReconciler(mgr, oathkeeper.NewReconciler()).SetupWithManager(mgr, rateLimiterCfg); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "APIGateway")
		os.Exit(1)
	}

	if err = certificate.NewCertificateReconciler(mgr).SetupWithManager(mgr, rateLimiterCfg); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "certificate")
		os.Exit(1)
	}

	if err = certificate.ReadCertificateSecret(context.Background(), k8sClient, setupLog); err != nil {
		setupLog.Error(err, "Unable to read certificate secret", "webhook", "certificate")
		os.Exit(1)
	}

	if err = (&gatewayv1beta1.APIRule{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create webhook", "webhook", "APIRule")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	setupLog.Info("Module version", "version", version.GetModuleVersion())

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}
