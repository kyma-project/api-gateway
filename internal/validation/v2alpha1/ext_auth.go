package v2alpha1

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getIstioConfigMap(ctx context.Context, client client.Client) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := client.Get(ctx, types.NamespacedName{Name: "istio", Namespace: "istio-system"}, configMap)
	return configMap, err
}

type meshData struct {
	ExtensionProviders []struct {
		Name              string `yaml:"name"`
		EnvoyExtAuthzHttp any    `yaml:"envoyExtAuthzHttp"`
	} `yaml:"extensionProviders"`
}

func validateExtAuthProviders(ctx context.Context, k8sClient client.Client, parentAttributePath string,
	rule gatewayv2alpha1.Rule) (problems []validation.Failure, err error) {
	istioConfigMap, err := getIstioConfigMap(ctx, k8sClient)
	if err != nil {
		return []validation.Failure{
			{
				AttributePath: parentAttributePath + ".extAuth",
				Message:       "Failed to get Istio ConfigMap",
			},
		}, nil
	}

	data, ok := istioConfigMap.Data["mesh"]
	if !ok {
		problems = append(problems, validation.Failure{
			AttributePath: parentAttributePath + ".extAuth.externalAuthorizers",
			Message:       "Istio ConfigMap does not contain mesh key",
		})

		// Since all following validation would require the mesh key, we can return early
		return problems, nil
	}

	var mesh meshData
	if err := yaml.Unmarshal([]byte(data), &mesh); err != nil {
		problems = append(problems, validation.Failure{
			AttributePath: parentAttributePath + ".extAuth.externalAuthorizers",
			Message:       "Failed to unmarshal mesh data",
		})
		// Since all following validation would require the mesh data, we can return early
		return problems, nil
	}

	for _, authorizer := range rule.ExtAuth.ExternalAuthorizers {
		p := validateAuthorizer(authorizer, mesh, parentAttributePath)
		problems = append(problems, p...)
	}

	return problems, nil
}

func validateAuthorizer(authorizer string, mesh meshData, parentAttributePath string) []validation.Failure {
	found := false
	for _, provider := range mesh.ExtensionProviders {
		if provider.Name == authorizer {
			if provider.EnvoyExtAuthzHttp == nil {
				return []validation.Failure{
					{
						AttributePath: parentAttributePath + ".extAuth.externalAuthorizers." + authorizer,
						Message:       "EnvoyExtAuthzHttp not found in Istio ConfigMap mesh data for authorizer",
					},
				}
			}
			found = true
			break
		}
	}
	if !found {
		return []validation.Failure{
			{
				AttributePath: parentAttributePath + ".extAuth.externalAuthorizers." + authorizer,
				Message:       "Authorizer not found in Istio ConfigMap mesh data",
			},
		}
	}

	return []validation.Failure{}
}
