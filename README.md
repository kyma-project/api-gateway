# Api-Gateway Controller (name to be changed)

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- access to K8s environment: minikube or a remote K8s cluster

## How to use it

- `start minikube`
- `make build` to build the binary and run tests
- `eval $(minikube docker-env)`
- `make build-image` to build a docker image
- Update the `patches` field in `config/crd/kustomization.yaml` and `config/default/kustomization.yaml` to `patchesStrategicMerge`
- `make deploy` to deploy controller to the minikube

## Example CR structure:

```yaml
---
gateway: kyma-gateway.kyma-system.svc.cluster.local
service:
  name: foo-service
  port: 8080
  host: foo.bar
  external: true/false
auth: 
  name: JWT
  config:
    issuer: http://dex.kyma.local
    jwks: []
    mode: 
      name: ALL
      config:
        scopes: []
---
gateway: kyma-gateway.kyma-system.svc.cluster.local
service:
  name: foo-service
  port: 8080
  host: foo.bar
  external: true/false
auth: 
  name: JWT
  config:
    issuer: http://dex.kyma.local
    jwks: []
    mode: 
      name: EXCLUDE
      config:
        - pathSuffix: '/c'
        - pathRegex: '/d/*'
        - pathPrefix: ''
        - pathExact: '/f/foobar.png'
---
gateway: kyma-gateway.kyma-system.svc.cluster.local
service:
  name: foo-service
  port: 8080
  host: foo.bar
  external: true/false
auth: 
  name: JWT
  config:
    issuer: http://dex.kyma.local
    jwks: []
    mode: 
      name: INCLUDE
      config:
        - path: '/a'
          scopes: 
            - read
          methods:
            - GET
            - POST
        - path: '/b'
          methods:
            - GET
---
gateway: kyma-gateway.kyma-system.svc.cluster.local
service:
  name: foo-service
  port: 8080
  host: foo.bar
  external: true/false
auth:
  name: PASSTHROUGH
---
gateway: kyma-gateway.kyma-system.svc.cluster.local
service:
  name: foo-service
  port: 8080
  host: foo.bar
  external: true/false
auth:
  name: OAUTH
  config:
    - path: '/a'
      scopes: 
        - write
      methods:
        - POST
    # Invalid or takes priority
    - path: '/*' 
      scopes: []
      methods: []
```