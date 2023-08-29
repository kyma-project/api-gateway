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
        - certificate: "goat-certificate" # If not defined, generate Gardener certificate
          DNSEntry: "goat-dns" # If not defined, generate Gardener DNSEntry
          hosts: # Creating  more than one for the same host:port configuration should result in  Warning
            - "goat.example.com"
            - "goat1.example.com"
          port:
            number: 443
            name: https
            protocol: HTTPS
        - hosts:
          DNSEntry: "goat-dns" # If not defined, generate Gardener DNSEntry
            - "goat.example.com"
            - "goat1.example.com"
          port:
            number: 80
            name: http
            protocol: HTTP
          httpsRedirect: true # If on Protocol = HTTPS, set Warning
        # We should consider configuration for MTLS gateway
    #gardenCertificates: # Adding certificates in non-Gardener cluster should result in Warning/Error
    #- namespace: "some-ns" # Required
    #  name: "goat-certificate" # Required
    #  commonName: "*.example.com" # Required
    #gardenDNSEntries: # Adding DNSEntries in non-Gardener cluster should result in Warning/Error
    #- commonName: "*.example.com" # Required
status:
  state: "Warning"
  description: "Cannot have same host on two gateways"
  conditions:
  - '[...]' # array of *metav1.Condition

```
