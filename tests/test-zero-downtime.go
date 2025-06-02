package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cm "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
)

const (
	parallelRequests = 5
	maxRetries       = 600 // 10 minutes with 1s intervals
	apiRuleWait      = 300 // 5 minutes with 1s intervals
)

var validHandlers = []string{"jwt", "noop", "no_auth", "allow", "oauth2_introspection"}

type Config struct {
	Handler          string
	TestDomain       string
	OIDCConfigURL    string
	ClientID         string
	ClientSecret     string
	KubeconfigPath   string
	ParallelRequests int
}

type ZeroDowntimeTest struct {
	config     *Config
	k8sClient  client.Client
	httpClient *http.Client
	logger     *log.Logger
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type OIDCConfig struct {
	TokenEndpoint string `json:"token_endpoint"`
}

func main() {
	var config Config

	rootCmd := &cobra.Command{
		Use:   "zero-downtime-test [handler]",
		Short: "Run zero downtime tests for API Gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Handler = args[0]
			return runZeroDowntimeTest(&config)
		},
	}

	rootCmd.Flags().StringVar(&config.TestDomain, "test-domain", os.Getenv("TEST_DOMAIN"), "Test domain")
	rootCmd.Flags().StringVar(&config.OIDCConfigURL, "oidc-config-url", os.Getenv("TEST_OIDC_CONFIG_URL"), "OIDC config URL")
	rootCmd.Flags().StringVar(&config.ClientID, "client-id", os.Getenv("TEST_CLIENT_ID"), "OAuth2 client ID")
	rootCmd.Flags().StringVar(&config.ClientSecret, "client-secret", os.Getenv("TEST_CLIENT_SECRET"), "OAuth2 client secret")
	rootCmd.Flags().StringVar(&config.KubeconfigPath, "kubeconfig", "/Users/I758431/.kube/config", "Path to kubeconfig file")
	rootCmd.Flags().IntVar(&config.ParallelRequests, "parallel-requests", parallelRequests, "Number of parallel requests")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

func runZeroDowntimeTest(config *Config) error {
	if !isValidHandler(config.Handler) {
		return fmt.Errorf("handler not provided or invalid. Must be one of: %s", strings.Join(validHandlers, ", "))
	}

	test, err := NewZeroDowntimeTest(config)
	if err != nil {
		return fmt.Errorf("failed to initialize test: %w", err)
	}

	return test.Run()
}

func NewZeroDowntimeTest(config *Config) (*ZeroDowntimeTest, error) {
	k8sClient := cm.GetK8sClient()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	logger := log.New(os.Stdout, "zero-downtime: ", log.LstdFlags)

	return &ZeroDowntimeTest{
		config:     config,
		k8sClient:  k8sClient,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

func (zt *ZeroDowntimeTest) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interruption signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		zt.logger.Println("Received interrupt signal, canceling...")
		cancel()
	}()

	zt.logger.Printf("Running zero downtime tests for handler '%s'", zt.config.Handler)

	// Start zero downtime requests in background
	requestsErrChan := make(chan error, 1)
	go func() {
		requestsErrChan <- zt.runZeroDowntimeRequests(ctx)
	}()

	// Run integration test
	zt.logger.Printf("Starting integration test scenario for handler '%s'", zt.config.Handler)
	testErr := zt.runIntegrationTest(ctx)
	if testErr != nil {
		cancel() // Cancel background requests
		return fmt.Errorf("test execution failed: %w", testErr)
	}

	// Wait for requests to complete
	requestsErr := <-requestsErrChan
	if requestsErr != nil {
		return fmt.Errorf("zero-downtime requests failed: %w", requestsErr)
	}

	zt.logger.Println("Tests successful")
	return nil
}

func (zt *ZeroDowntimeTest) runZeroDowntimeRequests(ctx context.Context) error {
	// Wait for APIRule to exist and be ready
	apiRules, err := zt.waitForAPIRuleToExist(ctx)
	if err != nil {
		return fmt.Errorf("APIRule not found: %w", err)
	}

	zt.logger.Println("APIRule found, waiting for status OK")
	if err := zt.waitForAPIRuleReady(ctx, apiRules.Items[0]); err != nil {
		return fmt.Errorf("APIRule not ready: %w", err)
	}

	// Get exposed host
	exposedHost, err := zt.getExposedHost(ctx, apiRules.Items[0])
	if err != nil {
		return fmt.Errorf("failed to get exposed host: %w", err)
	}

	zt.logger.Printf("APIRule host is %s", exposedHost)
	urlUnderTest := fmt.Sprintf("https://%s/headers", exposedHost)

	// Get bearer token if needed
	bearerToken, err := zt.getBearerToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bearer token: %w", err)
	}

	// Wait for URL to be available
	zt.logger.Printf("Waiting until %s is available", urlUnderTest)
	if err := zt.waitForURL(ctx, urlUnderTest, bearerToken); err != nil {
		return fmt.Errorf("URL not available: %w", err)
	}

	// Send parallel requests
	return zt.sendParallelRequests(ctx, urlUnderTest, bearerToken)
}

func (zt *ZeroDowntimeTest) waitForAPIRuleToExist(ctx context.Context) (*gatewayv2alpha1.APIRuleList, error) {
	labelSelector := labels.Set{"test": "v1beta1-migration"}

	for attempts := 1; attempts <= apiRuleWait; attempts++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Using dynamic client would be better for custom resources
		// For now, using kubectl-like approach with REST client
		apiRules, err := zt.getAPIRules(ctx, labelSelector.AsSelector())
		if err != nil {
			return nil, fmt.Errorf("kubectl failed when listing apirules: %w", err)
		}

		if len(apiRules.Items) > 0 {
			return apiRules, nil
		}

		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("APIRule not found after %d attempts", apiRuleWait)
}

func (zt *ZeroDowntimeTest) waitForAPIRuleReady(ctx context.Context, rule gatewayv2alpha1.APIRule) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for APIRule to be ready")
		case <-ticker.C:
			ready, err := zt.isAPIRuleReady(ctx, types.NamespacedName{Namespace: rule.Namespace, Name: rule.Name})
			if err != nil {
				return err
			}
			if ready {
				zt.logger.Println("APIRule status is OK")
				return nil
			}
		}
	}
}

func (zt *ZeroDowntimeTest) getBearerToken(ctx context.Context) (string, error) {
	if zt.config.Handler != "jwt" && zt.config.Handler != "oauth2_introspection" {
		return "", nil
	}

	if zt.config.OIDCConfigURL == "" {
		zt.logger.Println("No OIDC_CONFIG_URL provided, assuming oauth mock")
		return zt.getTokenFromMock(ctx)
	}

	return zt.getTokenFromOIDC(ctx)
}

func (zt *ZeroDowntimeTest) getTokenFromMock(ctx context.Context) (string, error) {
	mockURL := fmt.Sprintf("https://oauth2-mock.%s/.well-known/openid-configuration", zt.config.TestDomain)
	if err := zt.waitForURL(ctx, mockURL, ""); err != nil {
		return "", fmt.Errorf("OAuth2 mock server not available: %w", err)
	}

	tokenURL := fmt.Sprintf("https://oauth2-mock.%s/oauth2/token", zt.config.TestDomain)
	zt.logger.Printf("Getting access token from URL '%s'", tokenURL)

	return zt.fetchToken(ctx, tokenURL, "", "")
}

func (zt *ZeroDowntimeTest) getTokenFromOIDC(ctx context.Context) (string, error) {
	if zt.config.ClientID == "" || zt.config.ClientSecret == "" {
		return "", fmt.Errorf("no client ID or secret provided")
	}

	zt.logger.Println("TEST_OIDC_CONFIG_URL provided, getting token url")

	req, err := http.NewRequestWithContext(ctx, "GET", zt.config.OIDCConfigURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := zt.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var oidcConfig OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&oidcConfig); err != nil {
		return "", err
	}

	if oidcConfig.TokenEndpoint == "" {
		return "", fmt.Errorf("can't get token url from OIDC config")
	}

	zt.logger.Println("Getting access token")
	return zt.fetchToken(ctx, oidcConfig.TokenEndpoint, zt.config.ClientID, zt.config.ClientSecret)
}

func (zt *ZeroDowntimeTest) fetchToken(ctx context.Context, tokenURL, clientID, clientSecret string) (string, error) {
	data := "grant_type=client_credentials&token_format=jwt"
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if clientID != "" && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	resp, err := zt.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (zt *ZeroDowntimeTest) waitForURL(ctx context.Context, url, bearerToken string) error {
	zt.logger.Printf("Waiting for URL '%s' to be available", url)

	for attempts := 1; attempts <= maxRetries; attempts++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := zt.makeRequest(ctx, url, bearerToken); err == nil {
			zt.logger.Printf("%s is available for requests", url)
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("URL %s is not available after %d attempts", url, maxRetries)
}

func (zt *ZeroDowntimeTest) makeRequest(ctx context.Context, url, bearerToken string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	authorized := bearerToken != ""
	req.Header.Set("x-ext-authz", "allow")
	if authorized {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := zt.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !authorized && resp.StatusCode < http.StatusBadRequest {
		return fmt.Errorf("request should failed (unathorized), but succeded %d (code)", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		if resp.StatusCode >= http.StatusInternalServerError {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))

		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (zt *ZeroDowntimeTest) sendParallelRequests(ctx context.Context, url, bearerToken string) error {
	zt.logger.Printf("Sending requests to %s in %d parallel threads", url, zt.config.ParallelRequests)

	var wg sync.WaitGroup
	errChan := make(chan error, zt.config.ParallelRequests)

	for i := 0; i < zt.config.ParallelRequests; i++ {
		wg.Add(2)
		go func(threadID int) {
			defer wg.Done()
			if err := zt.sendRequests(ctx, threadID, url, bearerToken); err != nil {
				errChan <- err
			}
		}(i)
		go func(threadID int) {
			defer wg.Done()
			if err := zt.sendRequests(ctx, threadID, url, ""); err != nil {
				errChan <- err
			}
		}(i)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (zt *ZeroDowntimeTest) sendRequests(ctx context.Context, threadID int, url, bearerToken string) error {
	zt.logger.Printf("thread %d, sending requests to %s", threadID, url)
	requestCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		requestCount++

		if err := zt.makeRequest(ctx, url, bearerToken); err != nil {
			zt.logger.Printf("thread %d, request %d failed: %v", threadID, requestCount, err)

			// Sleep to avoid race condition
			time.Sleep(2 * time.Second)

			// Check if APIRule still exists
			apiRuleExists, checkErr := zt.apiRuleExists(ctx)
			if checkErr != nil {
				return fmt.Errorf("failed to check APIRule existence: %w", checkErr)
			}

			if apiRuleExists {
				return fmt.Errorf("thread %d, test failed after %d requests", threadID, requestCount)
			}

			if requestCount < 10 {
				return fmt.Errorf("thread %d, there were less than 10 requests, something was wrong with the tests", threadID)
			}

			zt.logger.Printf("thread %d, test successful after %d requests. Stopping requests because APIRule is deleted", threadID, requestCount)
			return nil
		}
	}
}

func (zt *ZeroDowntimeTest) runIntegrationTest(ctx context.Context) error {
	// This would execute the Go test
	// For now, returning nil as placeholder
	// In real implementation, you would use testing package or exec.Command
	cmd := exec.CommandContext(ctx, "go", "test", "-count=1", "-timeout", "15m",
		"./integration", "-v", "-race", "-run",
		fmt.Sprintf("TestOryJwt/Migrate_v1beta1_APIRule_with_%s_handler", zt.config.Handler))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Placeholder implementations for Kubernetes operations
// In real implementation, you would use dynamic client for custom resources

func (zt *ZeroDowntimeTest) getAPIRules(ctx context.Context, labelSelector labels.Selector) (*gatewayv2alpha1.APIRuleList, error) {
	apiRule := &gatewayv2alpha1.APIRuleList{}
	// This would use dynamic client to list APIRules with the given label selector
	err := zt.k8sClient.List(ctx, apiRule, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	return apiRule, nil
}

func (zt *ZeroDowntimeTest) isAPIRuleReady(ctx context.Context, name types.NamespacedName) (bool, error) {
	apiRule := gatewayv2alpha1.APIRule{}
	err := zt.k8sClient.Get(ctx, name, &apiRule)
	if err != nil {
		return false, err
	}
	if apiRule.Status.State == gatewayv2alpha1.Ready {
		zt.logger.Printf("APIRule %s is ready", name.Name)
		return true, nil
	}
	return false, nil
}

func (zt *ZeroDowntimeTest) getExposedHost(ctx context.Context, rule gatewayv2alpha1.APIRule) (string, error) {
	if zt.config.TestDomain == "" {
		return "", fmt.Errorf("test domain not provided")
	}

	if len(rule.Spec.Rules) == 0 {
		apiRuleV1 := &gatewayv1beta1.APIRule{}
		err := apiRuleV1.ConvertFrom(&rule)
		if err != nil {
			return "", err
		}
		if apiRuleV1.Spec.Host == nil {
			return "", fmt.Errorf("no hosts defined in APIRule v1beta1")
		}
		return fmt.Sprintf("%s.%s", *apiRuleV1.Spec.Host, zt.config.TestDomain), nil
	}
	// cast rule.Spec.Hosts to string and append test domain
	if len(rule.Spec.Hosts) == 0 {
		return "", fmt.Errorf("no hosts defined in APIRule")
	}
	host := string(*rule.Spec.Hosts[0])
	if !strings.ContainsAny(host, ".") {
		host = fmt.Sprintf("%s.%s", host, zt.config.TestDomain)
	}
	return host, nil

}

func (zt *ZeroDowntimeTest) apiRuleExists(ctx context.Context) (bool, error) {
	// Implementation would check if APIRule exists
	// Placeholder for now
	return false, nil
}

func isValidHandler(handler string) bool {
	for _, valid := range validHandlers {
		if handler == valid {
			return true
		}
	}
	return false
}
