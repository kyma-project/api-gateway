package validation

import (
	"context"
	"fmt"

	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const istioSidecarContainerName = "istio-proxy"

type InjectionValidator struct {
	Ctx    context.Context
	Client client.Client
}

func NewInjectionValidator(ctx context.Context, client client.Client) *InjectionValidator {
	return &InjectionValidator{Ctx: ctx, Client: client}
}

func (v *InjectionValidator) Validate(attributePath string, selector *apiv1beta1.WorkloadSelector, workloadNamespace string) (problems []Failure, err error) {
	if selector == nil {
		problems = append(problems, Failure{
			AttributePath: attributePath + ".injection",
			Message:       "Target service label selectors are not defined",
		})

		return problems, nil
	}

	var podList corev1.PodList
	err = v.Client.List(v.Ctx, &podList, client.InNamespace(workloadNamespace), client.MatchingLabels(selector.MatchLabels))
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if !containsSidecar(pod) {
			problems = append(problems, Failure{AttributePath: attributePath, Message: fmt.Sprintf("Pod %s/%s does not have an injected istio sidecar", pod.Namespace, pod.Name)})
		}
	}
	return problems, nil
}

func containsSidecar(pod corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == istioSidecarContainerName {
			return true
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if container.Name == istioSidecarContainerName {
			return true
		}
	}
	return false
}
