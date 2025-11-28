MODULE_NAME ?= api-gateway

#latest api-gateway release
LATEST_RELEASE = $(shell curl -sS "https://api.github.com/repos/kyma-project/api-gateway/releases/latest" | jq -r '.tag_name')

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

APP_NAME = api-gateway-manager

COMPONENT_CLI_VERSION ?= latest

# Upgrade integration test variables
TARGET_BRANCH ?= ""
TEST_UPGRADE_IMG ?= ""

IS_GARDENER ?= false

VERSION ?= dev

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: img-check
img-check:
	$(if $(IMG),,$(error IMG must be set))

##@ Development

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook output:webhook:artifacts:config=config/admission-webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate-upgrade-test-manifest
generate-upgrade-test-manifest: manifests kustomize module-version
	cd config/manager && $(KUSTOMIZE) edit set image controller=${TEST_UPGRADE_IMG}
	$(KUSTOMIZE) build config/default -o tests/integration/testsuites/upgrade/manifests/upgrade-test-generated-operator-manifest.yaml

.PHONY: deploy-latest-release
deploy-latest-release: create-namespace
	./tests/integration/scripts/deploy-latest-release-to-cluster.sh $(TARGET_BRANCH)

# Generate code
.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Generate manifests and run tests.
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test $(shell go list ./... | grep -v /tests/integration | grep -v /tests/e2e | grep -v /tests/ui) -coverprofile cover.out

.PHONY: test-integration
test-integration: test-integration-v2alpha1 test-integration-ory test-integration-istio test-integration-gateway

.PHONY: test-integration-v2alpha1
test-integration-v2alpha1: generate ## Run API Gateway integration tests with v2alpha1 API.
	go test -timeout 1h ./tests/integration -v -race -run TestV2alpha1

.PHONY: test-integration-ory
test-integration-ory: generate
	go test -timeout 1h ./tests/integration -v -race -run TestOryJwt

.PHONY: test-migration-zero-downtime-%
test-migration-zero-downtime-%: generate
	go test -timeout 1h ./tests/integration -v -race -run TestOryZeroDowntimeMigration/.*_$*_handler.*

.PHONY: test-integration-istio
test-integration-istio: generate
	go test -timeout 1h ./tests/integration -v -race -run TestIstioJwt

.PHONY: test-integration-gateway
test-integration-gateway: generate
	go test -timeout 1h ./tests/integration -run TestGateway -v -race

.PHONY: test-upgrade
test-upgrade: generate generate-upgrade-test-manifest install-istio deploy-latest-release ## Run API Gateway upgrade tests.
	go test -timeout 1h ./tests/integration -v -race -run TestUpgrade .

.PHONY: test-custom-domain
test-custom-domain: generate
	GODEBUG=netdns=cgo CGO_ENABLED=1 go test -timeout 1h ./tests/integration -run "^TestCustomDomain$$" -v -race

.PHONY: test-integration-rate-limit
test-integration-rate-limit: generate
	go test -timeout 1h ./tests/integration -run TestRateLimit -v -race

.PHONY: test-integration-v2
test-integration-v2: generate ## Run API Gateway integration tests with v2 API.
	go test -timeout 1h ./tests/integration -v -race -run "^TestV2$$"

.PHONY: install-istio
install-istio: create-namespace
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s


.PHONY: install-istio-manager
install-istio-manager: create-namespace
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet
	go run ./main.go

TARGET_OS ?= linux
TARGET_ARCH ?= amd64
.PHONY: docker-build
docker-build: img-check
	IMG=$(IMG) docker buildx build -t ${IMG} --platform=${TARGET_OS}/${TARGET_ARCH} --build-arg VERSION=${VERSION} .

.PHONY: docker-push
docker-push: img-check ## Push docker image with the manager.
	docker push ${IMG}

##@ Local

.PHONY: local-run
local-run:
	make -C hack/local run

.PHONY: local-stop
local-stop:
	make -C hack/local stop

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

# Install CRDs into a cluster
.PHONY: install
install: manifests kustomize module-version
	$(KUSTOMIZE) build config/crd | kubectl apply -f -
	@if ! kubectl get crd virtualservices.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/crds/networking.istio.io_virtualservice.yaml; fi;
	@if ! kubectl get crd peerauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/crds/security.istio.io_peerauthentication.yaml; fi;
	@if ! kubectl get crd authorizationpolicies.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/crds/security.istio.io_authorizationpolicy.yaml; fi;
	@if ! kubectl get crd requestauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/crds/security.istio.io_requestauthentication.yaml; fi;
	@if ! kubectl get crd gateways.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/crds/networking.istio.io_gateways.yaml; fi;
	@if ! kubectl get crd dnsentries.dns.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/crds/gardener/dns.gardener.cloud_dnsentry.yaml; fi;
	@if ! kubectl get crd certificates.cert.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/crds/gardener/cert.gardener.cloud_certificate.yaml; fi;

.PHONY: uninstall
uninstall: manifests kustomize module-version ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: create-namespace
create-namespace:
	kubectl create namespace kyma-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace kyma-system istio-injection=enabled --overwrite

.PHONY: deploy
deploy: img-check manifests kustomize module-version create-namespace ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: e2e-test
e2e-test:
	make -C tests/e2e/tests e2e-test

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.5
CONTROLLER_TOOLS_VERSION ?= v0.18.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: module-version
module-version:
	sed 's/VERSION/$(VERSION)/g' config/default/kustomization.template.yaml > config/default/kustomization.yaml

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: module-image
module-image: img-check docker-build docker-push ## Build the Module Image and push it to the registry
	echo "built and pushed module image $(IMG)"

.PHONY: generate-manifests
generate-manifests: img-check kustomize module-version
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > api-gateway-manager.yaml

.PHONY: get-latest-release
get-latest-release:
	@echo $(LATEST_RELEASE)

########## Performance Tests ###########
.PHONY: perf-test
perf-test:
	cd performance_tests && ./test.sh

########## Docs generation ###########
bin/crd-ref-docs:
	wget "https://github.com/elastic/crd-ref-docs/releases/download/v0.2.0/crd-ref-docs_0.2.0_${OS_TYPE}_${OS_ARCH}.tar.gz" -O bin/crd-ref-docs.tar.gz 
	mkdir -p bin/crd-ref-docs-x
	tar -xzf bin/crd-ref-docs.tar.gz -C bin/crd-ref-docs-x
	rm bin/crd-ref-docs.tar.gz
	mv bin/crd-ref-docs-x/crd-ref-docs bin/crd-ref-docs
	rm -r bin/crd-ref-docs-x

.PHONY: generate-crd-docs
generate-crd-docs: bin/crd-ref-docs ## Generate CRD reference docs
	./bin/crd-ref-docs \
	--max-depth=15 \
	--renderer=markdown \
	--config=crd-ref-docs/config.yaml \
	--templates-dir=crd-ref-docs/templates \
	--output-path=docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md \
	--source-path=apis/gateway/v2
	sed -i '' -e 's/Optional: \\{\\}/Optional/g' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	rm -f docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md-e

	./bin/crd-ref-docs \
	--max-depth=15 \
	--renderer=markdown \
	--config=crd-ref-docs/config.yaml \
	--templates-dir=crd-ref-docs/templates \
	--output-path=docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md \
	--source-path=apis/operator/v1alpha1
	sed -i '' -e 's/Optional: \\{\\}/Optional/g' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
    rm -f docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md-e

	./bin/crd-ref-docs \
	--max-depth=15 \
	--renderer=markdown \
	--config=crd-ref-docs/config.yaml \
	--templates-dir=crd-ref-docs/templates \
	--output-path=docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md \
	--source-path=apis/gateway/ratelimit/v1alpha1
	sed -i '' -e 's/Optional: \\{\\}/Optional/g' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md
	rm -f docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md-e