package api_gateway

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"gitlab.com/rodrigoodhin/gocure/report/html"

	"github.com/avast/retry-go"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/spf13/pflag"
	"github.com/tidwall/pretty"
	"github.com/vrischmann/envconfig"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	testIDLength              = 8
	manifestsDirectory        = "manifests/"
	testingAppFile            = "testing-app.yaml"
	globalCommonResourcesFile = "global-commons.yaml"
	resourceSeparator         = "---"
	exportResultVar           = "EXPORT_RESULT"
	junitFileName             = "junit-report.xml"
	cucumberFileName          = "cucumber-report.json"
	anyToken                  = "any"
	authorizationHeaderName   = "Authorization"
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
	TestConcurency   int           `envconfig:"TEST_CONCURENCY,default=1"`
}

type Scenario interface {
	GetApiResource() []unstructured.Unstructured
	GetUrl() string
	GetNamespace() string
}

type BaseScenario struct {
	namespace string
	url       string
}

func (b *BaseScenario) GetUrl() string {
	return b.url
}

func (b *BaseScenario) GetNamespace() string {
	return b.namespace
}

type UnstructuredScenario struct {
	BaseScenario
	apiResource []unstructured.Unstructured
}

func (u *UnstructuredScenario) GetApiResource() []unstructured.Unstructured {
	return u.apiResource
}

// ScenarioWithRawAPIResource is a scenario that doesn't create APIRule on scenario initialization, allowing further templating of APIRule manifest
type ScenarioWithRawAPIResource struct {
	BaseScenario
	apiResourceManifestPath string
	apiResourceDirectory    string
	manifestTemplate        map[string]string
}

type TwoStepScenario struct {
	BaseScenario
	apiResourceOne []unstructured.Unstructured
	apiResourceTwo []unstructured.Unstructured
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

func generateReport() {
	htmlOutputDir := "reports/"

	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: htmlOutputDir,
			Title:            "Kyma API-Gateway component tests",
			Metadata: models.Metadata{
				Platform:        runtime.GOOS,
				TestEnvironment: "Gardener GCP",
				Parallel:        "Scenarios",
				Executed:        "Remote",
				AppVersion:      "main",
				Browser:         "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
		if path == "reports" {
			return nil
		}

		data, err1 := os.ReadFile(path)
		if err1 != nil {
			return err
		}

		//Format all patterns like "&lt" to not be replaced later
		find := regexp.MustCompile(`&\w\w`)
		formatted := find.ReplaceAllFunc(data, func(b []byte) []byte {
			return []byte{b[0], ' ', b[1], b[2]}
		})

		err = os.WriteFile(path, formatted, fs.FileMode(02))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf(err.Error())
	}

	if artifactsDir, ok := os.LookupEnv("ARTIFACTS"); ok {
		err = filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
			if path == "reports" {
				return nil
			}

			_, err1 := copy(path, fmt.Sprintf("%s/report.html", artifactsDir))
			if err1 != nil {
				return err1
			}
			return nil
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		_, err = copy("./junit-report.xml", fmt.Sprintf("%s/junit-report.xml", artifactsDir))
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

}

func getApiRules() string {
	res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
	list, _ := k8sClient.Resource(res).List(context.Background(), v1.ListOptions{})

	toPrint, _ := json.Marshal(list)

	return string(pretty.Pretty(toPrint))
}

func CreateScenarioWithRawAPIResource(templateFileName string, namePrefix string, deploymentFile ...string) (*ScenarioWithRawAPIResource, error) {
	testID := generateRandomString(testIDLength)

	err := createCommonResources(testID, deploymentFile...)
	if err != nil {
		return nil, err
	}
	template := make(map[string]string)

	template["Namespace"] = namespace
	template["NamePrefix"] = namePrefix
	template["TestID"] = testID
	template["Domain"] = conf.Domain
	template["GatewayName"] = conf.GatewayName
	template["GatewayNamespace"] = conf.GatewayNamespace
	template["IssuerUrl"] = conf.IssuerUrl

	return &ScenarioWithRawAPIResource{
		BaseScenario: BaseScenario{
			namespace: namespace,
			url:       fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain),
		},
		manifestTemplate:        template,
		apiResourceManifestPath: templateFileName,
		apiResourceDirectory:    manifestsDirectory,
	}, nil
}

func createCommonResources(testID string, deploymentFile ...string) error {
	deploymentFileName := testingAppFile
	if len(deploymentFile) > 0 {
		deploymentFileName = deploymentFile[0]
	}

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(deploymentFileName, manifestsDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return err
	}
	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func getPodListReport() string {
	type returnedPodList struct {
		PodList []struct {
			Metadata struct {
				Name              string `json:"name"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	list, _ := k8sClient.Resource(res).Namespace("").List(context.Background(), v1.ListOptions{})

	p := returnedPodList{}
	toMarshal, _ := json.Marshal(list)
	err := json.Unmarshal(toMarshal, &p)
	if err != nil {
		log.Fatalf(err.Error())
	}
	toPrint, _ := json.Marshal(p)
	return string(pretty.Pretty(toPrint))
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
