package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Validate(ctx context.Context, k8sClient client.Client, rl v1alpha1.RateLimit) error {
	selectors := rl.Spec.SelectorLabels

	matchingPods := v1.PodList{}
	err := k8sClient.List(ctx, &matchingPods, client.InNamespace(rl.Namespace), client.MatchingLabels(selectors))
	if err != nil {
		return err
	}

	if len(matchingPods.Items) == 0 {
		// in case there is no pods matching for the given selectors declared in the RateLimit CR
		// we want to set the RateLimit CR to the warning state, therefore we fail validation returning an error
		return errors.New("no pods found with the given selectors")
	}

	if isIngressGateway(matchingPods.Items) {
		if rl.Namespace != "istio-system" {
			return fmt.Errorf("rateLimit CR is matching istio ingress gateway pod but it is not in the istio-system namespace")
		}
	} else {
		err = validateSidecarInjectionEnabled(matchingPods.Items)
		if err != nil {
			return err
		}
	}

	err = validateConflicts(ctx, k8sClient, rl, matchingPods)
	if err != nil {
		return err
	}

	return nil
}

func validateSidecarInjectionEnabled(podList []v1.Pod) error {
	var nonCompliantPods []v1.Pod
	for _, pod := range podList {
		_, ok := pod.Annotations["sidecar.istio.io/status"]
		if !ok {
			nonCompliantPods = append(nonCompliantPods, pod)
		}
	}

	if len(nonCompliantPods) > 0 {
		return sidecarInjectionError(nonCompliantPods)
	}

	return nil
}

func isIngressGateway(pods []v1.Pod) bool {
	for _, p := range pods {
		v, ok := p.Labels["istio"]
		if !ok || v != "ingressgateway" {
			return false
		}
	}
	return true
	//
}

func validateConflicts(ctx context.Context, k8sClient client.Client, rl v1alpha1.RateLimit, matchingPods v1.PodList) error {
	otherRateLimitsInTheNamespace := v1alpha1.RateLimitList{}
	err := k8sClient.List(ctx, &otherRateLimitsInTheNamespace, client.InNamespace(rl.Namespace))
	if err != nil {
		return err
	}

	podMap := map[string]v1.Pod{}

	for _, pod := range matchingPods.Items {
		podMap[pod.Name] = pod
	}

	for _, otherRL := range otherRateLimitsInTheNamespace.Items {
		if otherRL.Name == rl.Name {
			continue
		}

		otherRLSelectors := otherRL.Spec.SelectorLabels
		otherRLMatchingPods := v1.PodList{}
		err := k8sClient.List(ctx, &otherRLMatchingPods, client.InNamespace(rl.Namespace), client.MatchingLabels(otherRLSelectors))
		if err != nil {
			return err
		}

		for _, pod := range otherRLMatchingPods.Items {
			if _, ok := podMap[pod.Name]; ok {
				return fmt.Errorf("conflict detected between RateLimit CRs: %s and %s", rl.Name, otherRL.Name)
			}
		}
	}

	return nil
}

func sidecarInjectionError(nonCompliantPods []v1.Pod) error {
	var invalidPodsMessages []string
	for _, pod := range nonCompliantPods {
		msg := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		invalidPodsMessages = append(invalidPodsMessages, msg)
	}
	return fmt.Errorf("sidecar injection is not enabled for the following pods: %s", strings.Join(invalidPodsMessages, ", "))
}

// check if any workload exists with the given selector
// then we take all the selectors from the workload
// then we check if for those other selectors other ratelimits exists (if any RateLimit CR has a selector that is on the workload
// if yes then download other selectors, if no warning state something like there is no workload with the given selector
// check if for other selectors other

// First RateLimit CR: selectorLabels: app=boo
// Second RateLimit CR: selectorLabels: app=boo app=poo -> got into the warning state because EF is already applied to both workloads

// 1 pod: boo, poo -> First RL CR
// 2 pod: boo -> First RL CR

// Read selectors from RateLimit CR
// Check if any workload exists with the given selector
// Take all the workloads with the given selectors -> might have zoo + check if the workload is ingress-gateway or has sidecar injected
// Check if any other already existing RateLimit CRs applies to those workloads
