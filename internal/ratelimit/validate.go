package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Validate(ctx context.Context, k8sClient client.Client, rl v1alpha1.RateLimit) error {
	selectors := rl.Spec.SelectorLabels

	podList := v1.PodList{}
	err := k8sClient.List(ctx, &podList, client.MatchingLabels(selectors))
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		// in case there is no pods matching for the given selectors declared in the RateLimit CR
		// we want to set the RateLimit CR to the warning state, therefore we fail validation returning an error
		return errors.New("no pods found with the given selectors")
	}

	rateLimitList := v1alpha1.RateLimitList{}
	err = k8sClient.List(ctx, &rateLimitList)
	if err != nil {
		return err
	}

	// validate proxy presence
	err = validateSidecarPresence(podList.Items)
	if err != nil {
		return err
	}

	// validate conflicts
	err = validateConflicts(rl, podList.Items, rateLimitList.Items)
	if err != nil {
		return err
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
	return nil
}

func validateSidecarPresence(podList []v1.Pod) error {
	for i, pod := range podList {
		isSidecarPresent := false
		for _, c := range pod.Spec.Containers {
			if c.Name == "istio-proxy" {
				isSidecarPresent = true
			}
		}

		if i == len(podList)-1 && !isSidecarPresent {
			return errors.New("no sidecar found in the pod selected by the RateLimit CR")
		}
	}
	return nil
}

func validateConflicts(rl v1alpha1.RateLimit, podList []v1.Pod, rateLimitList []v1alpha1.RateLimit) error {
	if len(rateLimitList) < 2 {
		// we return nil here because if there is no other RateLimit CRs then there won't be any conflicting ones
		return nil
	}
	for _, pod := range podList {
		for _, rateLimit := range rateLimitList {
			// we want to skip the RateLimit CR that we are currently validating
			if rateLimit.Name == rl.Name {
				continue
			}

			for key, value := range rateLimit.Spec.SelectorLabels {
				if pod.Labels[key] == value {
					return fmt.Errorf("RateLimit CR already exists for the given pod: %s in namespace: %s", pod.Name, pod.Namespace)
				}
			}

		}
	}
	return nil
}
