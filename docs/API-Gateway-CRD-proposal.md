# API Gateway CRD proposal

This document describes proposed API for installing APIGateway component.

```yaml
kind: APIGateway
spec:
  apiRuleControllerConfig:
    defaultCors:
      origins:
        regex:
          - ".*"
      methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "PATCH"
      headers:
      - "Authorization"
      - "Content-Type"
      - "*"
    defaultDomain: "kyma-domain.com"
    k8s:
      hpaSpec:
        maxReplicas: 5
        minReplicas: 2
      strategy:
        rollingUpdate:
          maxSurge: 50%
          maxUnavailable: "0"
      resources:
        limits:
          cpu: 250m
          memory: 384Mi
        requests:
          cpu: 10m
          memory: 128Mi
  disableKymaGateway: false
  gateways:
    - namespace: "some-ns" # Required
      name: "gateway1" # Required
      servers:
        - credentialName: "goat-certificate" # Required if Protocol = HTTPS
          hosts: # Creating  more than one for the same host:port configuration should result in  Warning
            - "goat.example.com"
            - "goat1.example.com"
          port:
            number: 443
            name: https
            protocol: HTTPS
        - hosts:
            - "goat.example.com"
            - "goat1.example.com"
          port:
            number: 80
            name: http
            protocol: HTTP
          httpsRedirect: true # If on Protocol = HTTPS, set Warning
        # We should consider configuration for MTLS gateway
  gardenCertificates: # Adding certificates in non-Gardener cluster should result in Warning/Error
    - namespace: "some-ns" # Required
      name: "goat-certificate" # Required
      commonName: "*.example.com" # Required
status:
  state: "Warning"
  description: "Cannot have same host on two gateways"
  conditions:
  - '[...]'

```
