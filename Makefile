APP_NAME = api-gateway-controller
IMG ?= $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

CRD_OPTIONS ?= "crd:trivialVersions=true,crdVersions=v1"
KYMA_COMPONENTS ?= "hack/kyma-components.yaml"

# Example ory-oathkeeper
ifndef OATHKEEPER_SVC_ADDRESS
override OATHKEEPER_SVC_ADDRESS = change-me
endif

# Example 4455
ifndef OATHKEEPER_SVC_PORT
override OATHKEEPER_SVC_PORT = change-me
endif

# kubernetes.default service.namespace
ifndef SERVICE_BLOCKLIST
override SERVICE_BLOCKLIST = change-me
endif

# kyma.local foo.bar bar
ifndef DOMAIN_ALLOWLIST
override DOMAIN_ALLOWLIST = change-me
endif

# CORS
ifndef CORS_ALLOW_ORIGINS
override CORS_ALLOW_ORIGINS = regex:.*
endif

ifndef CORS_ALLOW_METHODS
override CORS_ALLOW_METHODS = GET,POST,PUT,DELETE
endif

ifndef CORS_ALLOW_HEADERS
override CORS_ALLOW_HEADERS = Authorization,Content-Type,*
endif

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.1

GOCMD=go

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GOCMD) env GOBIN))
GOBIN=$(shell $(GOCMD) env GOPATH)/bin
else
GOBIN=$(shell $(GOCMD) env GOBIN)
endif

GOTEST=$(GOCMD) test -timeout 1h

PULL_IMAGE_VERSION=PR-${PULL_NUMBER}
POST_IMAGE_VERSION=v$(shell date '+%Y%m%d')-$(shell printf %.8s ${PULL_BASE_SHA})

ifeq ($(JOB_TYPE), postsubmit)
ifeq ($(TEST_UPGRADE_IMG),)
	TEST_UPGRADE_IMG="europe-docker.pkg.dev/kyma-project/dev/api-gateway-controller:${POST_IMAGE_VERSION}"
endif
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


.EXPORT_ALL_VARIABLES:
GO111MODULE = on

.PHONY: all
all: build

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
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" $(GOTEST) $(shell go list ./... | grep -v /tests/integration) -coverprofile cover.out

.PHONY: test-for-release
test-for-release: envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" $(GOTEST) $(shell go list ./... | grep -v /tests/integration) -coverprofile cover.out

.PHONY: test-integration
test-integration: generate fmt vet envtest ## Run integration tests.
	source ./tests/integration/env_vars.sh && $(GOTEST) ./tests/integration -v -race -run TestIstioJwt . && $(GOTEST) ./tests/integration -v -race -run TestOryJwt .

.PHONY: test-upgrade
test-upgrade: generate fmt vet install ## Run API Gateway upgrade tests.
	source ./tests/integration/env_vars.sh && $(GOTEST) ./tests/integration -v -race -run TestUpgrade .

test-custom-domain:
	source ./tests/integration/env_vars_custom_domain.sh && bash -c "trap 'kubectl delete secret google-credentials -n default' EXIT; \
             kubectl create secret generic google-credentials -n default --from-file=serviceaccount.json=${TEST_SA_ACCESS_KEY_PATH}; \
             GODEBUG=netdns=cgo CGO_ENABLED=1 $(GOTEST) ./tests/integration -run "^TestCustomDomain$$" -v -race"

.PHONY: kyma-cli
kyma-cli:
ifeq ($(UPGRADE_JOB), true)
	curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/latest/download/kyma_linux_x86_64.tar.gz"
	tar -C /usr/bin -zxvf kyma.tar.gz kyma
	rm -f kyma.tar.gz
	chmod +x /usr/bin/kyma
else
	curl -Lo /usr/bin/kyma https://storage.googleapis.com/kyma-cli-unstable/kyma-linux
	chmod +x /usr/bin/kyma
endif

.PHONY: k3d
k3d:
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.0.0 bash

.PHONY: provision-k3d
provision-k3d:
	kyma provision k3d --ci

.PHONY: kustomize install-kyma
install-kyma:
	kyma version
ifeq ($(KYMA_COMPONENTS), hack/kyma-components-no-istio.yaml)
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s
endif
ifndef JOB_TYPE
	kyma deploy --ci -s main -c ${KYMA_COMPONENTS}
else ifeq ($(UPGRADE_JOB), true)
	kyma deploy --ci -c ${KYMA_COMPONENTS}
else ifeq ($(JOB_TYPE), presubmit)
	kyma deploy --ci -s main -c ${KYMA_COMPONENTS} \
	  --value api-gateway.global.images.api_gateway_controller.version=${PULL_IMAGE_VERSION} \
	  --value api-gateway.global.images.api_gateway_controller.directory=dev
else ifeq ($(JOB_TYPE), postsubmit)
	kyma deploy --ci -s main -c ${KYMA_COMPONENTS} \
		--value api-gateway.global.images.api_gateway_controller.version=${POST_IMAGE_VERSION}
endif
	make install

.PHONY: test-integration-k3d
test-integration-k3d: kyma-cli k3d provision-k3d install-kyma test-integration ## Run integration tests.
	source ./tests/integration/env_vars.sh && $(GOTEST) ./tests/integration -v -race -run TestIstioJwt . && $(GOTEST) ./tests/integration -v -race -run TestOryJwt .

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: build-release
build-release: generate
	go build -o bin/manager main.go

.PHONY: run
run: build
	go run . --oathkeeper-svc-address=${OATHKEEPER_SVC_ADDRESS} --oathkeeper-svc-port=${OATHKEEPER_SVC_PORT} --service-blocklist=${SERVICE_BLOCKLIST} --domain-allowlist=${DOMAIN_ALLOWLIST}

.PHONY: docker-build
docker-build: pull-licenses test ## Build docker image with the manager.
	docker build -t $(APP_NAME):latest .

.PHONY: docker-build-release
docker-build-release: pull-licenses test-for-release ## Build docker image with the manager.
	docker build -t $(APP_NAME):latest .

.PHONY: docker-push
docker-push:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(IMG):$(TAG)
else
	@echo "Image signing skipped"
endif

.PHONY: pull-licenses
pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

# Install CRDs into a cluster
.PHONY: install
install: manifests kustomize
	kustomize build config/crd | kubectl apply -f -
	@if ! kubectl get crd virtualservices.networking.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/networking.istio.io_virtualservice.yaml; fi;
	@if ! kubectl get crd rules.oathkeeper.ory.sh > /dev/null 2>&1 ; then kubectl apply -f hack/oathkeeper.ory.sh_rules.yaml; fi;
	@if ! kubectl get crd authorizationpolicies.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_authorizationpolicy.yaml; fi;
	@if ! kubectl get crd requestauthentications.security.istio.io > /dev/null 2>&1 ; then kubectl apply -f hack/security.istio.io_requestauthentication.yaml; fi;

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

# Deploy the controller using "api-gateway-controller:latest" Docker image to the Kubernetes cluster configured in ~/.kube/config
.PHONY: deploy-dev
deploy-dev: manifests patch-gen
	$(KUSTOMIZE) build config/development | kubectl apply -f -

.PHONY: deploy
deploy: generate manifests patch-gen kustomize install ## Deploy controller to the K8s cluster specified in ~/.kube/config.
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
CLIENT_GEN ?= $(LOCALBIN)/client-gen
INFORMER_GEN = $(LOCALBIN)/informer-gen
LISTER_GEN = $(LOCALBIN)/lister-gen
GOLANG_CI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.2
GOLANG_CI_LINT_VERSION ?= v1.46.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)


.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(GOCMD) install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(GOCMD) install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: client-gen
client-gen: ## Download client-gen
	GOBIN=$(LOCALBIN) $(GOCMD) install k8s.io/code-generator/cmd/client-gen@v0.23.4

.PHONY: informer-gen
informer-gen: ## Download informer-gen
	GOBIN=$(LOCALBIN) $(GOCMD) install k8s.io/code-generator/cmd/informer-gen@v0.23.4

.PHONY: lister-gen
lister-gen: ## Download lister-gen
	GOBIN=$(LOCALBIN) $(GOCMD) install k8s.io/code-generator/cmd/lister-gen@v0.23.4

.PHONY: lint
lint: ## Run golangci-lint against code.
	GOBIN=$(LOCALBIN) $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)
	$(LOCALBIN)/golangci-lint run

.PHONY: archive
archive:
	cp -r bin/* $(ARTIFACTS)

.PHONY: release
release:
	./hack/release.sh

##@ ci targets
.PHONY: ci-pr
ci-pr: build test docker-build docker-push

.PHONY: ci-main
ci-main: build docker-build docker-push

.PHONY: ci-release
ci-release: TAG=${shell git describe --abbrev=0 --tags}
ci-release: build-release docker-build-release docker-push archive release

.PHONY: clean
clean:
	rm -rf bin

# Augment kustomize patch files with env-specific variables
patch-gen:
	@cat config/default/manager_args_patch.yaml.tmpl |\
		sed -e 's|OATHKEEPER_SVC_ADDRESS|${OATHKEEPER_SVC_ADDRESS}|g' |\
		sed -e 's|OATHKEEPER_SVC_PORT|${OATHKEEPER_SVC_PORT}|g' |\
		sed -e 's|SERVICE_BLOCKLIST|${SERVICE_BLOCKLIST}|g' |\
		sed -e 's|DOMAIN_ALLOWLIST|${DOMAIN_ALLOWLIST}|g' |\
		sed -e 's|CORS_ALLOW_ORIGINS|${CORS_ALLOW_ORIGINS}|g' |\
		sed -e 's|CORS_ALLOW_METHODS|${CORS_ALLOW_METHODS}|g' |\
		sed -e 's|CORS_ALLOW_HEADERS|${CORS_ALLOW_HEADERS}|g' > config/default/manager_args_patch.yaml

# Generate static installation files
static: manifests patch-gen
	kustomize build config/released -o install/k8s


# Deploy controller using a released Docker image to the Kubernetes cluster configured in ~/.kube/config
deploy: manifests patch-gen
	kustomize build config/default | kubectl apply -f -

# Generate CRD with patches
gen-crd:
	kustomize build config/crd > config/crd/apirules.gateway.crd.yaml

samples-clean:
	kubectl delete -f config/samples/valid.yaml --ignore-not-found=true
	kubectl delete -f config/samples/invalid.yaml --ignore-not-found=true

.PHONY: samples
samples: samples-valid

.PHONY: samples-valid
samples-valid: samples-clean
	kubectl apply -f config/samples/valid.yaml

.PHONY: samples-invalid
samples-invalid: samples-clean
	kubectl apply -f config/samples/invalid.yaml

########## Performance Tests ###########
.PHONY: perf-test
perf-test:
	cd performance_tests && ./test.sh
