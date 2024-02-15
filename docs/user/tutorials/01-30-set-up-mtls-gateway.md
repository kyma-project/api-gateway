# Set Up an mTLS Gateway and Expose Workloads Behind It

This document showcases how to set up a mTLS Gateway in Istio and expose it with APIRule.

According to the official [CloudFlare documentation](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/):
>Mutual TLS, or mTLS for short, is a method for mutual authentication. mTLS ensures that the parties at each end of a network connection are who they claim to be by verifying that they both have the correct private key. The information within their respective TLS certificates provides additional verification.

To establish a working mTLS connection, several things are required:

1. A working DNS entry pointing to the Istio Gateway IP
2. A valid Root CA certificate and key
3. Generated client and server certificates with a private key
4. Istio and API-Gateway installed on a Kubernetes cluster

The procedure of setting up a working mTLS Gateway is described in the following steps. The tutorial uses a Gardener shoot cluster and its API.

The mTLS Gateway is exposed under `*.mtls.example.com` with a valid DNS `A` record.

## Create DNS Entry and generate wildcard certificate

> Note: This step is heavily dependent on the configuration of a hyperscaler. Always consult the official documentation of each cloud service.

For Gardener shoot clusters, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a self-signed Root CA and a client certificate.

This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

For detailed step-by-step guide on how to generate self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

3. Set up Istio Gateway in mutual mode.

Assuming that you have successfully created the server certificate and it is stored in the `kyma-mtls-certs` Secret within the default namespace, modify and apply the following Gateway custom resource on a cluster:

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

4. Create a Secret containing the Root CA certificate.

In order for the `MUTUAL` mode to work correctly, you must apply a Root CA on a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
Create an Opaque Secret containing the previously generated Root CA certificate:

```sh
    kubectl create secret generic -n default kyma-mtls-certs-cacert --from-file=cacert=cacert.crt
```

5. Create a custom workload and expose it using an APIRule.

To create a custom workload, follow [Create a Workload](01-00-create-workload.md).

Then expose the workload with the APIRule:

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

This configuration uses the newly created Gateway `kyma-mtls-gateway` and exposes all workloads behind mTLS.

6. Verify the connection.

Firstly, issue a curl command without providing the generated client certificate:
```
curl -X GET https://httpbin.mtls.example.com/status/418

curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0
```

Then, provide the client and Root CA certificates in the command:
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

If the following commands return the expected results, you have set up the mTLS gateway successfully.