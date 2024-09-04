MODULE_NAME ?= api-gateway

#latest api-gateway release
LATEST_RELEASE = $(shell curl -sS "https://api.github.com/repos/kyma-project/api-gateway/releases/latest" | jq -r '.tag_name')

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.3

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

# Image URL to use all building/pushing image targets
IMG_REGISTRY_PORT ?= $(MODULE_REGISTRY_PORT)
IMG_REGISTRY ?= op-skr-registry.localhost:$(IMG_REGISTRY_PORT)/unsigned/operator-images
IMG ?= $(IMG_REGISTRY)/$(MODULE_NAME)-operator:$(MODULE_VERSION)

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

##@ Development

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

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
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test $(shell go list ./... | grep -v /tests/integration) -coverprofile cover.out

.PHONY: test-integration
test-integration: test-integration-v2alpha1 test-integration-ory test-integration-istio test-integration-gateway

.PHONY: test-integration-v2alpha1
test-integration-v2alpha1: generate fmt vet
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestV2alpha1

.PHONY: test-integration-ory
test-integration-ory: generate fmt vet
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestOryJwt

.PHONY: test-integration-istio
test-integration-istio: generate fmt vet
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestIstioJwt

.PHONY: test-integration-gateway
test-integration-gateway: generate fmt vet
	IS_GARDENER=$(IS_GARDENER) source ./tests/integration/env_vars.sh && TEST_CONCURRENCY=1 go test -timeout 1h ./tests/integration -run "^TestGateway$$" -v -race

.PHONY: test-upgrade
test-upgrade: generate fmt vet generate-upgrade-test-manifest install-istio deploy-latest-release ## Run API Gateway upgrade tests.
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestUpgrade .

.PHONY: test-custom-domain
test-custom-domain: generate fmt vet
	source ./tests/integration/env_vars_custom_domain.sh && bash -c "trap 'kubectl delete secret google-credentials -n default' EXIT; \
             kubectl create secret generic google-credentials -n default --from-file=serviceaccount.json=${TEST_SA_ACCESS_KEY_PATH}; \
             GODEBUG=netdns=cgo CGO_ENABLED=1 go test -timeout 1h ./tests/integration -run "^TestCustomDomain$$" -v -race"

.PHONY: install-istio
install-istio: create-namespace
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests build
	go run ./main.go

.PHONY: docker-build
docker-build:
	IMG=$(IMG) docker build -t ${IMG} --build-arg TARGET_OS=${TARGET_OS} --build-arg TARGET_ARCH=${TARGET_ARCH} --build-arg VERSION=${VERSION} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
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
	@if ! kubectl get crd virtualservices.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/networking.istio.io_virtualservice.yaml; fi;
	@if ! kubectl get crd peerauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_peerauthentication.yaml; fi;
	@if ! kubectl get crd authorizationpolicies.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_authorizationpolicy.yaml; fi;
	@if ! kubectl get crd requestauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_requestauthentication.yaml; fi;
	@if ! kubectl get crd gateways.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/networking.istio.io_gateways.yaml; fi;
	@if ! kubectl get crd dnsentries.dns.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/dns.gardener.cloud_dnsentry.yaml; fi;
	@if ! kubectl get crd certificates.cert.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/cert.gardener.cloud_certificate.yaml; fi;

.PHONY: uninstall
uninstall: manifests kustomize module-version ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: create-namespace
create-namespace:
	kubectl create namespace kyma-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace kyma-system istio-injection=enabled --overwrite

.PHONY: deploy
deploy: manifests kustomize module-version create-namespace ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
YQUERY ?= $(LOCALBIN)/yq

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.5
CONTROLLER_TOOLS_VERSION ?= v0.14.0
YQ_VERSION ?= v4

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

.PHONY: yq
yq: $(YQUERY) ## Download yq locally if necessary.
$(YQUERY): $(LOCALBIN)
	test -s $(LOCALBIN)/yq || { go get github.com/mikefarah/yq/$(YQ_VERSION) ; GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/$(YQ_VERSION) ; }

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: module-image
module-image: docker-build docker-push ## Build the Module Image and push it to a registry defined in IMG_REGISTRY
	echo "built and pushed module image $(IMG)"

.PHONY: generate-manifests
generate-manifests: kustomize module-version
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > api-gateway-manager.yaml

.PHONY: get-latest-release
get-latest-release:
	@echo $(LATEST_RELEASE)

########## Performance Tests ###########
.PHONY: perf-test
perf-test:
	cd performance_tests && ./test.sh

