# Set Up an mTLS Gateway and Expose Workloads Behind It

This document showcases how to set up an mTLS Gateway in Istio and expose it with an APIRule.

<!-- markdown-link-check-disable-next-line -->
According to the official [CloudFlare documentation](https://cloudflare.com/learning/access-management/what-is-mutual-tls/):
>Mutual TLS, or mTLS for short, is a method for mutual authentication. mTLS ensures that the parties at each end of a network connection are who they claim to be by verifying that they both have the correct private key. The information within their respective TLS certificates provides additional verification.

To establish a working mTLS connection, several things are required:

1. A working DNS entry pointing to the Istio Gateway IP
2. A valid Root CA certificate and key
3. Generated client and server certificates with a private key
4. Istio and API-Gateway installed on a Kubernetes cluster

The procedure of setting up a working mTLS Gateway is described in the following steps. The tutorial uses a Gardener shoot cluster and its API. The mTLS Gateway is exposed under your domain with a valid DNS `A` record.

## Prerequisites

* [Set up your custom domain](./01-10-setup-custom-domain-for-workload.md).

## Steps

### Set Up an mTLS Gateway

1. Create a DNS Entry and generate a wildcard certificate.

    > [!NOTE]
    > This step is heavily dependent on the configuration of a hyperscaler. Always consult the official documentation of each cloud service.

    For Gardener shoot clusters, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a self-signed Root CA and a client certificate.

    This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

    For a detailed step-by-step guide on how to generate a self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

<!-- tabs:start -->
#### **Kyma Dashboard**

3. Set up Istio Gateway in mutual mode. To do this, go to **Istio > Gateways** and select **Create**. Then, provide the following configuration details:
    - **Name**: `kyma-mtls-gateway`
    - Add a server with the following configuration:
      - **Port Number**: `443`
      - **Name**: `mtls`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `MUTUAL`
      - **Credential Name**: `kyma-mtls-certs`
      - Add a host `*.{DOMAIN_TO_EXPOSE_WORKLOADS}`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your custom domain.
    - Select **Create**.

    > [!NOTE]
    >  The `kyma-mtls-certs` Secret must contain a valid certificate for your custom domain.
4. Create a Secret containing the Root CA certificate.

    In order for the `MUTUAL` mode to work correctly, you must apply a Root CA in a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace.

    Go to **Configuration > Secrets** and select **Create**. Provide the following configuration details.
    - **Name**: `kyma-mtls-certs-cacert`
      - **Type**: `Opaque`
      - In the `Data` section, choose **Read value from file**. Select the file that contains your Root CA certificate.

#### **kubectl**
3. Export the name of your custom domain and the Gateway as environment variables. To set up Istio Gateway in mutual mode, apply the Gateway custom resource in a cluster.

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```

    > [!NOTE]
    >  The `kyma-mtls-certs` Secret must contain a valid certificate you created for your custom domain within the default namespace.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: kyma-mtls-gateway
      namespace: default
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
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
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace. 

    Run the following command:

    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs-cacert --from-file=cacert=cacert.crt
    ```
<!-- tabs:end -->

### Verify the Connection

To verify the connection, create and expose a sample workload using the cretied mTLS Gateway.

1. Call the endpoints without providing the generated client certificate:
    ```bash
    curl -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/status/418
    ```
    You get:
    ```bash
    curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0
    ```

2. Provide the client and Root CA certificates in the command:
    ```bash
    curl --cert client.crt --key client.key --cacert cacert.crt -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/status/418
    ```
    You get:
    ```bash
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