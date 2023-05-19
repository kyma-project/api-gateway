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
	testCtx := testcontext.New("istio-jwt", config)
	orgJwtHandler, err := SwitchJwtHandler(testCtx, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(testCtx, orgJwtHandler)

	opts := createGoDogOpts(t, "testsuites/istio-jwt/features/", config.TestConcurrency)
	suite := godog.TestSuite{
		Name: testCtx.Name,
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			istiojwt.Init(ctx.ScenarioContext(), &testCtx)
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
	testCtx := testcontext.New("custom-domain", config)
	defer testCtx.TearDownCommonResources()
	opts := createGoDogOpts(t, "testsuites/custom-domain/features/", config.TestConcurrency)

	customDomainSuite := godog.TestSuite{
		Name: testCtx.Name,
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			customdomain.Init(ctx.ScenarioContext(), &testCtx)
		},
		Options: &opts,
	}

	testExitCode := customDomainSuite.Run()

	//Remove certificate
	res := schema.GroupVersionResource{Group: "cert.gardener.cloud", Version: "v1alpha1", Resource: "certificates"}
	err := testCtx.K8sClient.Resource(res).Namespace("istio-system").DeleteCollection(context.TODO(), v1.DeleteOptions{}, v1.ListOptions{LabelSelector: "owner=custom-domain-test"})
	if err != nil {
		log.Print(err.Error())
	}

	if shouldExportResults() {
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

	if shouldExportResults() {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(c testcontext.Context, orgJwtHandler string) {

	c.TearDownCommonResources()

	if shouldExportResults() {
		generateReport()
	}

	_, err := SwitchJwtHandler(c, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}

func shouldExportResults() bool {
	return os.Getenv("EXPORT_RESULT") == "true"
}
