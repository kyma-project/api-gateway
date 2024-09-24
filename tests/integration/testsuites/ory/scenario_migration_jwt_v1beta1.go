package ory

import (
	"context"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func initMigrationJwtV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-jwt-v1beta1.yaml", "migration-jwt-v1beta1")

	ctx.Step(`^migrationJwtV1beta1: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^migrationJwtV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationJwtV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationJwtV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^migrationJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination$`, scenario.thereIsApiRuleVirtualServiceWithHttpbinServiceDestination)
	ctx.Step(`^migrationJwtV1beta1: Resource of Kind "([^"]*)" owned by APIRule does not exist$`, scenario.resourceOwnedByApiRuleDoesNotExist)
	ctx.Step(`^migrationJwtV1beta1: Resource of Kind "([^"]*)" owned by APIRule exists$`, scenario.resourceOwnedByApiRuleExists)
	ctx.Step(`^migrationJwtV1beta1: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
}

func (s *scenario) thereIsApiRuleVirtualServiceWithHttpbinServiceDestination() error {
	res := resource.GetResourceGvr("VirtualService")
	ownerLabelSelector := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s-%s.%s", s.name, s.TestID, s.Namespace)

	return retry.Do(func() error {
		vsList, err := s.k8sClient.Resource(res).Namespace(s.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: ownerLabelSelector})
		if err != nil {
			return err
		}

		if len(vsList.Items) != 1 {
			return fmt.Errorf("expected 1 VirtualService for APIRule, got %d", len(vsList.Items))
		}

		vs := networkingv1alpha3.VirtualService{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(vsList.Items[0].Object, &vs)
		if err != nil {
			return err
		}

		if len(vs.Spec.Http) != 1 {
			return fmt.Errorf("expected 1 HTTP route, got %d", len(vs.Spec.Http))
		}
		if len(vs.Spec.Http[0].Route) != 1 {
			return fmt.Errorf("expected 1 route, got %d", len(vs.Spec.Http[0].Route))
		}

		expectedDestination := fmt.Sprintf("httpbin-%s.%s.svc.cluster.local", s.TestID, s.Namespace)
		if vs.Spec.Http[0].Route[0].Destination.Host != expectedDestination {
			return fmt.Errorf("expected destination host to be %s, got %s", expectedDestination, vs.Spec.Http[0].Route[0].Destination.Host)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) resourceOwnedByApiRuleDoesNotExist(resourceKind string) error {
	res := resource.GetResourceGvr(resourceKind)
	ownerLabelSelector := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s-%s.%s", s.name, s.TestID, s.Namespace)
	return retry.Do(func() error {
		list, err := s.k8sClient.Resource(res).Namespace(s.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: ownerLabelSelector})
		if err != nil {
			return err
		}

		if len(list.Items) > 0 {
			return fmt.Errorf("expected at least one %s owned by APIRule, got %d", resourceKind, len(list.Items))
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) resourceOwnedByApiRuleExists(resourceKind string) error {
	res := resource.GetResourceGvr(resourceKind)
	ownerLabelSelector := fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s-%s.%s", s.name, s.TestID, s.Namespace)
	return retry.Do(func() error {
		list, err := s.k8sClient.Resource(res).Namespace(s.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: ownerLabelSelector})
		if err != nil {
			return err
		}

		if len(list.Items) == 0 {
			return fmt.Errorf("expected at least one %s owned by APIRule, got 0", resourceKind)
		}

		return nil
	}, testcontext.GetRetryOpts()...)
}

func (s *scenario) thereIsAHttpbinServiceWithIstioInjection() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-app-istio-injection.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.Url = fmt.Sprintf("https://httpbin-%s.%s", s.TestID, s.Domain)

	return nil
}
