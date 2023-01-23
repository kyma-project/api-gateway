package api_gateway

import (
	"context"
	"crypto/tls"
	"encoding/base64"
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
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/kyma/common/ingressgateway"
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
	OauthClientSecretLength   = 8
	OauthClientIDLength       = 8
	manifestsDirectory        = "manifests/"
	testingAppFile            = "testing-app.yaml"
	globalCommonResourcesFile = "global-commons.yaml"
	istioJwtApiruleFile       = "istio-jwt-strategy.yaml"
	resourceSeparator         = "---"
	defaultHeaderName         = "Authorization"
	customDomainEnv           = "TEST_CUSTOM_DOMAIN"
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
	IasAddr          string        `envconfig:"TEST_IAS_ADDRESS"`
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
	IsMinikubeEnv    bool          `envconfig:"TEST_MINIKUBE_ENV,default=false"`
	TestConcurency   int           `envconfig:"TEST_CONCURENCY,default=1"`
}

type Scenario struct {
	namespace   string
	url         string
	apiResource []unstructured.Unstructured
}

type TwoStepScenario struct {
	namespace      string
	url            string
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

	if conf.IsMinikubeEnv {
		var err error
		log.Printf("Using dedicated ingress client")
		httpClient, err = ingressgateway.FromEnv().Client()
		if err != nil {
			log.Fatalf("Unable to initialize ingress gateway client: %v", err)
		}
	} else {
		log.Printf("Fallback to default http client")
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: conf.ClientTimeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
	randomSuffix6 := generateRandomString(6)
	oauthSecretName := fmt.Sprintf("%s-secret-%s", namePrefix, randomSuffix6)
	oauthClientName := fmt.Sprintf("%s-client-%s", namePrefix, randomSuffix6)
	log.Printf("Using namespace: %s\n", namespace)
	log.Printf("Using OAuth2Client with name: %s, secretName: %s\n", oauthClientName, oauthSecretName)
	oauth2Cfg = &clientcredentials.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", conf.IasAddr),
		Scopes:       []string{"read"},
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
	config, err := jwt.NewJwtConfig()
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig = &config
	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate(globalCommonResourcesFile, manifestsDirectory, resourceSeparator, struct {
		Namespace         string
		OauthClientSecret string
		OauthClientID     string
		OauthSecretName   string
	}{
		Namespace:         namespace,
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(conf.ClientID)),
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(conf.ClientSecret)),
		OauthSecretName:   oauthSecretName,
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

func getOAUTHToken(oauth2Cfg clientcredentials.Config) (*oauth2.Token, error) {
	var tokenOAUTH oauth2.Token
	err := retry.Do(
		func() error {
			token, err := oauth2Cfg.Token(context.Background())
			if err != nil {
				return fmt.Errorf("error during Token retrival: %+v", err)
			}

			if token == nil || token.AccessToken == "" {
				return fmt.Errorf("got empty OAuth2 token")
			}
			tokenOAUTH = *token

			return nil
		},
		retry.Delay(500*time.Millisecond), retry.Attempts(3))
	return &tokenOAUTH, err
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

func CreateScenario(templateFileName string, namePrefix string, deploymentFile ...string) (*Scenario, error) {
	testID := generateRandomString(testIDLength)
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
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRule, err := manifestprocessor.ParseFromFileWithTemplate(templateFileName, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
		IasAddr          string
	}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace, IasAddr: conf.IasAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}
	return &Scenario{namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), apiResource: accessRule}, nil
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
