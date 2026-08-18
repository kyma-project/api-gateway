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
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/extgateway --format testname --rerun-fails --junitfile extgateway-tests.xml --jsonfile extgateway-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/kyma_gateway --format testname --rerun-fails --junitfile kyma_gateway-tests.xml --jsonfile kyma_gateway-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/kyma_gateway_k3d --format testname --rerun-fails --junitfile kyma_gateway_k3d-tests.xml --jsonfile kyma_gateway_k3d-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/deletion --format testname --rerun-fails --junitfile deletion-tests.xml --jsonfile deletion-tests.json

	@echo "E2E tests completed successfully"

.PHONY: e2e-test-part1
e2e-test-part1: gotestsum
	@echo "Running e2e tests part 1"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/asterisk --format testname --rerun-fails --junitfile asterisk-tests.xml --jsonfile asterisk-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/cors --format testname --rerun-fails --junitfile cors-tests.xml --jsonfile cors-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/expose_methods_on_paths --format testname --rerun-fails --junitfile expose_methods_on_paths-tests.xml --jsonfile expose_methods_on_paths-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/kyma_gateway --format testname --rerun-fails --junitfile kyma_gateway-tests.xml --jsonfile kyma_gateway-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/kyma_gateway_k3d --format testname --rerun-fails --junitfile kyma_gateway_k3d-tests.xml --jsonfile kyma_gateway_k3d-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/request --format testname --rerun-fails --junitfile request-tests.xml --jsonfile request-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/validation --format testname --rerun-fails --junitfile validation-tests.xml --jsonfile validation-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/extgateway --format testname --rerun-fails --junitfile extgateway-tests.xml --jsonfile extgateway-tests.json

	@echo "E2E tests part 1 completed successfully"

.PHONY: e2e-test-part2
e2e-test-part2: gotestsum
	@echo "Running e2e tests part 2"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/extauth --format testname --rerun-fails --junitfile extauth-tests.xml --jsonfile extauth-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/jwt --format testname --rerun-fails --junitfile jwt-tests.xml --jsonfile jwt-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/no_auth --format testname --rerun-fails --junitfile no_auth-tests.xml --jsonfile no_auth-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/service --format testname --rerun-fails --junitfile service-tests.xml --jsonfile service-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/short_host --format testname --rerun-fails --junitfile short_host-tests.xml --jsonfile short_host-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/deletion --format testname --rerun-fails --junitfile deletion-tests.xml --jsonfile deletion-tests.json

	@echo "E2E tests part 2 completed successfully"

.PHONY: e2e-test-custom-domain
e2e-test-custom-domain: gotestsum
	@echo "Running e2e tests for custom domain"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/custom_domain --format testname --rerun-fails --junitfile custom_domain-tests.xml --jsonfile custom_domain-tests.json
	@echo "E2E tests for custom domain completed successfully"

.PHONY: e2e-test-gateway-gardener
e2e-test-gateway-gardener: gotestsum
	@echo "Running e2e tests for gateway gardener"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/gateway/kyma_gateway_gardener --format testname --rerun-fails --junitfile gardener-tests.xml --jsonfile gardener-tests.json
	@echo "E2E tests for gateway gardener completed successfully"
