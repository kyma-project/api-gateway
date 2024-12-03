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

	// do it by a namespace
	rateLimitList := v1alpha1.RateLimitList{}
	err = k8sClient.List(ctx, &rateLimitList)
	if err != nil {
		return err
	}

	if !isIngressGateway(podList.Items) {
		err = validateSidecarInjectionEnabled(podList.Items)
		if err != nil {
			return err
		}
	} else {
		if !isRlInIngressGatewayNamespace(rateLimitList.Items) {
			return fmt.Errorf("rateLimit CR is matching istio ingress gateway pod but it is not in the istio-system namespace")
		}
	}

	err = validateConflicts(rl, podList.Items, rateLimitList.Items)
	if err != nil {
		return err
	}

	return nil

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
}

func validateSidecarInjectionEnabled(podList []v1.Pod) error {
	for _, pod := range podList {
		_, ok := pod.Annotations["sidecar.istio.io/status"]
		if !ok {
			return fmt.Errorf("sidecar injection is not enabled for the pod: %s in namespace: %s", pod.Name, pod.Namespace)
		}
	}
	return nil
}

func isRlInIngressGatewayNamespace(rateLimits []v1alpha1.RateLimit) bool {
	for _, rl := range rateLimits {
		if rl.Namespace != "istio-system" {
			return false
		}
	}
	return true
}

func isIngressGateway(pods []v1.Pod) bool {
	for _, p := range pods {
		v, ok := p.Labels["istio"]
		if !ok || v != "ingressgateway" {
			return false
		}
	}
	return true
}

func validateConflicts(rl v1alpha1.RateLimit, podList []v1.Pod, rateLimitList []v1alpha1.RateLimit) error {
	// if there is only one RateLimit CR then there won't be any conflicts
	// we take pods that are matching the selectors from the RateLimit CR
	// then we check if any other RateLimit is matching any other selector on the pods
	// if yes then we return an error
	// if no then we return nil
	if len(rateLimitList) < 2 {
		// we return nil here because if there is no other RateLimit CRs then there won't be any conflicting ones
		return nil
	}

	// iterate over all rate limits
	// get all pods that are matching the selectors from the RateLimit CR (append map with the names of all pods)
	// check if pod names are unique

	for _, pod := range podList {
		for _, rateLimit := range rateLimitList {
			// we want to skip the RateLimit CR that we are currently validating
			if rateLimit.Name == rl.Name {
				continue
			}
			for key, value := range rateLimit.Spec.SelectorLabels {
				if v, ok := pod.Labels[key]; ok && v == value {
					return fmt.Errorf("RateLimit CR %s already matches pod: %s in namespace: %s", rateLimit.Name, pod.Name, pod.Namespace)
				}
			}
		}
	}
	return nil
}
