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

	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = ctrl.Log.WithName("controllers").WithName("Api")

var globalWebhookClient client.Client
var globalWebhookContext context.Context

func (r *APIRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	globalWebhookClient = mgr.GetClient()
	globalWebhookContext = context.Background()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-gateway-kyma-project-io-v1beta1-apirule,mutating=true,failurePolicy=fail,groups=gateway.kyma-project.io,resources=apirules,verbs=create;update,versions=v1beta1,name=mapirule.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &APIRule{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *APIRule) Default() {
	log.Info("WEBHOOK->Default()")
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-gateway-kyma-project-io-v1beta1-apirule,mutating=false,failurePolicy=fail,groups=gateway.kyma-project.io,resources=apirules,versions=v1beta1,name=vapirule.kb.io,admissionReviewVersions=v1beta1,sideEffects=None

var _ webhook.Validator = &APIRule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateCreate() error {
	log.Info("WEBHOOK->ValidateCreate()")
	return errors.New("ErrorOnCreate")
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateUpdate(old runtime.Object) error {
	log.Info("WEBHOOK->ValidateUpdate()")
	return errors.New("ErrorOnUpdate")
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateDelete() error {
	return nil
}
