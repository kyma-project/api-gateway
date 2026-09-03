MODULE_NAME ?= api-gateway

#latest api-gateway release
LATEST_RELEASE = $(shell curl -sS "https://api.github.com/repos/kyma-project/api-gateway/releases/latest" | jq -r '.tag_name')

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

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
# target descriptions by '##'. The awk command is responsible for reading the
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

.PHONY: sync-vendors-crds
sync-vendors-crds: ## Vendor third-party CRDs into hack/crds from Go module dependencies (istio.io/api, gardener/*). Run after bumping these modules in go.mod.
	@set -e; \
	GARDENER_CERT_DIR=$$(go list -m -f '{{.Dir}}' github.com/gardener/cert-management)/pkg/apis/cert/crds; \
	GARDENER_DNS_DIR=$$(go list -m -f '{{.Dir}}' github.com/gardener/external-dns-management)/pkg/apis/dns/crds; \
	ISTIO_CRDS=$$(go list -m -f '{{.Dir}}' istio.io/api)/kubernetes/customresourcedefinitions.gen.yaml; \
	cp $$GARDENER_CERT_DIR/cert.gardener.cloud_certificates.yaml hack/crds/gardener/cert.gardener.cloud_certificate.yaml; \
	cp $$GARDENER_DNS_DIR/dns.gardener.cloud_dnsentries.yaml     hack/crds/gardener/dns.gardener.cloud_dnsentry.yaml; \
	cp $$ISTIO_CRDS                                              hack/crds/istio.gen.yaml; \
	chmod u+w hack/crds/*.yaml hack/crds/gardener/*.yaml; \
	echo "Vendored gardener and istio CRDs from go.mod versions."
	@echo "NOTE: oathkeeper CRD (hack/crds/oathkeeper.ory.sh_rules.yaml) is not in any Go module — still manual."

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
test: manifests generate fmt vet setup-envtest ## Generate manifests and run tests.
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test $$(go list ./... | grep -v /tests/integration | grep -v /tests/e2e ) -coverprofile cover.out

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

.PHONY: install-istio-experimental
install-istio-experimental: create-namespace
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager-experimental.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s

DUAL_STACK_ENABLED ?= true

.PHONY: create-provisioning-info
create-provisioning-info: create-namespace
	printf 'networkDetails:\n  dualStackIPEnabled: %s\n' "$(DUAL_STACK_ENABLED)" \
	  | kubectl create configmap -n kyma-system kyma-provisioning-info \
		  --from-file=details=/dev/stdin \
		  --dry-run=client -o yaml \
	  | kubectl apply -f -

.PHONY: install-istio-manager
install-istio-manager: create-namespace
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet
	go run ./cmd/main.go

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

.PHONY: create-namespace
create-namespace:
	kubectl create namespace kyma-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace kyma-system istio-injection=enabled --overwrite

.PHONY: deploy
deploy: img-check manifests kustomize module-version create-namespace ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: e2e-test
e2e-test:
	make -C tests/e2e/tests e2e-test

.PHONY: e2e-test-part1
e2e-test-part1:
	make -C tests/e2e/tests e2e-test-part1

.PHONY: e2e-test-part2
e2e-test-part2:
	make -C tests/e2e/tests e2e-test-part2

.PHONY: e2e-test-custom-domain
e2e-test-custom-domain:
	make -C tests/e2e/tests e2e-test-custom-domain

.PHONY: e2e-test-gateway-gardener
e2e-test-gateway-gardener:
	make -C tests/e2e/tests e2e-test-gateway-gardener

.PHONY: dualstack-e2e-test-part1
dualstack-e2e-test-part1: ## Dualstack subset, part 1: LB-dialling suites only (asterisk, cors, expose_methods_on_paths, extgateway). Non-network suites (request, service, validation) are excluded — the family axis is inert for them.
	make -C tests/e2e/tests dualstack-e2e-test-part1

.PHONY: dualstack-e2e-test-part2
dualstack-e2e-test-part2: ## Dualstack subset, part 2: auth + in-cluster HTTP suites (extauth, jwt, no_auth, short_host).
	make -C tests/e2e/tests dualstack-e2e-test-part2

.PHONY: dualstack-e2e-test-custom-domain
dualstack-e2e-test-custom-domain: ## Dualstack: custom_domain suite, per-family HTTPS LB dials.
	make -C tests/e2e/tests dualstack-e2e-test-custom-domain

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOTESTSUM ?= $(LOCALBIN)/gotestsum

## Tool Versions
KUSTOMIZE_VERSION ?= v5.7.1
CONTROLLER_TOOLS_VERSION ?= v0.19.0
ENVTEST_VERSION ?= latest
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: module-version
module-version:
	sed 's/VERSION/$(VERSION)/g' config/default/kustomization.template.yaml > config/default/kustomization.yaml

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: gotestsum
gotestsum:
	$(call go-install-tool,$(GOTESTSUM),gotest.tools/gotestsum,latest)

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(realpath $(1)-$(3)) $(1)
endef

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

CRD_REF_DOCS_FLAGS = --max-depth=15 --renderer=markdown --config=crd-ref-docs/config.yaml --templates-dir=crd-ref-docs/templates

.PHONY: generate-crd-docs
generate-crd-docs: bin/crd-ref-docs ## Generate CRD docs
	./bin/crd-ref-docs $(CRD_REF_DOCS_FLAGS) --output-path=docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md --source-path=apis/gateway/v2
	sed -i'' -e 's/Optional: \\{\\}/Optional/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	sed -i'' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	sed -i'' -e 's/XIntOrString: \\{\\}/XIntOrString/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	sed -i'' -e 's/\\}/\}/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	sed -i'' -e 's/\\{/\{/g' docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md
	rm docs/user/custom-resources/apirule/04-10-apirule-custom-resource.md-e
	./bin/crd-ref-docs $(CRD_REF_DOCS_FLAGS) --output-path=docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md --source-path=apis/operator/v1alpha1
	sed -i'' -e 's/Optional: \\{\\}/Optional/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
	sed -i'' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
	sed -i'' -e 's/XIntOrString: \\{\\}/XIntOrString/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
	sed -i'' -e 's/\\}/\}/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
	sed -i'' -e 's/\\{/\{/g' docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md
	rm docs/user/custom-resources/apigateway/04-00-apigateway-custom-resource.md-e
	./bin/crd-ref-docs $(CRD_REF_DOCS_FLAGS) --output-path=docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md --source-path=apis/gateway/ratelimit/v1alpha1
	sed -i'' -e 's/Optional: \\{\\}/Optional/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md 
	sed -i'' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md 
	sed -i'' -e 's/XIntOrString: \\{\\}/XIntOrString/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md
	sed -i'' -e 's/\\}/\}/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md
	sed -i'' -e 's/\\{/\{/g' docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md
	rm docs/user/custom-resources/ratelimit/04-10-ratelimit-custom-resource.md-e
	./bin/crd-ref-docs $(CRD_REF_DOCS_FLAGS) --output-path=docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md --source-path=apis/gateway/external/v1alpha1
	sed -i'' -e 's/Optional: \\{\\}/Optional/g' docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md
	sed -i'' -e 's/Required: \\{\\}/Required/g' docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md
	sed -i'' -e 's/XIntOrString: \\{\\}/XIntOrString/g' docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md
	sed -i'' -e 's/\\}/\}/g' docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md
	sed -i'' -e 's/\\{/\{/g' docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md
	rm docs/user/custom-resources/externalgateway/externalgateway-custom-resource.md-e