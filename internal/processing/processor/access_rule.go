package processor

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRule struct {
	client            client.Client
	ctx               context.Context
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewAccessRule(config processing.ReconciliationConfig) AccessRule {
	return AccessRule{
		client:            config.Client,
		ctx:               config.Ctx,
		additionalLabels:  config.AdditionalLabels,
		defaultDomainName: config.DefaultDomainName,
	}
}

func (r AccessRule) EvaluateReconciliation(apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, gatewayv1beta1.StatusCode, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(r.ctx, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), gatewayv1beta1.StatusSkipped, err
	}

	c := r.getObjectChanges(desired, actual)

	return c, gatewayv1beta1.StatusOK, nil
}

func (r AccessRule) getObjectChanges(desiredRules map[string]*rulev1alpha1.Rule, actualRules map[string]*rulev1alpha1.Rule) []*processing.ObjectChange {
	arApplyCommands := make(map[string]*processing.ObjectChange)

	for path, rule := range desiredRules {

		if actualRules[path] != nil {
			actualRules[path].Spec = rule.Spec
			arApplyCommands[path] = processing.NewObjectUpdateAction(actualRules[path])
		} else {
			arApplyCommands[path] = processing.NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualRules {
		if desiredRules[path] == nil {
			arApplyCommands[path] = processing.NewObjectDeleteAction(rule)
		}
	}

	applyCommands := make([]*processing.ObjectChange, 0, len(arApplyCommands))

	for _, applyCommand := range arApplyCommands {
		applyCommands = append(applyCommands, applyCommand)
	}

	return applyCommands
}

func (r AccessRule) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := processing.HasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if processing.IsSecured(rule) {
			ar := generateAccessRule(api, rule, rule.AccessStrategies, r.additionalLabels, r.defaultDomainName)
			accessRules[setAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}

func (r AccessRule) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (map[string]*rulev1alpha1.Rule, error) {
	labels := processing.GetOwnerLabels(api)

	var arList rulev1alpha1.RuleList
	if err := r.client.List(ctx, &arList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	accessRules := make(map[string]*rulev1alpha1.Rule)
	pathDuplicates := processing.HasPathDuplicates(api.Spec.Rules)

	for i := range arList.Items {
		obj := arList.Items[i]
		accessRules[setAccessRuleKey(pathDuplicates, obj)] = &obj
	}

	return accessRules, nil
}

func setAccessRuleKey(hasPathDuplicates bool, rule rulev1alpha1.Rule) string {
	// TODO: We can keep it simple by always using the path and methods as key. This way we can remove the HasPathDuplicates.
	if hasPathDuplicates {
		return fmt.Sprintf("%s:%s", rule.Spec.Match.URL, rule.Spec.Match.Methods)
	}

	return rule.Spec.Match.URL
}

func generateAccessRule(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, accessStrategies []*gatewayv1beta1.Authenticator, additionalLabels map[string]string, defaultDomainName string) *rulev1alpha1.Rule {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := processing.GenerateOwnerRef(api)

	arBuilder := builders.AccessRule().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies, defaultDomainName))).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		arBuilder.Label(k, v)
	}

	return arBuilder.Get()
}

func generateAccessRuleSpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, accessStrategies []*gatewayv1beta1.Authenticator, defaultDomainName string) *rulev1alpha1.RuleSpec {
	accessRuleSpec := builders.AccessRuleSpec().
		Match(builders.Match().
			URL(fmt.Sprintf("<http|https>://%s<%s>", helpers.GetHostWithDomain(*api.Spec.Host, defaultDomainName), rule.Path)).
			Methods(rule.Methods)).
		Authorizer(builders.Authorizer().Handler(builders.Handler().
			Name("allow"))).
		Authenticators(builders.Authenticators().From(accessStrategies)).
		Mutators(builders.Mutators().From(rule.Mutators))

	serviceNamespace := helpers.FindServiceNamespace(api, &rule)

	// Use rule level service if it exists
	if rule.Service != nil {
		return accessRuleSpec.Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *rule.Service.Name, *serviceNamespace, int(*rule.Service.Port)))).Get()
	}
	// Otherwise use service defined on APIRule spec level
	return accessRuleSpec.Upstream(builders.Upstream().
		URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, *serviceNamespace, int(*api.Spec.Service.Port)))).Get()
}
