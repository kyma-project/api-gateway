package ory

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/tester"
	"golang.org/x/oauth2/clientcredentials"
	"log"
	"net/http"
	"strings"
	"time"
)

const requestTimeout = 10 * time.Second
const testInterval = 500 * time.Millisecond
const numberOfThreads = 5

// ZeroDowntimeTestRunner holds Tester instances created by the StartZeroDowntimeTest function,
// so then the FinishZeroDowntimeTests function knows which Testers to stop
// Every test scenario instance should have own copy of this structure to allow parallel execution of tests
type ZeroDowntimeTestRunner struct {
	testers   []tester.Tester
	jwtConfig *clientcredentials.Config
	host      string
}

func getClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: requestTimeout,
	}
}

func doSingleRequest(client *http.Client, host string, path string, url string, token string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Host = host

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("x-ext-authz", "allow")

	now := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request done at %s to host %s path %s failed because of: %w", now, host, path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request done at %s to host %s path %s failed with status code: %d", now, host, path, resp.StatusCode)
	}

	time.Sleep(testInterval)
	return nil
}

func (zd *ZeroDowntimeTestRunner) StartZeroDowntimeTest(ctx context.Context, path string) (context.Context, error) {
	var token string
	if zd.jwtConfig != nil {
		tokenJwt, err := auth.GetAccessToken(*zd.jwtConfig, "jwt")
		if err != nil {
			return ctx, fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		token = tokenJwt
	}

	url := fmt.Sprintf("https://%s/%s", zd.host, strings.TrimLeft(path, "/"))
	client := getClient()

	testFn := func() error {
		return doSingleRequest(client, zd.host, path, url, token)
	}

	tester := tester.NewTester(fmt.Sprintf("host-%s", zd.host), testFn, numberOfThreads)
	zd.testers = append(zd.testers, tester)

	log.Printf("Starting zero downtime tester %s", tester.Name())
	tester.Start()

	return ctx, nil
}

func (zd *ZeroDowntimeTestRunner) FinishZeroDowntimeTests(ctx context.Context) (context.Context, error) {
	allErrs := make([]error, 0)

	for _, tester := range zd.testers {
		log.Printf("Stopping zero downtime tester %s", tester.Name())
		results := tester.Stop()
		for _, result := range results {
			if result.Err != nil {
				log.Printf("Got result from worker %s: requests: %d, error: %v", result.WorkerName, result.TestCount, result.Err)
				allErrs = append(allErrs, result.Err)
			} else {
				log.Printf("Got result from worker %s: requests: %d", result.WorkerName, result.TestCount)
			}
		}
	}

	if len(allErrs) > 0 {
		return ctx, fmt.Errorf("errors happened during zero downtime tests: %w", errors.Join(allErrs...))
	}

	return ctx, nil
}

func (zd *ZeroDowntimeTestRunner) CleanZeroDowntimeTests(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	for _, tester := range zd.testers {
		log.Printf("Stopping zero downtime tester %s", tester.Name())
		tester.Stop()
	}
	return ctx, nil
}
