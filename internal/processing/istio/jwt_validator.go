package istio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kyma-project/api-gateway/api/v1beta1"
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

	hasFromHeaders, hasFromParams := false, false

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
		if len(authentication.FromHeaders) > 0 {
			if hasFromParams {
				attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".fromHeaders")
				failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mixture of multiple fromHeaders and fromParams is not supported"})
			}
			hasFromHeaders = true
		}
		if len(authentication.FromParams) > 0 {
			if hasFromHeaders {
				attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authentications", i, ".fromParams")
				failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "mixture of multiple fromHeaders and fromParams is not supported"})
			}
			hasFromParams = true
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

	authorizationsFailures := hasInvalidAuthorizations(attributePath, template.Authorizations)
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

func hasInvalidRequiredScopes(authorization v1beta1.JwtAuthorization) error {
	if authorization.RequiredScopes == nil {
		return nil
	}
	if len(authorization.RequiredScopes) == 0 {
		return errors.New("value is empty")
	}
	for _, scope := range authorization.RequiredScopes {
		if scope == "" {
			return errors.New("scope value is empty")
		}
	}
	return nil
}

func hasInvalidAudiences(authorization v1beta1.JwtAuthorization) error {
	if authorization.Audiences == nil {
		return nil
	}
	if len(authorization.Audiences) == 0 {
		return errors.New("value is empty")
	}
	for _, audience := range authorization.Audiences {
		if audience == "" {
			return errors.New("audience value is empty")
		}
	}
	return nil
}

func hasInvalidAuthorizations(attributePath string, authorizations []*v1beta1.JwtAuthorization) (failures []validation.Failure) {
	if authorizations == nil {
		return nil
	}
	if len(authorizations) == 0 {
		attrPath := fmt.Sprintf("%s%s", attributePath, ".config.authorizations")
		failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "value is empty"})
		return
	}

	for i, authorization := range authorizations {
		if authorization == nil {
			attrPath := fmt.Sprintf("%s%s[%d]", attributePath, ".config.authorizations", i)
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: "authorization is empty"})
			continue
		}

		err := hasInvalidRequiredScopes(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authorizations", i, ".requiredScopes")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: err.Error()})
		}

		err = hasInvalidAudiences(*authorization)
		if err != nil {
			attrPath := fmt.Sprintf("%s%s[%d]%s", attributePath, ".config.authorizations", i, ".audiences")
			failures = append(failures, validation.Failure{AttributePath: attrPath, Message: err.Error()})
		}
	}

	return
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

type rulesValidator struct {
}

func (v *rulesValidator) Validate(attrPath string, rules []gatewayv1beta1.Rule) []validation.Failure {
	var failures []validation.Failure
	var fromHeader *v1beta1.JwtHeader
	fromParam := ""
	for i, rule := range rules {
		for j, accessStrategy := range rule.AccessStrategies {
			attributePath := fmt.Sprintf("%s[%d].accessStrategy[%d]", attrPath, i, j)
			if accessStrategy.Config != nil {
				var template gatewayv1beta1.JwtConfig
				err := json.Unmarshal(accessStrategy.Config.Raw, &template)
				if err != nil {
					failures = append(failures, validation.Failure{AttributePath: attributePath, Message: "Can't read json: " + err.Error()})
					return failures
				}

				for k, authentication := range template.Authentications {
					if len(authentication.FromHeaders) > 0 {
						fmt.Printf("authentication[%d].FromHeaders[0].Name: %s", k, authentication.FromHeaders[0].Name)
						if fromParam != "" || (fromHeader != nil && !compareFromHeader(fromHeader, authentication.FromHeaders[0])) {
							attributeSubPath := fmt.Sprintf("%s%s[%d]", attributePath, ".config.authentications", k)
							failures = append(failures, validation.Failure{AttributePath: attributeSubPath, Message: "multiple fromHeaders and/or fromParams configuration for different rules is not supported"})
						}
						fromHeader = authentication.FromHeaders[0]
					}
					if len(authentication.FromParams) > 0 {
						if fromHeader != nil || (fromParam != "" && fromParam != authentication.FromParams[0]) {
							attributeSubPath := fmt.Sprintf("%s%s[%d]", attributePath, ".config.authentications", k)
							failures = append(failures, validation.Failure{AttributePath: attributeSubPath, Message: "multiple fromHeaders and/or fromParams configuration for different rules is not supported"})
						}
						fromParam = authentication.FromParams[0]
					}
				}
			}
		}
	}
	return failures
}

func compareFromHeader(h1 *v1beta1.JwtHeader, h2 *v1beta1.JwtHeader) bool {
	if h1 == nil || h2 == nil {
		return false
	}
	return h1.Name == h2.Name && h1.Prefix == h2.Prefix
}
