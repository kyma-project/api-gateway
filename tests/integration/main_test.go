package api_gateway

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	InitTestSuite()

	orgJwtHandler, err := SwitchJwtHandler("istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(orgJwtHandler)

	SetupCommonResources("istio-jwt")

	opts := goDogOpts
	opts.Paths = []string{"features/istio-jwt/"}
	opts.Concurrency = conf.TestConcurrency

	suite := godog.TestSuite{
		Name: "istio-jwt",
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			initIstioJwtScenarios(ctx.ScenarioContext())
		},
		Options: &opts,
	}

	testExitCode := suite.Run()
	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func TestCustomDomain(t *testing.T) {
	InitTestSuite()
	SetupCommonResources("custom-domain")

	customDomainOpts := goDogOpts
	customDomainOpts.Paths = []string{"features/custom-domain/custom_domain.feature"}
	customDomainOpts.Concurrency = conf.TestConcurrency
	if os.Getenv(exportResultVar) == "true" {
		customDomainOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}
	customDomainSuite := godog.TestSuite{
		Name: "custom-domain",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeScenarioCustomDomain(ctx.ScenarioContext())
		},
		Options: &customDomainOpts,
	}

	testExitCode := customDomainSuite.Run()

	//Remove namespace
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := k8sClient.Resource(res).Delete(context.Background(), namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	//Remove certificate
	res = schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	err = k8sClient.Resource(res).Namespace("istio-system").DeleteCollection(context.TODO(), v1.DeleteOptions{}, v1.ListOptions{LabelSelector: "owner=custom-domain-test"})
	if err != nil {
		log.Print(err.Error())
	}

	if os.Getenv(exportResultVar) == "true" {
		generateReport()
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func cleanUp(orgJwtHandler string) {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := k8sClient.Resource(res).Delete(context.Background(), namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	err = k8sClient.Resource(res).Delete(context.Background(), secondNamespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	if os.Getenv(exportResultVar) == "true" {
		generateReport()
	}

	_, err = SwitchJwtHandler(orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch back to original jwtHandler")
	}
}
