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

	"github.com/kyma-project/api-gateway/internal/access"

	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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
	return ctrl.NewWebhookManagedBy(mgr, ruleV1).
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

func (w *ValidatingWebhook) shouldBlockAPIRule() (bool, error) {
	accessedAllowed, err := access.ShouldAllowAccessToV1Beta1(context.Background(), w.Client)
	if err != nil {
		return true, err
	}

	if accessedAllowed {
		return false, nil
	}

	return true, nil
}

func (w *ValidatingWebhook) validationError() error {
	block, err := w.shouldBlockAPIRule()
	if err != nil || block {
		return errors.New("v1beta1 APIRule version is no longer supported, please use v2 instead")
	}
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateCreate(_ context.Context, _ *APIRule) (admission.Warnings, error) {
	v1beta1CreateCounter.Inc()
	return nil, w.validationError()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateUpdate(_ context.Context, _, _ *APIRule) (admission.Warnings, error) {
	v1beta1UpdateCounter.Inc()
	return nil, w.validationError()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ValidatingWebhook) ValidateDelete(_ context.Context, _ *APIRule) (admission.Warnings, error) {
	v1beta1DeleteCounter.Inc()
	return nil, w.validationError()
}
