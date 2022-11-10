# API-Gateway JWT Handler Feature Toggle

## Overview

We support two JWT handlers in API-Gateway at the moment. Switching between them is configurable via the following ConfigMap: `kyma-system/api-gateway-config`

## Swithing between JWT handlers

### Enabling `ory/oathkeeper` JWT handler

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: ory"}}'
```

### Enabling `istio` JWT handler

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: istio"}}'
```
