package processing

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRuleProcessor struct {
	Creator           AccessRuleCreator
	additionalLabels  map[string]string
	defaultDomainName string
}

type AccessRuleCreator interface {
	Create(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule
}

func (r AccessRuleProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*ObjectChange, 0), err
	}

	c := r.getObjectChanges(desired, actual)

	return c, nil
}

func (r AccessRuleProcessor) getObjectChanges(desiredRules map[string]*rulev1alpha1.Rule, actualRules map[string]*rulev1alpha1.Rule) []*ObjectChange {
	arChanges := make(map[string]*ObjectChange)

	for path, rule := range desiredRules {

		if actualRules[path] != nil {
			actualRules[path].Spec = rule.Spec
			arChanges[path] = NewObjectUpdateAction(actualRules[path])
		} else {
			arChanges[path] = NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualRules {
		if desiredRules[path] == nil {
			arChanges[path] = NewObjectDeleteAction(rule)
		}
	}

	arChangesToApply := make([]*ObjectChange, 0, len(arChanges))

	for _, applyCommand := range arChanges {
		arChangesToApply = append(arChangesToApply, applyCommand)
	}

	return arChangesToApply
}

func (r AccessRuleProcessor) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	return r.Creator.Create(api)
}

func (r AccessRuleProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]*rulev1alpha1.Rule, error) {
	labels := GetOwnerLabels(api)

	var arList rulev1alpha1.RuleList
	if err := client.List(ctx, &arList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	accessRules := make(map[string]*rulev1alpha1.Rule)
	pathDuplicates := HasPathDuplicates(api.Spec.Rules)

	for i := range arList.Items {
		obj := arList.Items[i]
		accessRules[SetAccessRuleKey(pathDuplicates, obj)] = &obj
	}

	return accessRules, nil
}

func SetAccessRuleKey(hasPathDuplicates bool, rule rulev1alpha1.Rule) string {

	if hasPathDuplicates {
		return fmt.Sprintf("%s:%s", rule.Spec.Match.URL, rule.Spec.Match.Methods)
	}

	return rule.Spec.Match.URL
}

func HasPathDuplicates(rules []gatewayv1beta1.Rule) bool {
	duplicates := map[string]bool{}
	for _, rule := range rules {
		if duplicates[rule.Path] {
			return true
		}
		duplicates[rule.Path] = true
	}

	return false
}
