package migration

import (
	"encoding/json"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewAccessRuleProcessor returns a AccessRuleProcessor with the desired state handling specific for the Ory handler.
func NewAccessRuleProcessor(config processing.ReconciliationConfig) processors.AccessRuleProcessor {
	return processors.AccessRuleProcessor{
		Creator: accessRuleCreator{
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type accessRuleCreator struct {
	defaultDomainName string
}

// Create returns a map of rules using the configuration of the APIRule. The key of the map is a unique combination of
// the match URL and methods of the rule.
func (r accessRuleCreator) Create(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := processors.HasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if processing.IsSecuredByOathkeeper(rule) {
			ctrl.Log.Info("Processing rule", "rule", rule, "access strategies", rule.AccessStrategies)

			// We need to migrate the access strategies to Ory format, because the APIRule in v2alpha1 uses the Istio format and applying
			// this directly to the Ory Rule will lead to an incorrect configuration and 500 response codes from Oathkeeper.
			accessStrategies, err := migrateAccessStrategies(rule.AccessStrategies)
			if err != nil {
				// TODO handle error
				return nil
			}
			ctrl.Log.Info("Migrated access strategies", "accessStrategies", accessStrategies)
			ar := processors.GenerateAccessRule(api, rule, accessStrategies, r.defaultDomainName)
			accessRules[processors.SetAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}

func migrateAccessStrategies(accessStrategies []*gatewayv1beta1.Authenticator) ([]*gatewayv1beta1.Authenticator, error) {
	var migrated []*gatewayv1beta1.Authenticator

	for _, strategy := range accessStrategies {

		ctrl.Log.Info("Migrating access strategy handler", "handler", strategy.Handler.Name)

		switch strategy.Handler.Name {
		case gatewayv1beta1.AccessStrategyJwt:

			rawConfig, err := getRawJwtConfig(strategy)
			if err != nil {
				return nil, err
			}

			accessStrategy := &gatewayv1beta1.Authenticator{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyJwt,
					Config: &runtime.RawExtension{
						Raw: rawConfig,
					},
				},
			}

			migrated = append(migrated, accessStrategy)
		default:
			migrated = append(migrated, strategy)
		}
	}

	return migrated, nil
}

func getRawJwtConfig(strategy *gatewayv1beta1.Authenticator) ([]byte, error) {
	jwtConfig, err := v2alpha1.ConvertIstioJwtAccessStrategy(strategy)
	if err != nil {
		return nil, err
	}

	var trustedIssuers []string
	var jwksUrls []string
	for _, a := range jwtConfig.Authentications {
		trustedIssuers = append(trustedIssuers, a.Issuer)
		jwksUrls = append(jwksUrls, a.JwksUri)
	}

	var audiences []string
	var requiredScopes []string
	for _, a := range jwtConfig.Authorizations {
		audiences = append(audiences, a.Audiences...)
		requiredScopes = append(requiredScopes, a.RequiredScopes...)
	}

	config := ory.JWTAccStrConfig{
		TrustedIssuers: trustedIssuers,
		JWKSUrls:       jwksUrls,
		RequiredScopes: requiredScopes,
		TargetAudience: audiences,
	}

	return json.Marshal(config)
}
