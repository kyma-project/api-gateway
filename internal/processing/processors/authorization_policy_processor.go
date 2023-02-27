package processors

import (
	"context"
	"strconv"
	"strings"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const AuthorizationPolicyAppSelectorLabel = "app"

// AuthorizationPolicyProcessor is the generic processor that handles the Istio JwtAuthorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule) (map[string][]*securityv1beta1.AuthorizationPolicy, error)
}

func (r AuthorizationPolicyProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(apiRule)
	if err != nil {
		return nil, err
	}
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r AuthorizationPolicyProcessor) getDesiredState(api *gatewayv1beta1.APIRule) (map[string][]*securityv1beta1.AuthorizationPolicy, error) {
	aps, err := r.Creator.Create(api)
	if err != nil {
		return nil, err
	}
	return aps, nil
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]map[string]*securityv1beta1.AuthorizationPolicy, error) {
	labels := processing.GetOwnerLabels(api)

	var apList securityv1beta1.AuthorizationPolicyList
	if err := client.List(ctx, &apList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	authorizationPolicies := make(map[string]map[string]*securityv1beta1.AuthorizationPolicy)
	for i, ap := range apList.Items {
		index, ok := ap.Labels[processing.IndexLabelName]
		if !ok {
			// Store unindexed APs so they will be deleted/updated with index
			index = "-" + strconv.Itoa(i)
		}

		if hash, ok := ap.Labels[processing.HashLabelName]; ok {
			if authorizationPolicies[hash] == nil {
				authorizationPolicies[hash] = make(map[string]*securityv1beta1.AuthorizationPolicy)
			}

			if _, ok := authorizationPolicies[hash][index]; ok {
				authorizationPolicies[hash][index+":"+strconv.Itoa(i)] = ap
			} else {
				authorizationPolicies[hash][index] = ap
			}
		} else {
			hashTo, err := helpers.GetAuthorizationPolicyHash(*ap)
			if err != nil {
				return nil, err
			}

			if authorizationPolicies[hashTo] == nil {
				authorizationPolicies[hashTo] = make(map[string]*securityv1beta1.AuthorizationPolicy)
			}

			if _, ok := authorizationPolicies[hash][index]; ok {
				authorizationPolicies[hash][index+":"+strconv.Itoa(i)] = ap
			} else {
				authorizationPolicies[hash][index] = ap
			}
		}
	}

	return authorizationPolicies, nil
}

func (r AuthorizationPolicyProcessor) getObjectChanges(desiredAps map[string][]*securityv1beta1.AuthorizationPolicy, actualAps map[string]map[string]*securityv1beta1.AuthorizationPolicy) []*processing.ObjectChange {
	var apObjectActionsToApply []*processing.ObjectChange

	for hashTo, toDesiredAPs := range desiredAps {
		for _, ap := range toDesiredAPs {
			// As both the order of Authorizations and APs is static we can update them according to array index
			index := ap.Labels[processing.IndexLabelName]
			if oldAP, ok := actualAps[hashTo][index]; ok {
				oldAP.Spec = ap.Spec
				oldAP.Labels = ap.Labels
				delete(actualAps[hashTo], index)
				apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectUpdateAction(oldAP))
			} else if len(actualAps[hashTo]) > 0 {
				found := false

				for key, oldAP := range actualAps[hashTo] {
					found = strings.HasPrefix(key, "-")
					// APs without already assigned index have negative key
					if found {
						oldAP.Spec = ap.Spec
						oldAP.Labels = ap.Labels
						delete(actualAps[hashTo], key)
						apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectUpdateAction(oldAP))
						break
					}
				}

				if !found {
					apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectCreateAction(ap))
				}
			} else {
				apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectCreateAction(ap))
			}
		}
	}

	for _, aps := range actualAps {
		for _, ap := range aps {
			apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectDeleteAction(ap))
		}
	}

	return apObjectActionsToApply
}
