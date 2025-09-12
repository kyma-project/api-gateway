.PHONY: gotestsum
gotestsum:
	@test -s $(LOCALBIN)/gotestsum || GOBIN=$(LOCALBIN) go install gotest.tools/gotestsum@latest

.PHONY: e2e-test
e2e-test: gotestsum
	@echo "Running e2e tests"
	go clean -testcache
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/extauth --format testname --rerun-fails --junitfile extauth-tests.xml --jsonfile extauth-tests.json
	gotestsum --packages github.com/kyma-project/api-gateway/tests/e2e/tests/jwt --format testname --rerun-fails --junitfile jwt-tests.xml --jsonfile jwt-tests.json
	@echo "E2E tests completed successfully"
