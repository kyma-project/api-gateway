package processors

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"github.com/kyma-project/api-gateway/internal/processing"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AccessRuleProcessor is the generic processor that handles the Ory Rules in the reconciliation of API Rule.
type AccessRuleProcessor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator AccessRuleCreator
}

// AccessRuleCreator provides the creation of Rules using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AccessRuleCreator interface {
	Create(api *gatewayv2alpha1.APIRule) map[string]*rulev1alpha1.Rule
}

func (r AccessRuleProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired := r.getDesiredState(r.ApiRule)
	actual, err := r.getActualState(ctx, client, r.ApiRule)
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

func (r AccessRuleProcessor) getDesiredState(api *gatewayv2alpha1.APIRule) map[string]*rulev1alpha1.Rule {
	return r.Creator.Create(api)
}

func (r AccessRuleProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*rulev1alpha1.Rule, error) {
	labels := processing.GetOwnerLabels(api)

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

func HasPathDuplicates(rules []gatewayv2alpha1.Rule) bool {
	duplicates := map[string]bool{}
	for _, rule := range rules {
		if duplicates[rule.Path] {
			return true
		}
		duplicates[rule.Path] = true
	}

	return false
}
