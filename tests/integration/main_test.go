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

	suite.Run()
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
