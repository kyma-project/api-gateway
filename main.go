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
	"flag"
	"fmt"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	"os"
	"strings"
	"time"

	gatewayv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/pkg/errors"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	operatorcontrollers "github.com/kyma-project/api-gateway/controllers/operator"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

type FlagVar struct {
	metricsAddr                 string
	enableLeaderElection        bool
	probeAddr                   string
	rateLimiterFailureBaseDelay time.Duration
	rateLimiterFailureMaxDelay  time.Duration
	rateLimiterFrequency        int
	rateLimiterBurst            int
	// TODO: Remove not-relevant startup flags, e.g. cors, domainName
	blockListedServices                                  string
	allowListedDomains                                   string
	domainName                                           string
	corsAllowOrigins, corsAllowMethods, corsAllowHeaders string
	generatedObjectsLabels                               string
	reconciliationInterval                               time.Duration
	errorReconciliationPeriod                            uint
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gatewayv1alpha1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))

	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(rulev1alpha1.AddToScheme(scheme))
	utilruntime.Must(securityv1beta1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func defineFlagVar() *FlagVar {
	flagVar := new(FlagVar)
	flag.StringVar(&flagVar.metricsAddr, "metrics-bind-address", ":8090", "The address the metric endpoint binds to.")
	flag.StringVar(&flagVar.probeAddr, "health-probe-bind-address", ":8091", "The address the probe endpoint binds to.")
	flag.BoolVar(&flagVar.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&flagVar.rateLimiterBurst, "rate-limiter-burst", 200,
		"Indicates the burst value for the bucket rate limiter.")
	flag.IntVar(&flagVar.rateLimiterFrequency, "rate-limiter-frequency", 30,
		"Indicates the bucket rate limiter frequency, signifying no. of events per second.")
	flag.DurationVar(&flagVar.rateLimiterFailureBaseDelay, "failure-base-delay", 1*time.Second,
		"Indicates the failure base delay for rate limiter.")
	flag.DurationVar(&flagVar.rateLimiterFailureMaxDelay, "failure-max-delay", 1000*time.Second,
		"Indicates the failure max delay.")
	flag.StringVar(&flagVar.blockListedServices, "service-blocklist", "kubernetes.default,kube-dns.kube-system", "List of services to be blocklisted from exposure.")
	flag.StringVar(&flagVar.allowListedDomains, "domain-allowlist", "", "List of domains to be allowed.")
	flag.StringVar(&flagVar.domainName, "default-domain-name", "", "A default domain name for hostnames with no domain provided. Optional.")
	flag.StringVar(&flagVar.corsAllowOrigins, "cors-allow-origins", "regex:.*", "list of allowed origins")
	flag.StringVar(&flagVar.corsAllowMethods, "cors-allow-methods", "GET,POST,PUT,DELETE", "list of allowed methods")
	flag.StringVar(&flagVar.corsAllowHeaders, "cors-allow-headers", "Authorization,Content-Type,*", "list of allowed headers")
	flag.StringVar(&flagVar.generatedObjectsLabels, "generated-objects-labels", "", "Comma-separated list of key=value pairs used to label generated objects")
	flag.DurationVar(&flagVar.reconciliationInterval, "reconciliation-interval", 1*time.Hour, "Indicates the time based reconciliation interval.")
	// TODO we don't have an error reconciliation period in istio operator and therefore might want to remove it here too to have the same handling.
	flag.UintVar(&flagVar.errorReconciliationPeriod, "error-reconciliation-period", 60, "Reconciliation period after an error happened in the previous run [s]")

	return flagVar
}

func main() {
	flagVar := defineFlagVar()
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// TODO: Add RateLimiter

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: flagVar.metricsAddr,
		},
		HealthProbeBindAddress: flagVar.probeAddr,
		LeaderElection:         flagVar.enableLeaderElection,
		LeaderElectionID:       "69358922.kyma-project.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	additionalLabels, err := parseLabels(flagVar.generatedObjectsLabels)
	if err != nil {
		setupLog.Error(err, "parsing labels failed")
		os.Exit(1)
	}

	config := gateway.ApiRuleReconcilerConfiguration{
		AllowListedDomains:        flagVar.allowListedDomains,
		BlockListedServices:       flagVar.blockListedServices,
		DomainName:                flagVar.domainName,
		CorsAllowOrigins:          flagVar.corsAllowOrigins,
		CorsAllowMethods:          flagVar.corsAllowMethods,
		CorsAllowHeaders:          flagVar.corsAllowHeaders,
		AdditionalLabels:          additionalLabels,
		ReconciliationPeriod:      uint(flagVar.reconciliationInterval.Seconds()),
		ErrorReconciliationPeriod: flagVar.errorReconciliationPeriod,
	}

	apiRuleReconciler, err := gateway.NewApiRuleReconciler(mgr, config)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIRule")
		os.Exit(1)
	}
	if err = apiRuleReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup controller", "controller", "APIRule")
		os.Exit(1)
	}
	apiGatewayReconciler := operatorcontrollers.NewAPIGatewayReconciler(mgr)
	if err = (apiGatewayReconciler).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIGateway")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func parseLabels(labelsString string) (map[string]string, error) {

	output := make(map[string]string)

	if labelsString == "" {
		return output, nil
	}

	var err error

	for _, labelString := range strings.Split(labelsString, ",") {
		trim := strings.TrimSpace(labelString)
		if trim != "" {
			label := strings.Split(trim, "=")
			if len(label) != 2 {
				return nil, errors.New("invalid label format")
			}

			key, value := label[0], label[1]

			if err = validation.VerifyLabelKey(key); err != nil {
				return nil, errors.Wrap(err, "invalid label key")
			}

			if err = validation.VerifyLabelValue(value); err != nil {
				return nil, errors.Wrap(err, "invalid label value")
			}

			_, exists := output[key]
			if exists {
				return nil, fmt.Errorf("duplicated label: %s", key)
			}

			output[key] = value
		}
	}

	return output, nil
}
