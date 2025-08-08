.PHONY: gotestsum
gotestsum:
	@test -s $(LOCALBIN)/gotestsum || GOBIN=$(LOCALBIN) go install gotest.tools/gotestsum@latest

.PHONY: e2e-test
e2e-test: gotestsum
	@echo "Running e2e tests"
	go clean -testcache
	gotestsum --format testname --rerun-fails --junitfile tests.xml --jsonfile tests.json
	@echo "E2E tests completed successfully"
