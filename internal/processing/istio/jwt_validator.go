package istio

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	oryjwt "github.com/kyma-project/api-gateway/internal/types/ory"
	"github.com/kyma-project/api-gateway/internal/validation"
	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	istioSidecarContainerName string = "istio-proxy"
)

type handlerValidator struct{}

func (o *handlerValidator) Validate(attributePath string, handler *gatewayv1beta1.Handler) []validation.Failure {
	var failures []validation.Failure
	var template gatewayv1beta1.JwtConfig

	if !validation.ConfigNotEmpty(handler.Config) {
		failures = append(failures, validation.Failure{AttributePath: attributePath + ".config", Message: "supplied config cannot be empty"})
		return failures
	}

	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		failures = append(failures, validation.Failure{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()})
		return failures
	}

	failures = append(failures, checkForOryConfig(attributePath, handler)...)

	for i, authentication := range template.Authentications {
		invalidIssuer, err := validation.IsInvalidURL(authentication.Issuer)
		if invalidIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".issuer")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		// The https:// configuration for TrustedIssuers is not necessary in terms of security best practices,
		// however it is part of "secure by default" configuration, as this is the most common use case for iss claim.
		// If we want to allow some weaker configurations, we should have a dedicated configuration which allows that.
		unsecuredIssuer, err := validation.IsUnsecuredURL(authentication.Issuer)
		if unsecuredIssuer {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".issuer")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
		invalidJwksUri, err := validation.IsInvalidURL(authentication.JwksUri)
		if invalidJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".jwksUri")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is empty or not a valid url err=%s", err)})
		}
		unsecuredJwksUri, err := validation.IsUnsecuredURL(authentication.JwksUri)
		if unsecuredJwksUri {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".jwksUri")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("value is not a secured url err=%s", err)})
		}
		if len(authentication.FromHeaders) > 1 {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".fromHeaders")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "multiple fromHeaders are not supported"})
		}
		if len(authentication.FromParams) > 1 {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".fromParams")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "multiple fromParams are not supported"})
		}
	}

	authorizationsFailures := validation.HasInvalidAuthorizations(attributePath, template.Authorizations)
	failures = append(failures, authorizationsFailures...)

	return failures
}

func checkForOryConfig(attributePath string, handler *gatewayv1beta1.Handler) (problems []validation.Failure) {
	var template oryjwt.JWTAccStrConfig
	err := json.Unmarshal(handler.Config.Raw, &template)
	if err != nil {
		return []validation.Failure{{AttributePath: attributePath + ".config", Message: "Can't read json: " + err.Error()}}
	}

	if len(template.JWKSUrls) > 0 {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config" + ".jwks_urls", Message: "Configuration for jwks_urls is not supported with Istio handler"})
	}

	if len(template.RequiredScopes) > 0 {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config" + ".required_scopes", Message: "Configuration for required_scopes is not supported with Istio handler"})
	}

	if len(template.TrustedIssuers) > 0 {
		problems = append(problems, validation.Failure{AttributePath: attributePath + ".config" + ".trusted_issuers", Message: "Configuration for trusted_issuers is not supported with Istio handler"})
	}

	return problems
}

type injectionValidator struct {
	ctx    context.Context
	client client.Client
}

func (v *injectionValidator) Validate(attributePath string, selector *apiv1beta1.WorkloadSelector, namespace string) (problems []validation.Failure, err error) {
	var podList corev1.PodList
	err = v.client.List(v.ctx, &podList, client.MatchingLabels(selector.MatchLabels))
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if !containsSidecar(pod) {
			problems = append(problems, validation.Failure{AttributePath: attributePath, Message: fmt.Sprintf("Pod %s/%s does not have an injected istio sidecar", pod.Namespace, pod.Name)})
		}
	}
	return problems, nil
}

func containsSidecar(pod corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == istioSidecarContainerName {
			return true
		}
	}
	return false
}
