PROJECT_ROOT ?= ../..

MODULE_VERSION ?= 0.0.1

MODULE_SHA ?= ${shell git rev-parse --short HEAD}

.PHONY: kyma
kyma:
	@make -C ${PROJECT_ROOT} kyma

.PHONY: module-build
module-build:
	@make -C ${PROJECT_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}-${MODULE_SHA} \
		MODULE_CHANNEL=fast
	@echo "\n- creating template for regular channel"
	@sed -e 's/([- ])fast/\\1regular/g' <${PROJECT_ROOT}/template-fast.yaml >${PROJECT_ROOT}/template-regular.yaml

.PHONY: module-template-arifact
module-template-artifact:
	@cp ${PROJECT_ROOT}/template-fast.yaml ${ARTIFACTS}
	@cp ${PROJECT_ROOT}/template-regular.yaml ${ARTIFACTS}

.PHONY: module-template-print
module-template-print:
	@echo "\n~~~~~~~~~~~~BEGINNING OF MODULE TEMPLATE FAST CHANNEL~~~~~~~~~~~~~~"
	@cat ${PROJECT_ROOT}/template-fast.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE FAST CHANNEL~~~~~~~~~~~~~~~~"
	@echo "\n~~~~~~~~~~~~BEGINNING OF MODULE TEMPLATE REGULAR CHANNEL~~~~~~~~~~~~~~"
	@cat ${PROJECT_ROOT}/template-regular.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE REGULAR CHANNEL~~~~~~~~~~~~~~~~"

.PHONY: module-build-and-print
all: module-build module-template-artifact module-template-print

