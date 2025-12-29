package cleaner

import (
	"context"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"
	"github.com/kyma-project/api-gateway/internal/subresources/authorizationpolicy"
	"github.com/kyma-project/api-gateway/internal/subresources/requestauthentication"
	"github.com/kyma-project/api-gateway/internal/subresources/virtualservice"
)

// DeleteAPIRuleSubresources deletes all subresources (AuthorizationPolicies, RequestAuthentications,
// VirtualServices, and AccessRules) that are owned by the given APIRule
func DeleteAPIRuleSubresources(k8sClient client.Client, ctx context.Context, apiRule processing.Labeler) error {

	// Delete AuthorizationPolicies
	apRepo := authorizationpolicy.NewRepository(k8sClient)
	var apCRD apiextensionsv1.CustomResourceDefinition
	err := k8sClient.Get(ctx, client.ObjectKey{Name: "authorizationpolicies.security.istio.io"}, &apCRD)
	if err == nil {
		if err := apRepo.DeleteAll(ctx, apiRule); err != nil {
			return err
		}
	} else if client.IgnoreNotFound(err) != nil {
		return err
	}

	// Delete RequestAuthentications
	raRepo := requestauthentication.NewRepository(k8sClient)
	var raCRD apiextensionsv1.CustomResourceDefinition
	err = k8sClient.Get(ctx, client.ObjectKey{Name: "requestauthentications.security.istio.io"}, &raCRD)
	if err == nil {
		if err := raRepo.DeleteAll(ctx, apiRule); err != nil {
			return err
		}
	} else if client.IgnoreNotFound(err) != nil {
		return err
	}

	// Delete VirtualServices
	vsRepo := virtualservice.NewRepository(k8sClient)
	var vsCRD apiextensionsv1.CustomResourceDefinition
	err = k8sClient.Get(ctx, client.ObjectKey{Name: "virtualservices.networking.istio.io"}, &vsCRD)
	if err == nil {
		if err := vsRepo.DeleteAll(ctx, apiRule); err != nil {
			return err
		}
	} else if client.IgnoreNotFound(err) != nil {
		return err
	}

	// Delete AccessRules (Ory Rules) if CRD exists
	arRepo := accessrule.NewRepository(k8sClient)
	var oryCRD apiextensionsv1.CustomResourceDefinition
	err = k8sClient.Get(ctx, client.ObjectKey{Name: "rules.oathkeeper.ory.sh"}, &oryCRD)
	if err == nil {
		if err := arRepo.DeleteAll(ctx, apiRule); err != nil {
			return err
		}
	} else if client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}
