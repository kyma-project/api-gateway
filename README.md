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
- `make deploy` to deploy controller to the minikube

## Example CR structure:

```yaml
---
application:
  service:
    name: foo-service
    port: 8080
  hostURL: https://foo.bar
authentication:
  type: 
  - name: JWT
    config:
      issuer: http://dex.kyma.local
      jwks: []
      mode: 
      - name: ALL
        config:
          scopes: []
      - name: EXCLUDE
        config:
          - pathSuffix: '/c'
          - pathRegex: '/d/*'
          - pathPrefix: ''
          - pathExact: '/f/foobar.png'
      - name: INCLUDE
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
  - name: PASSTHROUGH
    config: {}  
  - name: OAUTH
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