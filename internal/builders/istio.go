package builders

import (
	"encoding/json"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"istio.io/api/security/v1beta1"
	apiv1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// AuthorizationPolicyBuilder returns a builder for istio.io/client-go/pkg/apis/security/v1beta1/AuthorizationPolicy type
func AuthorizationPolicyBuilder() *AuthorizationPolicy {
	return &AuthorizationPolicy{
		value: &securityv1beta1.AuthorizationPolicy{},
	}
}

type AuthorizationPolicy struct {
	value *securityv1beta1.AuthorizationPolicy
}

func (ap *AuthorizationPolicy) Get() *securityv1beta1.AuthorizationPolicy {
	return ap.value
}

func (ap *AuthorizationPolicy) From(val *securityv1beta1.AuthorizationPolicy) *AuthorizationPolicy {
	ap.value = val
	return ap
}

func (ap *AuthorizationPolicy) Name(val string) *AuthorizationPolicy {
	ap.value.Name = val
	return ap
}

func (ap *AuthorizationPolicy) GenerateName(val string) *AuthorizationPolicy {
	ap.value.Name = ""
	ap.value.GenerateName = val
	return ap
}

func (ap *AuthorizationPolicy) Namespace(val string) *AuthorizationPolicy {
	ap.value.Namespace = val
	return ap
}

func (ap *AuthorizationPolicy) Owner(val *ownerReference) *AuthorizationPolicy {
	ap.value.OwnerReferences = append(ap.value.OwnerReferences, *val.Get())
	return ap
}

func (ap *AuthorizationPolicy) Label(key, val string) *AuthorizationPolicy {
	if ap.value.Labels == nil {
		ap.value.Labels = make(map[string]string)
	}
	ap.value.Labels[key] = val
	return ap
}

func (ap *AuthorizationPolicy) Spec(val *AuthorizationPolicySpec) *AuthorizationPolicy {
	ap.value.Spec = *val.Get()
	return ap
}

// AuthorizationPolicySpecBuilder returns builder for istio.io/api/security/v1beta1/AuthorizationPolicy type
func AuthorizationPolicySpecBuilder() *AuthorizationPolicySpec {
	return &AuthorizationPolicySpec{
		value: &v1beta1.AuthorizationPolicy{},
	}
}

type AuthorizationPolicySpec struct {
	value *v1beta1.AuthorizationPolicy
}

func (aps *AuthorizationPolicySpec) Get() *v1beta1.AuthorizationPolicy {
	return aps.value
}

func (aps *AuthorizationPolicySpec) From(val *v1beta1.AuthorizationPolicy) *AuthorizationPolicySpec {
	aps.value = val
	return aps
}

func (aps *AuthorizationPolicySpec) Selector(val *Selector) *AuthorizationPolicySpec {
	aps.value.Selector = val.Get()
	return aps
}

func (aps *AuthorizationPolicySpec) Rule(val *Rule) *AuthorizationPolicySpec {
	aps.value.Rules = append(aps.value.Rules, val.Get())
	return aps
}

// RuleBuilder returns builder for istio.io/api/security/v1beta1/Rule type
func RuleBuilder() *Rule {
	return &Rule{
		value: &v1beta1.Rule{},
	}
}

type Rule struct {
	value *v1beta1.Rule
}

func (r *Rule) Get() *v1beta1.Rule {
	return r.value
}

func (r *Rule) RuleFrom(val *RuleFrom) *Rule {
	r.value.From = append(r.value.From, val.Get())
	return r
}

func (r *Rule) RuleTo(val *RuleTo) *Rule {
	r.value.To = append(r.value.To, val.Get())
	return r
}

// RuleFromBuilder returns builder for istio.io/api/security/v1beta1/Rule_From type
func RuleFromBuilder() *RuleFrom {
	return &RuleFrom{
		value: &v1beta1.Rule_From{},
	}
}

type RuleFrom struct {
	value *v1beta1.Rule_From
}

func (rf *RuleFrom) Get() *v1beta1.Rule_From {
	return rf.value
}

func (rf *RuleFrom) Source() *RuleFrom {
	// Only support one source at the moment
	source := v1beta1.Source{RequestPrincipals: []string{"*"}}
	rf.value.Source = &source
	return rf
}

func (rf *RuleFrom) IngressGatewaySource() *RuleFrom {
	source := v1beta1.Source{Principals: []string{"cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"}}
	rf.value.Source = &source
	return rf
}

func (rf *RuleFrom) OathkeeperProxySource() *RuleFrom {
	source := v1beta1.Source{Principals: []string{"cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"}}
	rf.value.Source = &source
	return rf
}

// RuleToBuilder returns builder for istio.io/apis/security/v1beta1/Rule_To type
func RuleToBuilder() *RuleTo {
	return &RuleTo{
		value: &v1beta1.Rule_To{},
	}
}

type RuleTo struct {
	value *v1beta1.Rule_To
}

func (rt *RuleTo) Get() *v1beta1.Rule_To {
	return rt.value
}

func (rt *RuleTo) Operation(val *Operation) *RuleTo {
	rt.value.Operation = val.Get()
	return rt
}

// OperationBuilder returns builder for istio.io/api/security/v1beta1/Operation type
func OperationBuilder() *Operation {
	return &Operation{
		value: &v1beta1.Operation{},
	}
}

type Operation struct {
	value *v1beta1.Operation
}

func (o *Operation) Get() *v1beta1.Operation {
	return o.value
}

func (o *Operation) Methods(val []string) *Operation {
	o.value.Methods = val
	return o
}

func (o *Operation) Path(val string) *Operation {
	o.value.Paths = append(o.value.Paths, val)
	return o
}

// RequestAuthenticationBuilder returns a builder for istio.io/client-go/pkg/apis/security/v1beta1/RequestAuthentication type
func RequestAuthenticationBuilder() *RequestAuthentication {
	return &RequestAuthentication{
		value: &securityv1beta1.RequestAuthentication{},
	}
}

type RequestAuthentication struct {
	value *securityv1beta1.RequestAuthentication
}

func (ra *RequestAuthentication) Get() *securityv1beta1.RequestAuthentication {
	return ra.value
}

func (ra *RequestAuthentication) From(val *securityv1beta1.RequestAuthentication) *RequestAuthentication {
	ra.value = val
	return ra
}

func (ra *RequestAuthentication) Name(val string) *RequestAuthentication {
	ra.value.Name = val
	return ra
}

func (ra *RequestAuthentication) GenerateName(val string) *RequestAuthentication {
	ra.value.Name = ""
	ra.value.GenerateName = val
	return ra
}

func (ra *RequestAuthentication) Namespace(val string) *RequestAuthentication {
	ra.value.Namespace = val
	return ra
}

func (ra *RequestAuthentication) Owner(val *ownerReference) *RequestAuthentication {
	ra.value.OwnerReferences = append(ra.value.OwnerReferences, *val.Get())
	return ra
}

func (ra *RequestAuthentication) Label(key, val string) *RequestAuthentication {
	if ra.value.Labels == nil {
		ra.value.Labels = make(map[string]string)
	}
	ra.value.Labels[key] = val
	return ra
}

func (ra *RequestAuthentication) Spec(val *RequestAuthenticationSpec) *RequestAuthentication {
	ra.value.Spec = *val.Get()
	return ra
}

// RequestAuthenticationSpecBuilder returns a builder for istio.io/api/security/v1beta1/RequestAuthentication type
func RequestAuthenticationSpecBuilder() *RequestAuthenticationSpec {
	return &RequestAuthenticationSpec{
		value: &v1beta1.RequestAuthentication{},
	}
}

type RequestAuthenticationSpec struct {
	value *v1beta1.RequestAuthentication
}

func (ras *RequestAuthenticationSpec) Get() *v1beta1.RequestAuthentication {
	return ras.value
}

func (ras *RequestAuthenticationSpec) From(val *v1beta1.RequestAuthentication) *RequestAuthenticationSpec {
	ras.value = val
	return ras
}

func (ras *RequestAuthenticationSpec) Selector(val *Selector) *RequestAuthenticationSpec {
	ras.value.Selector = val.Get()
	return ras
}

func (ras *RequestAuthenticationSpec) JwtRules(val *JwtRule) *RequestAuthenticationSpec {
	ras.value.JwtRules = *val.Get()
	return ras
}

// JwtRuleBuilder returns builder for istio.io/api/security/v1beta1/JWTRule type
func JwtRuleBuilder() *JwtRule {
	return &JwtRule{
		value: &[]*v1beta1.JWTRule{},
	}
}

type JwtRule struct {
	value *[]*v1beta1.JWTRule
}

func (jr *JwtRule) Get() *[]*v1beta1.JWTRule {
	return jr.value
}

func (jr *JwtRule) From(val []*gatewayv1beta1.Authenticator) *JwtRule {
	for _, accessStrategy := range val {
		authentications := &Authentications{
			Authentications: []*Authentication{},
		}
		if accessStrategy.Config != nil {
			_ = json.Unmarshal(accessStrategy.Config.Raw, authentications)
		}
		for _, authentication := range authentications.Authentications {
			*jr.value = append(*jr.value, &v1beta1.JWTRule{
				Issuer:  authentication.Issuer,
				JwksUri: authentication.JwksUri,
			})
		}
	}
	return jr
}

// SelectorBuilder returns builder for istio.io/api/type/v1beta1/WorkloadSelector type
func SelectorBuilder() *Selector {
	return &Selector{
		value: &apiv1beta1.WorkloadSelector{},
	}
}

type Selector struct {
	value *apiv1beta1.WorkloadSelector
}

func (s *Selector) Get() *apiv1beta1.WorkloadSelector {
	return s.value
}

func (s *Selector) MatchLabels(key, val string) *Selector {
	if s.value.MatchLabels == nil {
		s.value.MatchLabels = make(map[string]string)
	}
	s.value.MatchLabels[key] = val
	return s
}

type Authentications struct {
	Authentications []*Authentication `json:"authentications"`
}

type Authentication struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}
