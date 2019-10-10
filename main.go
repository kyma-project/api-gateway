/*

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
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"os"
	"strings"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = gatewayv1alpha1.AddToScheme(scheme)
	_ = networkingv1alpha3.AddToScheme(scheme)
	_ = rulev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var jwksURI string
	var oathkeeperSvcAddr string
	var oathkeeperSvcPort uint
	var blackListedServices string
	var whiteListedDomains string
	var corsAllowOrigin, corsAllowMethods, corsAllowHeaders string

	flag.StringVar(&oathkeeperSvcAddr, "oathkeeper-svc-address", "", "Oathkeeper proxy service")
	flag.UintVar(&oathkeeperSvcPort, "oathkeeper-svc-port", 0, "Oathkeeper proxy service port")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&jwksURI, "jwks-uri", "", "URL of the provider's public key set to validate signature of the JWT")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&blackListedServices, "service-blacklist", "kubernetes.default,kube-dns.kube-system", "List of services to be blacklisted from exposure.")
	flag.StringVar(&whiteListedDomains, "domain-whitelist", "", "List of domains to be allowed.")
	flag.StringVar(&corsAllowOrigin, "cors-allow-origin", "*", "list of allowed origins")
	flag.StringVar(&corsAllowMethods, "cors-allow-methods", "GET,POST,PUT,DELETE", "list of allowed methods")
	flag.StringVar(&corsAllowHeaders, "cors-allow-headers", "Authorization,Content-Type,*", "list of allowed headers")

	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

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
	if whiteListedDomains == "" {
		setupLog.Error(fmt.Errorf("domain-whitelist can't be empty"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	} else {
		for _, domain := range getList(whiteListedDomains) {
			if !validation.ValidateDomainName(domain) {
				setupLog.Error(fmt.Errorf("invalid domain in domain-whitelist"), "unable to create controller", "controller", "Api")
				os.Exit(1)
			}
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.APIReconciler{
		Client:            mgr.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     oathkeeperSvcAddr,
		OathkeeperSvcPort: uint32(oathkeeperSvcPort),
		JWKSURI:           jwksURI,
		Validator: &validation.APIRule{
			ServiceBlackList: getNamespaceServiceMap(blackListedServices),
			DomainWhiteList:  getList(whiteListedDomains),
		},
		CorsConfig: &processing.CorsConfig{
			AllowHeaders: getList(corsAllowHeaders),
			AllowMethods: getList(corsAllowMethods),
			AllowOrigin:  getList(corsAllowOrigin),
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

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
func getNamespaceServiceMap(raw string) map[string][]string {
	result := make(map[string][]string)
	for _, s := range getList(raw) {
		if !validation.ValidateServiceName(s) {
			setupLog.Error(fmt.Errorf("invalid service in service-blacklist"), "unable to create controller", "controller", "Api")
			os.Exit(1)
		}
		namespacedService := strings.Split(s, ".")
		namespace := namespacedService[1]
		service := namespacedService[0]
		result[namespace] = append(result[namespace], service)
	}
	return result
}
