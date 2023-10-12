# Module Name used for bundling the OCI Image and later on for referencing in the Kyma Modules
MODULE_NAME ?= api-gateway

# Module Registry used for pushing the image
MODULE_REGISTRY_PORT ?= 8888
MODULE_REGISTRY ?= op-kcp-registry.localhost:$(MODULE_REGISTRY_PORT)/unsigned
# Desired Channel of the Generated Module Template
MODULE_TEMPLATE_CHANNEL ?= stable
MODULE_CHANNEL ?= fast

ifndef MODULE_VERSION
    MODULE_VERSION = 0.0.1
endif

#latest api-gateway release
LATEST_RELEASE = $(shell curl -sS "https://api.github.com/repos/kyma-project/api-gateway/releases/latest" | jq -r '.tag_name')

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.2

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

# It is required for upgrade integration test
TARGET_BRANCH ?= ""
TEST_UPGRADE_IMG ?= ""

IS_GARDENER ?= false

# This will change the flags of the `kyma alpha module create` command in case we spot credentials
# Otherwise we will assume http-based local registries without authentication (e.g. for k3d)
ifneq (,$(PROW_JOB_ID))
GCP_ACCESS_TOKEN=$(shell gcloud auth application-default print-access-token)
MODULE_CREATION_FLAGS=--registry $(MODULE_REGISTRY) --module-archive-version-overwrite -c oauth2accesstoken:$(GCP_ACCESS_TOKEN)
else ifeq (,$(MODULE_CREDENTIALS))
MODULE_CREATION_FLAGS=--registry $(MODULE_REGISTRY) --module-archive-version-overwrite --insecure
else
MODULE_CREATION_FLAGS=--registry $(MODULE_REGISTRY) --module-archive-version-overwrite -c $(MODULE_CREDENTIALS)
endif

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
generate-upgrade-test-manifest: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${TEST_UPGRADE_IMG}
	$(KUSTOMIZE) build config/default -o tests/integration/testsuites/upgrade/manifests/upgrade-test-generated-operator-manifest.yaml

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
test-integration: generate fmt vet envtest ## Run integration tests.
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestIstioJwt . && go test -timeout 1h ./tests/integration -v -race -run TestOryJwt . && TEST_CONCURRENCY=1 go test -timeout 1h ./tests/integration -v -race -run TestGateway .

.PHONY: test-upgrade
test-upgrade: generate fmt vet generate-upgrade-test-manifest ## Run API Gateway upgrade tests.
	source ./tests/integration/env_vars.sh && go test -timeout 1h ./tests/integration -v -race -run TestUpgrade .

.PHONY: test-custom-domain
test-custom-domain: generate fmt vet
	source ./tests/integration/env_vars_custom_domain.sh && bash -c "trap 'kubectl delete secret google-credentials -n default' EXIT; \
             kubectl create secret generic google-credentials -n default --from-file=serviceaccount.json=${TEST_SA_ACCESS_KEY_PATH}; \
             GODEBUG=netdns=cgo CGO_ENABLED=1 go test -timeout 1h ./tests/integration -run "^TestCustomDomain$$" -v -race"

.PHONY: test-integration-gateway
test-integration-gateway:
	IS_GARDENER=$(IS_GARDENER) source ./tests/integration/env_vars.sh && TEST_CONCURRENCY=1 go test -timeout 1h ./tests/integration -run "^TestGateway$$" -v -race

.PHONY: install-prerequisites
install-prerequisites:
	kyma deploy --ci -s main -c hack/kyma-components.yaml

.PHONY: install-prerequisites-with-istio-from-manifest
install-prerequisites-with-istio-from-manifest:
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s
	kyma deploy --ci -s main -c hack/kyma-components-no-istio.yaml

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests build
	go run ./main.go

.PHONY: docker-build
docker-build:
	IMG=$(IMG) docker build -t ${IMG} --build-arg TARGETOS=${TARGETOS} --build-arg TARGETARCH=${TARGETARCH} .

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
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -
	@if ! kubectl get crd virtualservices.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/networking.istio.io_virtualservice.yaml; fi;
	@if ! kubectl get crd peerauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_peerauthentication.yaml; fi;
	@if ! kubectl get crd authorizationpolicies.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_authorizationpolicy.yaml; fi;
	@if ! kubectl get crd requestauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_requestauthentication.yaml; fi;
	@if ! kubectl get crd gateways.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/networking.istio.io_gateways.yaml; fi;
	@if ! kubectl get crd dnsentries.dns.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/dns.gardener.cloud_dnsentry.yaml; fi;
	@if ! kubectl get crd certificates.cert.gardener.cloud > /dev/null 2>&1 ; then kubectl apply -f hack/cert.gardener.cloud_certificate.yaml; fi;

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
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
CONTROLLER_TOOLS_VERSION ?= v0.10.0
YQ_VERSION ?= v4

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

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

##@ Module
.PHONY: all
all: module-build

.PHONY: module-image
module-image: docker-build docker-push ## Build the Module Image and push it to a registry defined in IMG_REGISTRY
	echo "built and pushed module image $(IMG)"

.PHONY: module-build
module-build: kyma kustomize ## Build the Module and push it to a registry defined in MODULE_REGISTRY
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KYMA) alpha create module \
		--kubebuilder-project $(SECURITY_SCAN_OPTIONS) --channel=${MODULE_CHANNEL} \
		--name kyma-project.io/module/$(MODULE_NAME) --version $(MODULE_VERSION) \
		--default-cr ./config/samples/operator_v1alpha1_apigateway.yaml \
		--path . --output "template-${MODULE_CHANNEL}.yaml" $(MODULE_CREATION_FLAGS)

.PHONY: generate-manifests
generate-manifests: kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > api-gateway-manager.yaml

##@ Tools

REPOSITORY=
GITHUB_URL=

.PHONY: get-latest-release
get-latest-release:
	@echo $(LATEST_RELEASE)

########## Kyma CLI ###########
KYMA_STABILITY ?= unstable

# $(call os_error, os-type, os-architecture)
define os_error
$(error Error: unsuported platform OS_TYPE:$1, OS_ARCH:$2; to mitigate this problem set variable KYMA with absolute path to kyma-cli binary compatible with your operating system and architecture)
endef

KYMA_FILE_NAME ?= $(shell ./hack/get_kyma_file_name.sh ${OS_TYPE} ${OS_ARCH})
KYMA ?= $(LOCALBIN)/kyma-$(KYMA_STABILITY)
kyma: $(LOCALBIN) $(KYMA) ## Download kyma locally if necessary.
$(KYMA):
	## Detecting operating system to download proper kyma CLI binary
	$(if $(KYMA_FILE_NAME),,$(call os_error, ${OS_TYPE}, ${OS_ARCH}))
	test -f $@ || curl -s -Lo $(KYMA) https://storage.googleapis.com/kyma-cli-$(KYMA_STABILITY)/$(KYMA_FILE_NAME)
	chmod 0100 $(KYMA)


########## Performance Tests ###########
.PHONY: perf-test
perf-test:
	cd performance_tests && ./test.sh