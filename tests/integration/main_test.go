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
	opts := goDogOpts
	opts.Paths = []string{"features/istio-jwt/istio_jwt.feature"}
	opts.Concurrency = conf.TestConcurency

	suite := godog.TestSuite{
		Name: "istio-jwt",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeScenarioIstioJWT(ctx.ScenarioContext())
		},
		Options: &opts,
	}

	runIstioJwtTest(t, suite)
}

func TestIstioJwtScopes(t *testing.T) {
	opts := goDogOpts
	opts.Paths = []string{"features/istio-jwt/istio_jwt_scopes.feature"}
	opts.Concurrency = conf.TestConcurency

	suite := godog.TestSuite{
		Name: "istio-jwt-scopes",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeScenarioIstioJWTScopes(ctx.ScenarioContext())
		},
		Options: &opts,
	}

	runIstioJwtTest(t, suite)
}

func runIstioJwtTest(t *testing.T, suite godog.TestSuite) {
	InitTestSuite()

	orgJwtHandler, err := SwitchJwtHandler("istio")
	if err != nil {
		log.Print(err.Error())
		t.Fatalf("unable to switch to Istio jwtHandler")
	}
	defer cleanUp(orgJwtHandler)

	SetupCommonResources("istio-jwt")

	testExitCode := suite.Run()
	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests, Pod list: %s\n APIRules: %s\n", getPodListReport(), getApiRules())
	}
}

func cleanUp(orgJwtHandler string) {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := k8sClient.Resource(res).Delete(context.Background(), namespace, v1.DeleteOptions{})

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
