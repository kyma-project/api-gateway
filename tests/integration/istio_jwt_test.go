package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"

	"github.com/cucumber/godog"
)

type tokenFrom struct {
	From     string
	Prefix   string
	AsHeader bool
}

type istioJwtManifestScenario struct {
	*ScenarioWithRawAPIResource
}

func initIstioJwtScenarios(ctx *godog.ScenarioContext) {
	initCommon(ctx)
	initRequiredScopes(ctx)
	initAudience(ctx)
	initJwtAndAllow(ctx)
	initJwtAndOauth(ctx)
	initJwtTwoNamespaces(ctx)
	initJwtServiceFallback(ctx)
	initDiffServiceSameMethods(ctx)
	initJwtUnavailableIssuer(ctx)
	initJwtIssuerJwksNotMatch(ctx)
	initMutatorCookie(ctx)
	initMutatorHeader(ctx)
	initMultipleMutators(ctx)
	initMutatorsOverwrite(ctx)
	initTokenFromHeaders(ctx)
	initTokenFromParams(ctx)
}

func (s *istioJwtManifestScenario) theAPIRuleIsApplied() error {
	resource, err := manifestprocessor.ParseFromFileWithTemplate(s.apiResourceManifestPath, s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, resource)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithAValidToken(path, tokenType, _, _ string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     authorizationHeaderName,
		Prefix:   authorizationHeaderPrefix,
		AsHeader: true,
	}
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     authorizationHeaderName,
		Prefix:   authorizationHeaderPrefix,
		AsHeader: true,
	}
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween(path, tokenType string, fromHeader string, prefix string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromHeader,
		Prefix:   prefix,
		AsHeader: true,
	}
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween(path, tokenType string, fromParameter string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromParameter,
		AsHeader: false,
	}
	fmt.Printf("callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween: %s", fromParameter)
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, asserter, nil, &tokenFrom)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenShouldResultInBodyContaining(path, tokenType string, bodyContent string) error {
	asserter := &helpers.BodyContainsPredicate{Expected: []string{bodyContent}}
	tokenFrom := tokenFrom{
		From:     authorizationHeaderName,
		Prefix:   authorizationHeaderPrefix,
		AsHeader: true,
	}
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, asserter, nil, &tokenFrom)
}

func callingEndpointWithHeadersWithRetries(url string, path string, tokenType string, asserter helpers.HttpResponseAsserter, requestHeaders map[string]string, tokenFrom *tokenFrom) error {
	if requestHeaders == nil {
		requestHeaders = make(map[string]string)
	}

	switch tokenType {
	case "Opaque":
		token, err := jwt.GetAccessToken(*oauth2Cfg, jwtConfig, strings.ToLower(tokenType))
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		requestHeaders[opaqueHeaderName] = token
	case "JWT":
		token, err := jwt.GetAccessToken(*oauth2Cfg, jwtConfig, strings.ToLower(tokenType))
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		if tokenFrom.From == "" {
			return errors.New("jwt from header or parameter name not specified")
		}
		if tokenFrom.AsHeader {
			if tokenFrom.Prefix != "" {
				token = fmt.Sprintf("%s %s", tokenFrom.Prefix, token)
			}
			requestHeaders[tokenFrom.From] = token
		} else {
			path = fmt.Sprintf("%s?%s=%s", path, tokenFrom.From, token)
		}
	default:
		return fmt.Errorf("unsupported token type: %s", tokenType)
	}

	return helper.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", url, path), asserter)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *istioJwtManifestScenario) thereAreTwoNamespaces() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("second-namespace.yaml", s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	_, err = batch.CreateResources(k8sClient, resources...)
	return err
}

func (s *istioJwtManifestScenario) thereIsAnJwtSecuredPath(path string) {
	s.manifestTemplate["jwtSecuredPath"] = path
}

func (s *istioJwtManifestScenario) emptyStep() {
}

func (s *istioJwtManifestScenario) thereIsAHttpbinService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app.yaml", s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	_, err = batch.CreateResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	s.url = fmt.Sprintf("https://httpbin-%s.%s", s.testID, s.domain)

	return nil
}
