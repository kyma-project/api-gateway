CLUSTER_NAME ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry

MANAGER_IMAGE_NAME ?= api-gateway-manager-dev-local
MANAGER_IMAGE_TAG ?= 0.0.1

PROJECT_ROOT ?= ../../..

KYMA ?= ${PROJECT_ROOT}/bin/kyma-unstable

.PHONY: kyma
kyma:
	@make -C ${PROJECT_ROOT} kyma

.PHONY: delete-k3d-registry
delete-k3d-registry:
	-k3d registry delete ${REGISTRY_NAME}

.PHONY: delete-k3d-cluster
delete-k3d-cluster:
	-k3d cluster delete ${CLUSTER_NAME}

.PHONY: create-k3d
create-k3d:
	${KYMA} provision k3d --registry-port ${REGISTRY_PORT} --name ${CLUSTER_NAME} --ci

.PHONY: delete-k3d
delete-k3d: delete-k3d-cluster delete-k3d-registry

.PHONY: create-kyma-system-ns
create-kyma-system-ns:
	kubectl create ns kyma-system

.PHONY: install-crd
install-crd:
	@make -C ${PROJECT_ROOT} install

.PHONY: module-image
module-image:
	@make -C ${PROJECT_ROOT} module-image \
		IMG=localhost:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}

.PHONY: clean-charts
clean-charts:
	@rm -rf ${PROJECT_ROOT}/charts

.PHONY: module-build
module-build: clean-charts
	@make -C ${PROJECT_ROOT} module-build \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG} \
		MODULE_REGISTRY=localhost:${REGISTRY_PORT}/unsigned
		MODULE_CHANNEL=fast

.PHONY: yq-install
yq-install:
	@make -C ${PROJECT_ROOT} yq

.PHONY: fix-template
fix-template:
	@cat ${PROJECT_ROOT}/template-fast.yaml \
	| sed -e 's/remote/control-plane/g' \
		-e 's/${REGISTRY_PORT}/5000/g' \
	    -e 's/localhost/k3d-${REGISTRY_NAME}.localhost/g' \
		-e 's/\/unsigned//g' \
	> ${PROJECT_ROOT}/template-k3d.yaml
	${PROJECT_ROOT}/bin/yq -i 'select(.metadata.labels | . += {"operator.kyma-project.io/use-local-template": "true"})' ${PROJECT_ROOT}/template-k3d.yaml

.PHONY: enable-module
enable-module:
	$(KYMA) alpha enable module api-gateway -c fast

.PHONY: patch-mod-mgr-role
patch-mod-mgr-role:
	kubectl patch clusterrole \
		lifecycle-manager-manager-role \
		--patch-file=./k3d-patches/patch-k3d-lifecycle-manager-clusterrole.yaml

.PHONY: install-module
install-module:
	${KYMA} alpha deploy --ci
	kubectl apply -f ${PROJECT_ROOT}/template-k3d.yaml

.PHONY: run
run: kyma \
	create-k3d \
	create-kyma-system-ns \
	install-crd \
	module-image \
	yq-install \
	module-build \
	fix-template \
	install-module \
	patch-mod-mgr-role \
	enable-module \
	verify

.PHONY: stop
stop: delete-k3d

.PHONY: verify
verify:
	@${PROJECT_ROOT}/hack/verify_kyma_status.sh

