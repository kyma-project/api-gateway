package api_gateway

import (
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/custom-domain"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/gateway"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/istio-jwt"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/ory"
	"github.com/kyma-project/api-gateway/tests/integration/testsuites/upgrade"
	"log"
	"os"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestIstioJwt(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, istiojwt.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Istio JWT testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	runTestsuite(t, ts, config)
}

func TestCustomDomain(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, customdomain.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Custom domain testsuite %s", err.Error())
	}
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func TestUpgrade(t *testing.T) {
	config := testcontext.GetConfig()
	config.TestConcurrency = 1
	ts, err := testcontext.New(config, upgrade.NewTestsuite)

	if err != nil {
		t.Fatalf("Failed to create Upgrade testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func TestOryJwt(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, ory.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Ory testsuite %s", err.Error())
	}
	originalJwtHandler, err := SwitchJwtHandler(ts, "ory")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Ory jwtHandler")
	}
	defer cleanUp(ts, originalJwtHandler)
	runTestsuite(t, ts, config)
}

func TestGateway(t *testing.T) {
	config := testcontext.GetConfig()
	ts, err := testcontext.New(config, gateway.NewTestsuite)
	if err != nil {
		t.Fatalf("Failed to create Gateway testsuite %s", err.Error())
	}
	defer ts.TearDown()
	runTestsuite(t, ts, config)
}

func runTestsuite(t *testing.T, testsuite testcontext.Testsuite, config testcontext.Config) {
	opts := createGoDogOpts(t, testsuite.FeaturePath(), config.TestConcurrency)
	suite := godog.TestSuite{
		Name: testsuite.Name(),
		// We are not using ScenarioInitializer, as this function only needs to set up global resources
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.BeforeSuite(func() {
				if err := createApiGatewayCR(config); err != nil {
					t.Fatalf("Cannot create api-gateway CR: %s", err.Error())
				}
			})

			testsuite.InitScenarios(ctx.ScenarioContext())

			ctx.AfterSuite(func() {
				if err := deleteApiGatewayCR(config); err != nil {
					t.Fatalf("Cannot delete api-gateway CR: %s", err.Error())
				}
			})
		},
		Options: &opts,
	}

	testExitCode := suite.Run()

	if shouldExportResults() {
		generateReport(testsuite)
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
		Strict:      true,
	}

	if shouldExportResults() {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	return goDogOpts
}

func cleanUp(c testcontext.Testsuite, orgJwtHandler string) {

	c.TearDown()

	_, err := SwitchJwtHandler(c, orgJwtHandler)
	if err != nil {
		log.Print(err.Error())
		panic("unable to switch back to original jwtHandler")
	}
}

func shouldExportResults() bool {
	return os.Getenv("EXPORT_RESULT") == "true"
}

func createApiGatewayCR(config testcontext.Config) error {
	apiGatewayCR, err := manifestprocessor.ParseFromFileWithTemplate("api-gateway.yaml", "manifests/", struct {
		NamePrefix string
	}{NamePrefix: resource.TestGatewayOperatorName})
	if err != nil {
		log.Fatalf("failed to process api-gateway manifest file, details %v", err)
		return err
	}

	k8sClient, err := client.GetDynamicClient()
	if err != nil {
		return err
	}

	rm := resource.NewManager(testcontext.GetRetryOpts(config))
	_, err = rm.CreateGateway(k8sClient, apiGatewayCR...)
	if err != nil {
		return err
	}

	return nil
}

func deleteApiGatewayCR(config testcontext.Config) error {
	apiGatewayCR, err := manifestprocessor.ParseFromFileWithTemplate("api-gateway.yaml", "manifests/", struct {
		NamePrefix string
	}{NamePrefix: resource.TestGatewayOperatorName})
	if err != nil {
		log.Fatalf("failed to process api-gateway manifest file, details %v", err)
		return err
	}

	k8sClient, err := client.GetDynamicClient()
	if err != nil {
		return err
	}

	rm := resource.NewManager(testcontext.GetRetryOpts(config))
	err = rm.DeleteResourcesWithoutNS(k8sClient, apiGatewayCR...)
	if err != nil {
		return err
	}

	return nil
}
