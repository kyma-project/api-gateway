# Expose workloads behind mTLS-enabled Gateway

This document showcases how to set up a mTLS Gateway in Istio and expose it with APIRule.

According to the official [CloudFlare documentation](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/):
>Mutual TLS, or mTLS for short, is a method for mutual authentication. mTLS ensures that the parties at each end of a network connection are who they claim to be by verifying that they both have the correct private key. The information within their respective TLS certificates provides additional verification.

To have a working mTLS connection, several things are required:

1. Working DNS entry pointing to the Istio Gateway IP
2. Valid Root CA certificate and key
3. Generated client and server certificates with a private key
4. Istio and API-Gateway installed on a Kubernetes cluster

The procedure of getting working mTLS Gateway are described in the following steps. For the needs of this tutorial, a Gardener shoot cluster and its API will be used.

The mTLS Gateway is exposed under `*.mtls.example.com` with valid DNS `A` record.

## Create DNS Entry and generate wildcard certificate

> Note: This step is heavily dependent on the configuration of a hyperscaler. Always consult official documentation of each cloud service.

For Gardener shoot clusters follow ["set up a custom domain for workload" guide](01-10-setup-custom-domain-for-workload.md).

## Generate self-signed Root CA and client certificate

This step is required for mTLS validation, so Istio could verify the authenticity of a client host.

For detailed step-by-step guide on how to generate self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

## Set up Istio Gateway in Mutual Mode

Assuming that the server certificate has been successfully created and is stored under `kyma-mtls-certs` secret in default namespace, modify and apply the following Gateway Custom Resource on a cluster:

```sh
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kyma-mtls-gateway
  namespace: default
spec:
  selector:
    app: istio-ingressgateway
    istio: ingressgateway # use istio default ingress gateway
  servers:
    - port:
        number: 443
        name: mtls
        protocol: HTTPS
      tls:
        mode: MUTUAL
        credentialName: kyma-mtls-certs
      hosts:
        - "*.mtls.example.com"
EOF
```

## Create secret containing Root CA cert

In order for `MUTUAL` mode working correctly, a Root CA must be applied on a cluster and follow [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) to be used by it.
Create an Opaque secret containing previously generated Root CA certificate:

```sh
    kubectl create secret generic -n default kyma-mtls-certs-cacert --from-file=cacert=cacert.crt
```

## Create custom workload and expose it using APIRule

To create a custom workload, follow ["create a workload" guide](01-00-create-workload.md).

Then expose the workload with an APIRule:

```sh
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  labels:
    app.kubernetes.io/name: httpbin-mtls
  name: httpbin-mtls
  namespace: default
spec:
  gateway: kyma-mtls-gateway.default.svc.cluster.local
  host: httpbin.mtls.example.com
  rules:
  - accessStrategies:
    - handler: allow
    methods:
    - GET
    path: /.*
  service:
    name: httpbin
    port: 80
```

This configuration uses newly created Gateway `kyma-mtls-gateway` and exposes all workloads behind mTLS.

## Verify the connection

Firstly, issue a `curl` command without providing a generated client certificate:
```
curl -X GET https://httpbin.mtls.example.com/status/418

curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0
```

Then issue a command with client and Root CA certificate provided to it:
```
curl --cert client.crt --key client.key --cacert cacert.crt -X GET https://httpbin.mtls.example.com/status/418


    -=[ teapot ]=-

       _...._
     .'  _ _ `.
    | ."` ^ `". _,
    \_;`"---"`|//
      |       ;/
      \_     _/
        `"""`
```

If the following commands return expected results, then the mTLS gateway has been successfully set up.