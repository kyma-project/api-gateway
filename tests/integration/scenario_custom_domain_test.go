package api_gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"path"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

type CustomDomainScenario struct {
	domain         string
	loadBalancerIP net.IP
	testID         string
	namespace      string
	url            string
	apiResourceOne []unstructured.Unstructured
	apiResourceTwo []unstructured.Unstructured
}

func InitializeScenarioCustomDomain(ctx *godog.ScenarioContext) {
	scenario, err := CreateCustomDomainScenario("no-access-strategy.yaml", "oauth-strategy.yaml", "custom-domain", "testing-app.yaml")

	if err != nil {
		log.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}

	ctx.Step(`^there is a "([^"]*)" DNS cloud credentials secret in "([^"]*)" namespace$`, scenario.thereIsAnCloudCredentialsSecret)
	ctx.Step(`^there is an "([^"]*)" service in "([^"]*)" namespace$`, scenario.thereIsAnExposedService)
	ctx.Step(`^create custom domain resources$`, scenario.createResources)
	ctx.Step(`^ensure that DNS record is ready$`, scenario.isDNSReady)
	ctx.Step(`^there is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^calling the "([^"]*)" endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBetween)
	ctx.Step(`^endpoint is secured with OAuth2$`, scenario.secureWithOAuth2)
	ctx.Step(`^calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint with a valid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidTokenShouldResultInStatusBetween)
	ctx.Step(`^calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBetween)
}

func CreateCustomDomainScenario(templateFileNameOne string, templateFileNameTwo string, namePrefix string, deploymentFile string) (*CustomDomainScenario, error) {
	testID := generateRandomString(testIDLength)
	customDomainManifestDirectory := path.Join(manifestsDirectory, customDomainDirectory)

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(deploymentFile, customDomainManifestDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRuleOne, err := manifestprocessor.ParseFromFileWithTemplate(templateFileNameOne, customDomainManifestDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: fmt.Sprintf("%s.%s", testID, conf.CustomDomain), GatewayName: fmt.Sprintf("%s-%s", namePrefix, testID),
		GatewayNamespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}
	accessRuleTwo, err := manifestprocessor.ParseFromFileWithTemplate(templateFileNameTwo, customDomainManifestDirectory, resourceSeparator, struct {
		Namespace          string
		NamePrefix         string
		TestID             string
		Domain             string
		GatewayName        string
		GatewayNamespace   string
		IssuerUrl          string
		EncodedCredentials string
	}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: fmt.Sprintf("%s.%s", testID, conf.CustomDomain), GatewayName: fmt.Sprintf("%s-%s", namePrefix, testID), GatewayNamespace: namespace, IssuerUrl: conf.IssuerUrl, EncodedCredentials: base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", conf.ClientID, conf.ClientSecret)))})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &CustomDomainScenario{domain: conf.CustomDomain, testID: testID, namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s.%s", testID, testID, conf.CustomDomain), apiResourceOne: accessRuleOne, apiResourceTwo: accessRuleTwo}, nil
}

func (c *CustomDomainScenario) createResources() error {
	customDomainResources, err := manifestprocessor.ParseFromFileWithTemplate("resources.yaml", "manifests/custom-domain", resourceSeparator, struct {
		Namespace      string
		NamePrefix     string
		TestID         string
		Domain         string
		Subdomain      string
		LoadBalancerIP string
	}{Namespace: namespace, NamePrefix: "custom-domain", TestID: c.testID, Domain: c.domain, Subdomain: fmt.Sprintf("%s.%s", c.testID, c.domain), LoadBalancerIP: c.loadBalancerIP.String()})
	if err != nil {
		return fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, customDomainResources...)

	if err != nil {
		return err
	}

	return nil
}

func (c *CustomDomainScenario) thereIsAnCloudCredentialsSecret(secretName string, secretNamespace string) error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	_, err := k8sClient.Resource(res).Namespace(secretNamespace).Get(context.Background(), secretName, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("cloud credenials secret could not be found")
	}

	return nil
}

func (c *CustomDomainScenario) isDNSReady() error {
	err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		testName := generateRandomString(3)
		ips, err := net.LookupIP(fmt.Sprintf("%s.%s.%s", testName, c.testID, c.domain))
		if err != nil {
			return false, nil
		}
		if len(ips) != 0 {
			for _, ip := range ips {

				if ip.Equal(c.loadBalancerIP) {
					fmt.Printf("Found %s.%s.%s. IN A %s\n", testName, c.testID, c.domain, ip.String())
					return true, nil
				}
			}
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("DNS record could not be looked up: %s", err)
	}
	return nil
}

func (c *CustomDomainScenario) thereIsAnExposedService(svcName string, svcNamespace string) error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	svc, err := k8sClient.Resource(res).Namespace(svcNamespace).Get(context.Background(), svcName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("istio-ingressgateway service was not found")
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil || found != true {
		return fmt.Errorf("could not get load balancer status from the service: %s", err)
	}
	loadBalancerIngress, _ := ingress[0].(map[string]interface{})

	loadBalancerIP, err := helpers.GetLoadBalancerIp(loadBalancerIngress)
	if err != nil {
		return fmt.Errorf("could not extract load balancer IP from istio service: %s", err)
	}
	c.loadBalancerIP = loadBalancerIP

	return nil
}

func (c *CustomDomainScenario) thereIsAnUnsecuredEndpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, c.apiResourceOne)
}

func (c *CustomDomainScenario) callingTheEndpointWithAnyTokenShouldResultInStatusBetween(endpoint string, arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(map[string]string{authorizationHeaderName: anyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (c *CustomDomainScenario) secureWithOAuth2() error {
	return helper.APIRuleWithRetries(batch.UpdateResources, batch.UpdateResources, k8sClient, c.apiResourceTwo)
}

func (c *CustomDomainScenario) callingTheEndpointWithAInvalidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(map[string]string{authorizationHeaderName: anyToken}, fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (c *CustomDomainScenario) callingTheEndpointWithAValidTokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), "Opaque", &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}, nil, nil)
}

func (c *CustomDomainScenario) callingTheEndpointWithoutATokenShouldResultInStatusBetween(endpoint string, lower int, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s/%s", c.url, strings.TrimLeft(endpoint, "/")), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
