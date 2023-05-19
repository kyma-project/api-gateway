package api_gateway

import (
	"context"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/custom-domain"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/istio-jwt"
	"log"
	"os"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	config := testcontext.GetConfig()
	ts := testcontext.NewTestSuite("istio-jwt", config)
	orgJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(ts, orgJwtHandler)

	opts := createGoDogOpts(t, "testsuites/istio-jwt/features/", config.TestConcurrency)
	suite := godog.TestSuite{
		Name: ts.Name,
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			istiojwt.Init(ctx.ScenarioContext(), ts)
		},
		Options: &opts,
	}

	testExitCode := suite.Run()
	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func TestCustomDomain(t *testing.T) {
	config := testcontext.GetConfig()
	ts := testcontext.NewTestSuite("custom-domain", config)
	defer ts.TearDownCommonResources()
	opts := createGoDogOpts(t, "features/custom-domain/custom_domain.feature", config.TestConcurrency)

	customDomainSuite := godog.TestSuite{
		Name: ts.Name,
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			customdomain.Init(ctx.ScenarioContext(), ts)
		},
		Options: &opts,
	}

	testExitCode := customDomainSuite.Run()

	//Remove certificate
	res := schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	err := ts.K8sClient.Resource(res).Namespace("istio-system").DeleteCollection(context.TODO(), v1.DeleteOptions{}, v1.ListOptions{LabelSelector: "owner=custom-domain-test"})
	if err != nil {
		log.Print(err.Error())
	}

	if os.Getenv(testcontext.ExportResultVar) == "true" {
		generateReport()
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func createGoDogOpts(t *testing.T, featuresPath string, concurrency int) godog.Options {
	goDogOpts := godog.Options{
		Output:      colors.Colored(os.Stdout),
		Format:      "pretty",
		Paths:       []string{featuresPath},
		Concurrency: concurrency,
		TestingT:    t,
	}

	if os.Getenv(testcontext.ExportResultVar) == "true" {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(ts testcontext.Testsuite, orgJwtHandler string) {

	ts.TearDownCommonResources()

	if os.Getenv(testcontext.ExportResultVar) == "true" {
		generateReport()
	}

	_, err := SwitchJwtHandler(ts, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}
