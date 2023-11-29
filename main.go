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
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	"os"
	"strings"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/controllers/gateway"
	operatorcontrollers "github.com/kyma-project/api-gateway/controllers/operator"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

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

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/pkg/errors"
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
	generatedObjectsLabels      string
	reconciliationInterval      time.Duration
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))
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
	flag.StringVar(&flagVar.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&flagVar.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&flagVar.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&flagVar.rateLimiterBurst, "rate-limiter-burst", controllers.RateLimiterBurst,
		"Indicates the burst value for the bucket rate limiter.")
	flag.IntVar(&flagVar.rateLimiterFrequency, "rate-limiter-frequency", controllers.RateLimiterFrequency,
		"Indicates the bucket rate limiter frequency, signifying no. of events per second.")
	flag.DurationVar(&flagVar.rateLimiterFailureBaseDelay, "failure-base-delay", controllers.RateLimiterFailureBaseDelay,
		"Indicates the failure base delay for rate limiter.")
	flag.DurationVar(&flagVar.rateLimiterFailureMaxDelay, "failure-max-delay", controllers.RateLimiterFailureMaxDelay,
		"Indicates the failure max delay for rate limiter. .")
	flag.StringVar(&flagVar.generatedObjectsLabels, "generated-objects-labels", "", "Comma-separated list of key=value pairs used to label generated objects")
	flag.DurationVar(&flagVar.reconciliationInterval, "reconciliation-interval", 1*time.Hour, "Indicates the time based reconciliation interval of APIRule.")

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
		OathkeeperSvcAddr:         "ory-oathkeeper-proxy.kyma-system.svc.cluster.local",
		OathkeeperSvcPort:         4455,
		CorsAllowOrigins:          "regex:.*",
		CorsAllowMethods:          "GET,POST,PUT,DELETE,PATCH",
		CorsAllowHeaders:          "Authorization,Content-Type,*",
		AdditionalLabels:          additionalLabels,
		ReconciliationPeriod:      uint(flagVar.reconciliationInterval.Seconds()),
		ErrorReconciliationPeriod: 60,
	}

	rateLimiterCfg := controllers.RateLimiterConfig{
		Burst:            flagVar.rateLimiterBurst,
		Frequency:        flagVar.rateLimiterFrequency,
		FailureBaseDelay: flagVar.rateLimiterFailureBaseDelay,
		FailureMaxDelay:  flagVar.rateLimiterFailureMaxDelay,
	}

	apiRuleReconciler, err := gateway.NewApiRuleReconciler(mgr, config)
	if err != nil {
		setupLog.Error(err, "unable to create APIRule reconciler", "controller", "APIRule")
		os.Exit(1)
	}
	if err = apiRuleReconciler.SetupWithManager(mgr, rateLimiterCfg); err != nil {
		setupLog.Error(err, "unable to setup controller", "controller", "APIRule")
		os.Exit(1)
	}

	if err = operatorcontrollers.NewAPIGatewayReconciler(mgr, oathkeeper.NewReconciler()).SetupWithManager(mgr, rateLimiterCfg); err != nil {
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
