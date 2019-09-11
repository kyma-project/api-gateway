# Api-Gateway Controller (name to be changed)

## Overview

The API Gateway Controller manages Istio authentication Policies, VirtualServices and Oathkeeper Rule. The controller allows to expose services using instances of the `gate.gateway.kyma-project.io` custom resource (CR).

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- access to K8s environment: minikube or a remote K8s cluster

## Details

### Run the controller locally

- `start minikube`
- `make build` to build the binary and run tests
- `eval $(minikube docker-env)`
- `make build-image` to build a docker image
- export `OATHKEEPER_SVC_ADDRESS`, `OATHKEEPER_SVC_PORT` and `JWKS_URI` variables
- `make deploy` to deploy controller

### Use command-line flags

| Name | Required | Description | Possible values |
|------|----------|-------------|-----------------|
| **oathkeeper-svc-address** | yes | ory oathkeeper-proxy service address. | `ory-oathkeeper-proxy.kyma-system.svc.cluster.local` |
| **oathkeeper-svc-port** | yes | ory oathkeeper-proxy service port. | `4455` |
| **jwks-uri** | yes | default jwksUri in the Policy. | any string |

### Example CR structure

Valid examples of the Gate CR can be found in the `config/samples` catalog. 