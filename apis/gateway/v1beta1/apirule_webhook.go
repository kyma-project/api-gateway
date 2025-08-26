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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	configMapName      = "api-gateway-config.operator.kyma-project.io"
	configMapNamespace = "kyma-system"
	configMapKey       = "enableDeprecatedV1beta1APIRule"
)

var (
	v1beta1CreateCounter prometheus.Counter
	v1beta1UpdateCounter prometheus.Counter
	v1beta1DeleteCounter prometheus.Counter
)

func (ruleV1 *APIRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	v1beta1CreateCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "apirule_v1beta1_create_actions_total",
		Namespace: "api_gateway",
		Help:      "The total number of APIRule v1beta1 CREATE actions received",
	})
	v1beta1UpdateCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "apirule_v1beta1_update_actions_total",
		Namespace: "api_gateway",
		Help:      "The total number of APIRule v1beta1 UPDATE actions received",
	})
	v1beta1DeleteCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "apirule_v1beta1_delete_actions_total",
		Namespace: "api_gateway",
		Help:      "The total number of APIRule v1beta1 DELETE actions received",
	})
	ctrlmetrics.Registry.MustRegister(v1beta1CreateCounter)
	ctrlmetrics.Registry.MustRegister(v1beta1UpdateCounter)
	ctrlmetrics.Registry.MustRegister(v1beta1DeleteCounter)
	return ctrl.NewWebhookManagedBy(mgr).
		For(ruleV1).
		WithValidator(&ValidatingWebhook{
			Client: mgr.GetClient(),
		}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-gateway-kyma-project-io-v1beta1-apirule,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kyma-project.io,resources=apirules,verbs=create;update;delete,versions=v1beta1,name=v1beta1-admission.apirule.gateway.kyma-project.io,admissionReviewVersions=v1,servicePort=9443,serviceName=api-gateway-webhook-service,serviceNamespace=kyma-system,matchPolicy=Exact
// +kubebuilder:object:generate=false
type ValidatingWebhook struct {
	Client client.Client
}

func (w *ValidatingWebhook) shouldBlockAPIRule() bool {
	configMap := &corev1.ConfigMap{}
	err := w.Client.Get(context.Background(), types.NamespacedName{
		Namespace: configMapNamespace,
		Name:      configMapName,
	}, configMap)
	if err != nil {
		return true
	}

	if configMap.Data[configMapKey] == "true" {
		return false
	}

	return true
}

func (w *ValidatingWebhook) validationError() error {
	block := w.shouldBlockAPIRule()
	if block {
		return errors.New("v1beta1 APIRule version is no longer supported, please use v2 instead")
	}
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateCreate(_ context.Context, o runtime.Object) (admission.Warnings, error) {
	v1beta1CreateCounter.Inc()
	return nil, w.validationError()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateUpdate(_ context.Context, _, _ runtime.Object) (admission.Warnings, error) {
	v1beta1UpdateCounter.Inc()
	return nil, w.validationError()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	v1beta1DeleteCounter.Inc()
	return nil, w.validationError()
}
