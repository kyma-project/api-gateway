.PHONY: run-k3d-with-api-gateway
run-k3d-with-api-gateway:
	@make -C ./api-gateway run

.PHONY: verify
verify:
	@make -C ./api-gateway verify

.PHONY: run
run: run-k3d-with-api-gateway

.PHONY: test
test: run verify
	@echo '### api-gateway operator operational'

.PHONY: stop
stop:
	@make -C ./api-gateway stop
