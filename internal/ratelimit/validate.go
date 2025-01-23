package ratelimit

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// Validate checks the validity of the given RateLimit custom resource.
// It ensures that there are pods matching the specified selector labels in the same namespace.
// If no matching pods are found, or if the sidecar injection is not enabled for the pods,
// or if there are conflicting RateLimit resources, an error is returned.
//
// Parameters:
// - ctx: The context for the validation operation.
// - k8sClient: The Kubernetes client used to interact with the cluster.
// - rl: The RateLimit custom resource to validate.
//
// Returns:
// - An error if the validation fails, otherwise nil.
func Validate(ctx context.Context, k8sClient client.Client, rl v1alpha1.RateLimit) error {
	selectors := rl.Spec.SelectorLabels

	if err := validateIntervals(rl); err != nil {
		return err
	}

	matchingPods := v1.PodList{}
	err := k8sClient.List(ctx, &matchingPods, client.InNamespace(rl.Namespace), client.MatchingLabels(selectors))
	if err != nil {
		return err
	}

	if len(matchingPods.Items) == 0 {
		// in case there is no pods matching for the given selectors declared in the RateLimit CR
		// we want to set the RateLimit CR to the warning state, therefore we fail validation returning an error
		return fmt.Errorf("no pods found with the given selectors: %v in namespace %s", selectors, rl.Namespace)
	}

	if !isIngressGateway(matchingPods.Items) {
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

func validateIntervals(rl v1alpha1.RateLimit) error {
	if rl.Spec.Local.DefaultBucket.FillInterval == nil {
		// API validation ensures that this field is not empty
		return nil
	}
	globalFillInterval := rl.Spec.Local.DefaultBucket.FillInterval.Duration
	if globalFillInterval < 50*time.Millisecond {
		return fmt.Errorf("defaultBucket: fillInterval must be greater or equal 50ms")
	}

	for i, b := range rl.Spec.Local.Buckets {
		if b.Bucket.FillInterval == nil {
			// API validation ensures that this field is not empty
			return nil
		}
		bucketFillInterval := b.Bucket.FillInterval.Duration
		if bucketFillInterval < 50*time.Millisecond {
			return fmt.Errorf("bucket '[%d]': fillInterval must be greater or equal 50ms", i)
		}
		if bucketFillInterval%globalFillInterval != 0 {
			return fmt.Errorf("bucket '[%d]': fillInterval must be a multiple of defaultBucket fillInterval", i)
		}
	}
	return nil
}

func validateConflicts(ctx context.Context, k8sClient client.Client, rl v1alpha1.RateLimit, matchingPods v1.PodList) error {
	otherRateLimitsInTheNamespace := v1alpha1.RateLimitList{}
	err := k8sClient.List(ctx, &otherRateLimitsInTheNamespace, client.InNamespace(rl.Namespace))
	if err != nil {
		return err
	}

	podMap := map[string]v1.Pod{}
	var conflictingRateLimits []v1alpha1.RateLimit

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
				conflictingRateLimits = append(conflictingRateLimits, otherRL)
				break
			}
		}
	}

	if len(conflictingRateLimits) > 0 {
		return conflictError(conflictingRateLimits)
	}
	return nil
}

func validateSidecarInjectionEnabled(podList []v1.Pod) error {
	var nonInjectedPods []v1.Pod
	for _, pod := range podList {
		_, ok := pod.Annotations["sidecar.istio.io/status"]
		if !ok {
			nonInjectedPods = append(nonInjectedPods, pod)
		}
	}

	if len(nonInjectedPods) > 0 {
		return sidecarInjectionError(nonInjectedPods)
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
}

func sidecarInjectionError(nonCompliantPods []v1.Pod) error {
	var invalidPodsMessages []string
	for _, pod := range nonCompliantPods {
		msg := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		invalidPodsMessages = append(invalidPodsMessages, msg)
	}
	return fmt.Errorf("sidecar injection is not enabled for the following pods: %s", strings.Join(invalidPodsMessages, ", "))
}

func conflictError(conflictingRateLimits []v1alpha1.RateLimit) error {
	var conflictingRateLimitsMessages []string
	for _, rl := range conflictingRateLimits {
		conflictingRateLimitsMessages = append(conflictingRateLimitsMessages, rl.Name)
	}
	return fmt.Errorf("conflicting with the following RateLimit CRs: %s", strings.Join(conflictingRateLimitsMessages, ", "))
}
