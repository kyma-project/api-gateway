package api_gateway

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/spf13/pflag"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	testIDLength              = 8
	manifestsDirectory        = "manifests/"
	globalCommonResourcesFile = "global-commons.yaml"
	resourceSeparator         = "---"
	exportResultVar           = "EXPORT_RESULT"
	cucumberFileName          = "cucumber-report.json"
	anyToken                  = "any"
	authorizationHeaderName   = "Authorization"
	authorizationHeaderPrefix = "Bearer"
	opaqueHeaderName          = "opaque-token"
	defaultNS                 = "kyma-system"
	configMapName             = "api-gateway-config"
)

var (
	resourceManager *resource.Manager
	conf            Config
	httpClient      *http.Client
	k8sClient       dynamic.Interface
	helper          *helpers.Helper
	oauth2Cfg       *clientcredentials.Config
	jwtConfig       *jwt.Config
	batch           *resource.Batch
	namespace       string
	secondNamespace string
	jwtHeaderName   string
	fromParamName   string
)

var t *testing.T
var goDogOpts = godog.Options{
	Output:   colors.Colored(os.Stdout),
	Format:   "pretty",
	TestingT: t,
}

type Config struct {
	CustomDomain     string        `envconfig:"TEST_CUSTOM_DOMAIN,default=test.domain.kyma"`
	IssuerUrl        string        `envconfig:"TEST_OIDC_ISSUER_URL"`
	ClientID         string        `envconfig:"TEST_CLIENT_ID"`
	ClientSecret     string        `envconfig:"TEST_CLIENT_SECRET"`
	User             string        `envconfig:"TEST_USER_EMAIL,default=admin@kyma.cx"`
	Pwd              string        `envconfig:"TEST_USER_PASSWORD,default=1234"`
	ReqTimeout       uint          `envconfig:"TEST_REQUEST_TIMEOUT,default=180"`
	ReqDelay         uint          `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain           string        `envconfig:"TEST_DOMAIN"`
	GatewayName      string        `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace string        `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	ClientTimeout    time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` // Don't forget the unit!
	TestConcurrency  int           `envconfig:"TEST_CONCURRENCY,default=1"`
}

type Scenario interface {
	GetUrl() string
	GetNamespace() string
}

type BaseScenario struct {
	namespace string
	domain    string
	testID    string
}

func (b *BaseScenario) GetNamespace() string {
	return b.namespace
}

// ScenarioWithRawAPIResource is a scenario that doesn't create APIRule on scenario initialization, allowing further templating of APIRule manifest
type ScenarioWithRawAPIResource struct {
	BaseScenario
	apiResourceManifestPath string
	apiResourceDirectory    string
	manifestTemplate        map[string]string
	url                     string
}

func InitTestSuite() {
	pflag.Parse()
	goDogOpts.Paths = pflag.Args()

	if os.Getenv(exportResultVar) == "true" {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	if err := envconfig.Init(&conf); err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}

	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: conf.ClientTimeout,
	}

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(conf.ReqDelay) * time.Second),
		retry.Attempts(conf.ReqTimeout / conf.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	}

	helper = helpers.NewHelper(httpClient, commonRetryOpts)

	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		t.Fatal(err)
	}

	client, err := client.GetDynamicClient()
	if err != nil {
		t.Fatal(err)
	}

	k8sClient = client
	resourceManager = &resource.Manager{RetryOptions: commonRetryOpts}

	batch = &resource.Batch{
		ResourceManager: resourceManager,
		Mapper:          mapper,
	}
}

func SetupCommonResources(namePrefix string) {
	namespace = fmt.Sprintf("%s-%s", namePrefix, generateRandomString(6))
	secondNamespace = fmt.Sprintf("%s-2", namespace)
	log.Printf("Using namespace: %s\n", namespace)

	oauth2Cfg = &clientcredentials.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", conf.IssuerUrl),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	config, err := jwt.NewJwtConfig()
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig = &config

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate(globalCommonResourcesFile, manifestsDirectory, resourceSeparator, struct {
		Namespace string
	}{
		Namespace: namespace,
	})
	if err != nil {
		log.Fatal(err)
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := batch.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	err = resourceManager.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	_, err = batch.CreateResources(k8sClient, globalCommonResources...)
	if err != nil {
		log.Fatal(err)
	}
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateScenarioWithRawAPIResource(templateFileName string, namePrefix string) (*ScenarioWithRawAPIResource, error) {
	testID := generateRandomString(testIDLength)

	template := make(map[string]string)
	template["Namespace"] = namespace
	template["NamePrefix"] = namePrefix
	template["TestID"] = testID
	template["Domain"] = conf.Domain
	template["GatewayName"] = conf.GatewayName
	template["GatewayNamespace"] = conf.GatewayNamespace
	template["IssuerUrl"] = conf.IssuerUrl
	template["JWTHeaderName"] = jwtHeaderName
	template["FromParamName"] = fromParamName
	template["EncodedCredentials"] = base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", conf.ClientID, conf.ClientSecret)))

	return &ScenarioWithRawAPIResource{
		BaseScenario: BaseScenario{
			namespace: namespace,
			testID:    testID,
			domain:    conf.Domain,
		},
		manifestTemplate:        template,
		apiResourceManifestPath: templateFileName,
		apiResourceDirectory:    manifestsDirectory,
	}, nil
}

func SwitchJwtHandler(jwtHandler string) (string, error) {
	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		return "", err
	}
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind("ConfigMap"))
	if err != nil {
		return "", err
	}
	currentJwtHandler, configMap, err := getConfigMapJwtHandler(mapping.Resource)
	if err != nil {
		configMap := unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "ConfigMap",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      configMapName,
					"namespace": defaultNS,
				},
				"data": map[string]interface{}{
					"api-gateway-config": "jwtHandler: " + jwtHandler,
				},
			},
		}
		currentJwtHandler = jwtHandler
		err = resourceManager.CreateResource(k8sClient, mapping.Resource, defaultNS, configMap)
	}
	if err != nil {
		return "", fmt.Errorf("could not get or create jwtHandler config:\n %+v", err)
	}
	if currentJwtHandler != jwtHandler {
		configMap.Object["data"].(map[string]interface{})["api-gateway-config"] = "jwtHandler: " + jwtHandler
		err = resourceManager.UpdateResource(k8sClient, mapping.Resource, defaultNS, configMapName, *configMap)
		if err != nil {
			return "", fmt.Errorf("unable to update ConfigMap:\n %+v", err)
		}
	}
	return currentJwtHandler, err
}

func getConfigMapJwtHandler(gvr schema.GroupVersionResource) (string, *unstructured.Unstructured, error) {
	res, err := resourceManager.GetResource(k8sClient, gvr, defaultNS, configMapName)
	if err != nil {
		return "", res, fmt.Errorf("could not get ConfigMap:\n %+v", err)
	}
	data, found, err := unstructured.NestedMap(res.Object, "data")
	if err != nil || !found {
		return "", res, fmt.Errorf("could not find data in the ConfigMap:\n %+v", err)
	}
	return strings.Split(data["api-gateway-config"].(string), ": ")[1], res, nil
}
