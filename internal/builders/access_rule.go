package builders

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AccessRule returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Rule instances
func AccessRule() *accessRule {
	return &accessRule{
		value: &rulev1alpha1.Rule{},
	}
}

type accessRule struct {
	value *rulev1alpha1.Rule
}

func (ar *accessRule) From(val *rulev1alpha1.Rule) *accessRule {
	ar.value = val
	return ar
}

func (ar *accessRule) Name(val string) *accessRule {
	ar.value.Name = val
	return ar
}

func (ar *accessRule) GenerateName(val string) *accessRule {
	ar.value.Name = ""
	ar.value.GenerateName = val
	return ar
}

func (ar *accessRule) Namespace(val string) *accessRule {
	ar.value.Namespace = val
	return ar
}

func (ar *accessRule) Owner(val *ownerReference) *accessRule {
	ar.value.OwnerReferences = append(ar.value.OwnerReferences, *val.Get())
	return ar
}

func (ar *accessRule) Label(key, val string) *accessRule {
	if ar.value.Labels == nil {
		ar.value.Labels = make(map[string]string)
	}
	ar.value.Labels[key] = val
	return ar
}

func (ar *accessRule) Spec(val *accessRuleSpec) *accessRule {
	ar.value.Spec = *val.Get()
	return ar
}

func (ar *accessRule) Get() *rulev1alpha1.Rule {
	return ar.value
}

// AccessRuleSpec returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/RuleSpec instances
func AccessRuleSpec() *accessRuleSpec {
	return &accessRuleSpec{
		value: &rulev1alpha1.RuleSpec{},
	}
}

type accessRuleSpec struct {
	value *rulev1alpha1.RuleSpec
}

func (ars *accessRuleSpec) From(val *rulev1alpha1.RuleSpec) *accessRuleSpec {
	ars.value = val
	return ars
}

func (ars *accessRuleSpec) Get() *rulev1alpha1.RuleSpec {
	return ars.value
}

func (ars *accessRuleSpec) Upstream(val *upstream) *accessRuleSpec {
	ars.value.Upstream = val.Get()
	return ars
}

func (ars *accessRuleSpec) Match(val *match) *accessRuleSpec {
	ars.value.Match = val.Get()
	return ars
}
func (ars *accessRuleSpec) Authorizer(val *authorizer) *accessRuleSpec {
	ars.value.Authorizer = val.Get()
	return ars
}
func (ars *accessRuleSpec) Authenticators(val *authenticators) *accessRuleSpec {
	ars.value.Authenticators = val.Get()
	return ars
}
func (ars *accessRuleSpec) Mutators(val *mutators) *accessRuleSpec {
	ars.value.Mutators = val.Get()
	return ars
}

// Upstream returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Upstream instances
func Upstream() *upstream {
	return &upstream{
		value: &rulev1alpha1.Upstream{},
	}
}

type upstream struct {
	value *rulev1alpha1.Upstream
}

func (u *upstream) URL(val string) *upstream {
	u.value.URL = val
	return u
}

func (u *upstream) StripPath(val *string) *upstream {
	u.value.StripPath = val
	return u
}

func (u *upstream) PreserveHost(val *bool) *upstream {
	u.value.PreserveHost = val
	return u
}

func (u *upstream) Get() *rulev1alpha1.Upstream {
	return u.value
}

// Match returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Match instances
func Match() *match {
	return &match{
		value: &rulev1alpha1.Match{},
	}
}

type match struct {
	value *rulev1alpha1.Match
}

func (m *match) URL(val string) *match {
	m.value.URL = val
	return m
}

func (m *match) Methods(val []string) *match {
	m.value.Methods = val
	return m
}

func (m *match) Get() *rulev1alpha1.Match {
	return m.value
}

// Handler returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Handler instances
func Handler() *handler {
	return &handler{
		value: &rulev1alpha1.Handler{},
	}
}

type handler struct {
	value *rulev1alpha1.Handler
}

func (h *handler) Get() *rulev1alpha1.Handler {
	return h.value
}

func (h *handler) Name(val string) *handler {
	h.value.Name = val
	return h
}

func (h *handler) Config(val *runtime.RawExtension) *handler {
	h.value.Config = val
	return h
}

// Authorizer returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Authorizer instances
func Authorizer() *authorizer {
	return &authorizer{
		value: &rulev1alpha1.Authorizer{},
	}
}

type authorizer struct {
	value *rulev1alpha1.Authorizer
}

func (a *authorizer) Handler(val *handler) *authorizer {
	a.value.Handler = val.Get()
	return a
}

func (a *authorizer) Get() *rulev1alpha1.Authorizer {
	return a.value
}

func (a *authorizer) From(val *rulev1alpha1.Authorizer) *authorizer {
	a.value = val
	return a
}

// Authenticators returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Authenticators instances
func Authenticators() *authenticators {
	return &authenticators{
		value: []*rulev1alpha1.Authenticator{},
	}
}

type authenticators struct {
	value []*rulev1alpha1.Authenticator
}

func (a *authenticators) Handler(val *handler) *authenticators {
	a.value = append(a.value, &rulev1alpha1.Authenticator{Handler: val.Get()})
	return a
}

func (a *authenticators) Get() []*rulev1alpha1.Authenticator {
	return a.value
}

func (a *authenticators) From(val []*gatewayv1beta1.Authenticator) *authenticators {
	if val == nil {
		a.value = nil
	} else {
		targetList := make([]*rulev1alpha1.Authenticator, len(val))
		for i := 0; i < len(val); i++ {
			targetObj := rulev1alpha1.Authenticator{
				Handler: convertHandler(val[i].Handler),
			}
			targetList[i] = &targetObj
		}
		a.value = targetList
	}

	return a
}

// Mutators returns a builder for github.com/ory/oathkeeper-maester/api/v1alpha1/Mutators instances
func Mutators() *mutators {
	return &mutators{
		value: []*rulev1alpha1.Mutator{},
	}
}

type mutators struct {
	value []*rulev1alpha1.Mutator
}

func (m *mutators) Handler(val *handler) *mutators {
	m.value = append(m.value, &rulev1alpha1.Mutator{Handler: val.Get()})
	return m
}

func (m *mutators) Get() []*rulev1alpha1.Mutator {
	return m.value
}

func (m *mutators) From(val []*gatewayv1beta1.Mutator) *mutators {
	if val == nil {
		m.value = nil
	} else {
		targetList := make([]*rulev1alpha1.Mutator, len(val))
		for i := 0; i < len(val); i++ {
			targetObj := rulev1alpha1.Mutator{
				Handler: convertHandler(val[i].Handler),
			}
			targetList[i] = &targetObj
		}
		m.value = targetList
	}

	return m
}

func convertHandler(src *gatewayv1beta1.Handler) *rulev1alpha1.Handler {
	if src == nil {
		return nil
	}

	res := rulev1alpha1.Handler{
		Name:   src.Name,
		Config: src.Config,
	}

	return &res
}
