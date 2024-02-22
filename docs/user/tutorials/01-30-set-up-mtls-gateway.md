# Set Up an mTLS Gateway and Expose Workloads Behind It

This document showcases how to set up an mTLS Gateway in Istio and expose it with an APIRule.
<!-- markdown-link-check-disable-next-line -->
According to the official [CloudFlare documentation](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/):
>Mutual TLS, or mTLS for short, is a method for mutual authentication. mTLS ensures that the parties at each end of a network connection are who they claim to be by verifying that they both have the correct private key. The information within their respective TLS certificates provides additional verification.

To establish a working mTLS connection, several things are required:

1. A working DNS entry pointing to the Istio Gateway IP
2. A valid Root CA certificate and key
3. Generated client and server certificates with a private key
4. Istio and API-Gateway installed on a Kubernetes cluster

The procedure of setting up a working mTLS Gateway is described in the following steps. The tutorial uses a Gardener shoot cluster and its API.

The mTLS Gateway is exposed under `*.mtls.example.com` with a valid DNS `A` record.

## Prerequisites

* Deploy [a sample HTTPBin Service](./01-00-create-workload.md).
* Set up [your custom domain](./01-10-setup-custom-domain-for-workload.md) and export the following values as environment variables:

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```

## Steps

1. Create a DNS Entry and generate a wildcard certificate.

    > [!NOTE] 
    > This step is heavily dependent on the configuration of a hyperscaler. Always consult the official documentation of each cloud service.

    For Gardener shoot clusters, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a self-signed Root CA and a client certificate.

    This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

    For a detailed step-by-step guide on how to generate a self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

3. Set up Istio Gateway in mutual mode.

    Assuming that you have successfully created the server certificate and it is stored in the `kyma-mtls-certs` Secret within the default namespace, modify and apply the following Gateway custom resource in a cluster:

    > [!NOTE]
    >  The `kyma-mtls-certs` Secret must contain a valid certificate for the `*.mtls.example.com` common name.

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
            - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
    EOF
    ```

4. Create a Secret containing the Root CA certificate.

    In order for the `MUTUAL` mode to work correctly, you must apply a Root CA in a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace:

    ```sh
        kubectl create secret generic -n istio-system kyma-mtls-certs-cacert --from-file=cacert=cacert.crt
    ```

5. Expose a custom workload using an APIRule.

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
      gateway: default/kyma-mtls-gateway
      host: httpbin.mtls.example.com
      rules:
      - accessStrategies:
        - handler: no_auth
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
    ```
    ```
    curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0
    ```

    Then, provide the client and Root CA certificates in the command:
    ```
    curl --cert client.crt --key client.key --cacert cacert.crt -X GET https://httpbin.mtls.example.com/status/418
    ```
    ```

        -=[ teapot ]=-

          _...._
        .'  _ _ `.
        | ."` ^ `". _,
        \_;`"---"`|//
          |       ;/
          \_     _/
            `"""`
    ```

    If the commands return the expected results, you have set up the mTLS Gateway successfully.