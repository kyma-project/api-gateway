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
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"istio.io/api/networking/v1beta1"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/vrischmann/envconfig"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"github.com/pkg/errors"
	//+kubebuilder:scaffold:imports
)

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=api-gateway-webhook-service"`
	WebhookSecretName  string `envconfig:"default=api-gateway-webhook-service"`
	WebhookPort        int    `envconfig:"default=9443"`
}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gatewayv1alpha1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1beta1.AddToScheme(scheme))

	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(rulev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var healthProbeAddr string
	var enableLeaderElection bool
	var jwksURI string
	var oathkeeperSvcAddr string
	var oathkeeperSvcPort uint
	var blockListedServices string
	var allowListedDomains string
	var domainName string
	var corsAllowOrigins, corsAllowMethods, corsAllowHeaders string
	var generatedObjectsLabels string

	const blockListedSubdomains string = "api"

	flag.StringVar(&oathkeeperSvcAddr, "oathkeeper-svc-address", "", "Oathkeeper proxy service")
	flag.UintVar(&oathkeeperSvcPort, "oathkeeper-svc-port", 0, "Oathkeeper proxy service port")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&healthProbeAddr, "health-probe-addr", ":8081", "The address the health probe endpoint binds to.")
	flag.StringVar(&jwksURI, "jwks-uri", "", "URL of the provider's public key set to validate signature of the JWT")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&blockListedServices, "service-blocklist", "kubernetes.default,kube-dns.kube-system", "List of services to be blocklisted from exposure.")
	flag.StringVar(&allowListedDomains, "domain-allowlist", "", "List of domains to be allowed.")
	flag.StringVar(&domainName, "default-domain-name", "", "A default domain name for hostnames with no domain provided. Optional.")
	flag.StringVar(&corsAllowOrigins, "cors-allow-origins", "regex:.*", "list of allowed origins")
	flag.StringVar(&corsAllowMethods, "cors-allow-methods", "GET,POST,PUT,DELETE", "list of allowed methods")
	flag.StringVar(&corsAllowHeaders, "cors-allow-headers", "Authorization,Content-Type,*", "list of allowed headers")
	flag.StringVar(&generatedObjectsLabels, "generated-objects-labels", "", "Comma-separated list of key=value pairs used to label generated objects")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	if jwksURI == "" {
		setupLog.Error(fmt.Errorf("jwks-uri required, but not supplied"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	if oathkeeperSvcAddr == "" {
		setupLog.Error(fmt.Errorf("oathkeeper-svc-address can't be empty"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	if oathkeeperSvcPort == 0 {
		setupLog.Error(fmt.Errorf("oathkeeper-svc-port can't be empty"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	if allowListedDomains != "" {
		for _, domain := range getList(allowListedDomains) {
			if !validation.ValidateDomainName(domain) {
				setupLog.Error(fmt.Errorf("invalid domain in domain-allowlist"), "unable to create controller", "controller", "Api")
				os.Exit(1)
			}
		}
	}

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := &config{}
	setupLog.Info("reading webhook configuration")
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: healthProbeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "69358922.kyma-project.io",
		CertDir:                "/tmp/k8s-webhook-server/serving-certs",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	additionalLabels, err := parseLabels(generatedObjectsLabels)
	if err != nil {
		setupLog.Error(err, "parsing labels failed")
		os.Exit(1)
	}

	if err = (&controllers.APIRuleReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     oathkeeperSvcAddr,
		OathkeeperSvcPort: uint32(oathkeeperSvcPort),
		JWKSURI:           jwksURI,
		ServiceBlockList:  getNamespaceServiceMap(blockListedServices),
		DomainAllowList:   getList(allowListedDomains),
		HostBlockList:     getHostBlockListFrom(blockListedSubdomains, domainName),
		DefaultDomainName: domainName,
		CorsConfig: &processing.CorsConfig{
			AllowHeaders: getList(corsAllowHeaders),
			AllowMethods: getList(corsAllowMethods),
			AllowOrigins: getStringMatch(corsAllowOrigins),
		},
		GeneratedObjectsLabels: additionalLabels,
		Scheme:                 mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIRule")
		os.Exit(1)
	}
	if err = (&gatewayv1beta1.APIRule{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "APIRule")
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

func getList(raw string) []string {
	var result []string
	for _, s := range strings.Split(raw, ",") {
		trim := strings.TrimSpace(s)
		if trim != "" {
			result = append(result, trim)
		}
	}
	return result
}

func getStringMatch(raw string) []*v1beta1.StringMatch {
	var result []*v1beta1.StringMatch
	for _, s := range getList(raw) {
		matchTypePair := strings.SplitN(s, ":", 2)
		matchType := matchTypePair[0]
		value := matchTypePair[1]
		var stringMatch *v1beta1.StringMatch
		switch {
		case matchType == "regex":
			stringMatch = regex(value)
		case matchType == "prefix":
			stringMatch = prefix(value)
		case matchType == "exact":
			stringMatch = exact(value)
		}
		result = append(result, stringMatch)
	}
	return result
}

func getNamespaceServiceMap(raw string) map[string][]string {
	result := make(map[string][]string)
	for _, s := range getList(raw) {
		if !validation.ValidateServiceName(s) {
			setupLog.Error(fmt.Errorf("invalid service in service-blocklist"), "unable to create controller", "controller", "Api")
			os.Exit(1)
		}
		namespacedService := strings.Split(s, ".")
		namespace := namespacedService[1]
		service := namespacedService[0]
		result[namespace] = append(result[namespace], service)
	}
	return result
}

func getHostBlockListFrom(blockListedSubdomains string, domainName string) []string {
	var result []string
	for _, subdomain := range getList(blockListedSubdomains) {
		if !validation.ValidateSubdomainName(subdomain) {
			setupLog.Error(fmt.Errorf("invalid subdomain in subdomain-blocklist"), "unable to create controller", "controller", "Api")
			os.Exit(1)
		}
		blockedHost := strings.Join([]string{subdomain, domainName}, ".")
		result = append(result, blockedHost)
	}
	return result
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

func regex(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Regex{Regex: val},
	}
}

func prefix(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Prefix{Prefix: val},
	}
}

func exact(val string) *v1beta1.StringMatch {
	return &v1beta1.StringMatch{
		MatchType: &v1beta1.StringMatch_Exact{Exact: val},
	}
}
