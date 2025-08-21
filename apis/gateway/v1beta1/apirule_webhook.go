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

package v1beta1

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var apirulelog = logf.Log.WithName("apirule-resource")

var (
	v1beta1CreationCounter prometheus.Counter
	v1beta1UpdateCounter   prometheus.Counter
)

func (ruleV1 *APIRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	v1beta1CreationCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "apirule_v1beta1_creation_actions_total",
		Namespace: "api_gateway",
		Help:      "The total number of APIRule v1beta1 CREATE actions received",
	})
	v1beta1UpdateCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "apirule_v1beta1_update_actions_total",
		Namespace: "api_gateway",
		Help:      "The total number of APIRule v1beta1 UPDATE actions received",
	})

	ctrlmetrics.Registry.MustRegister(v1beta1CreationCounter)
	ctrlmetrics.Registry.MustRegister(v1beta1UpdateCounter)
	return ctrl.NewWebhookManagedBy(mgr).
		For(ruleV1).
		WithValidator(&ValidatingWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/validate-gateway-kyma-project-io-v1beta1-apirule,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kyma-project.io,resources=apirules,verbs=create;update,versions=v1beta1,name=v1beta1-admission.apirule.gateway.kyma-project.io,admissionReviewVersions=v1,servicePort=9443,serviceName=api-gateway-webhook-service,serviceNamespace=kyma-system,matchPolicy=Exact

type ValidatingWebhook struct{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateCreate(_ context.Context, o runtime.Object) (admission.Warnings, error) {
	v1beta1CreationCounter.Inc()
	return nil, errors.New("v1beta1 APIRule version is no longer supported, please use v2 instead")
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateUpdate(_ context.Context, _, _ runtime.Object) (admission.Warnings, error) {
	v1beta1UpdateCounter.Inc()
	return nil, errors.New("v1beta1 APIRule version is no longer supported, please use v2 instead")
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
