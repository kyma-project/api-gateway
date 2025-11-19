.PHONY: gotestsum
gotestsum:
	@test -s $(LOCALBIN)/gotestsum || GOBIN=$(LOCALBIN) go install gotest.tools/gotestsum@latest

.PHONY: e2e-test
e2e-test: gotestsum
	@echo "Running e2e tests"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/asterisk --format testname --rerun-fails --junitfile asterisk-tests.xml --jsonfile asterisk-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/cors --format testname --rerun-fails --junitfile cors-tests.xml --jsonfile cors-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/expose_methods_on_paths --format testname --rerun-fails --junitfile expose_methods_on_paths-tests.xml --jsonfile expose_methods_on_paths-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/extauth --format testname --rerun-fails --junitfile extauth-tests.xml --jsonfile extauth-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/jwt --format testname --rerun-fails --junitfile jwt-tests.xml --jsonfile jwt-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/no_auth --format testname --rerun-fails --junitfile no_auth-tests.xml --jsonfile no_auth-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/request --format testname --rerun-fails --junitfile request-tests.xml --jsonfile request-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/service --format testname --rerun-fails --junitfile service-tests.xml --jsonfile service-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/short_host --format testname --rerun-fails --junitfile short_host-tests.xml --jsonfile short_host-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/validation --format testname --rerun-fails --junitfile validation-tests.xml --jsonfile validation-tests.json

	@echo "E2E tests completed successfully"
