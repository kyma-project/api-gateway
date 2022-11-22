package ory

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRuleProcessor struct {
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewAccessRuleProcessor(config processing.ReconciliationConfig) AccessRuleProcessor {
	return AccessRuleProcessor{
		additionalLabels:  config.AdditionalLabels,
		defaultDomainName: config.DefaultDomainName,
	}
}

func (r AccessRuleProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	c := r.getObjectChanges(desired, actual)

	return c, nil
}

func (r AccessRuleProcessor) getObjectChanges(desiredRules map[string]*rulev1alpha1.Rule, actualRules map[string]*rulev1alpha1.Rule) []*processing.ObjectChange {
	arChanges := make(map[string]*processing.ObjectChange)

	for path, rule := range desiredRules {

		if actualRules[path] != nil {
			actualRules[path].Spec = rule.Spec
			arChanges[path] = processing.NewObjectUpdateAction(actualRules[path])
		} else {
			arChanges[path] = processing.NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualRules {
		if desiredRules[path] == nil {
			arChanges[path] = processing.NewObjectDeleteAction(rule)
		}
	}

	arChangesToApply := make([]*processing.ObjectChange, 0, len(arChanges))

	for _, applyCommand := range arChanges {
		arChangesToApply = append(arChangesToApply, applyCommand)
	}

	return arChangesToApply
}

func (r AccessRuleProcessor) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := hasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if processing.IsSecured(rule) {
			ar := generateAccessRule(api, rule, rule.AccessStrategies, r.additionalLabels, r.defaultDomainName)
			accessRules[setAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}

func (r AccessRuleProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]*rulev1alpha1.Rule, error) {
	labels := processing.GetOwnerLabels(api)

	var arList rulev1alpha1.RuleList
	if err := client.List(ctx, &arList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	accessRules := make(map[string]*rulev1alpha1.Rule)
	pathDuplicates := hasPathDuplicates(api.Spec.Rules)

	for i := range arList.Items {
		obj := arList.Items[i]
		accessRules[setAccessRuleKey(pathDuplicates, obj)] = &obj
	}

	return accessRules, nil
}

func setAccessRuleKey(hasPathDuplicates bool, rule rulev1alpha1.Rule) string {
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

	if rule.Service != nil {
		return accessRuleSpec.Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *rule.Service.Name, *serviceNamespace, int(*rule.Service.Port)))).Get()
	} else {
		return accessRuleSpec.Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, *serviceNamespace, int(*api.Spec.Service.Port)))).Get()
	}

}

func hasPathDuplicates(rules []gatewayv1beta1.Rule) bool {
	duplicates := map[string]bool{}
	for _, rule := range rules {
		if duplicates[rule.Path] {
			return true
		}
		duplicates[rule.Path] = true
	}

	return false
}
