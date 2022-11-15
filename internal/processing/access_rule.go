package processing

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRuleProcessor struct {
	client            client.Client
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewAccessRuleProcessor(client client.Client, additionalLabels map[string]string, defaultDomainName string) AccessRuleProcessor {
	return AccessRuleProcessor{
		client:            client,
		additionalLabels:  additionalLabels,
		defaultDomainName: defaultDomainName,
	}
}

func (a *AccessRuleProcessor) GetDiff(desiredRules map[string]*rulev1alpha1.Rule, actualRules map[string]*rulev1alpha1.Rule) []*ObjToPatch {
	arPatch := make(map[string]*ObjToPatch)

	for path, rule := range desiredRules {

		if actualRules[path] != nil {
			modifyAccessRule(actualRules[path], rule)
			arPatch[path] = NewUpdateObjectAction(actualRules[path])
		} else {
			arPatch[path] = NewCreateObjectAction(rule)
		}

	}

	for path, rule := range actualRules {
		if desiredRules[path] == nil {
			arPatch[path] = NewDeleteObjectAction(rule)
		}
	}

	objectsToApply := make([]*ObjToPatch, 0, len(arPatch))

	for _, objectToApply := range arPatch {
		objectsToApply = append(objectsToApply, objectToApply)
	}

	return objectsToApply
}

func (a *AccessRuleProcessor) GetDesiredObject(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := hasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if isSecured(rule) {
			ar := generateAccessRule(api, rule, rule.AccessStrategies, a.additionalLabels, a.defaultDomainName)
			accessRules[setAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}

func (a *AccessRuleProcessor) GetActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (map[string]*rulev1alpha1.Rule, error) {
	labels := getOwnerLabels(api)

	var arList rulev1alpha1.RuleList
	if err := a.client.List(ctx, &arList, client.MatchingLabels(labels)); err != nil {
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

func modifyAccessRule(existing, required *rulev1alpha1.Rule) {
	existing.Spec = required.Spec
}

func generateAccessRule(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, accessStrategies []*gatewayv1beta1.Authenticator, additionalLabels map[string]string, defaultDomainName string) *rulev1alpha1.Rule {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := generateOwnerRef(api)

	arBuilder := builders.AccessRule().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies, defaultDomainName))).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

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
