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

* [Deploy a sample HTTPBin Service](./01-00-create-workload.md).
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

### Expose Workloads Behind Your mTLS Gateway

To expose a custom workload, create an APIRule. You can either use version `v1beta1` or `v2alpha1`.

### Use APIRule `v1beta1`

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to the `default` namespace.
2. Go to **Discovery and Network > API Rules** and select **Create**.
3. Provide the following configuration details.
    - **Name**: `httpbin-mtls`
    - In the Gateway section, select:
      - **Namespace**: `default`
      - **Gateway**: `kyma-mtls-gateway`
    - Add the host `httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your custom domain.
    - In the `Rules` section, select:
      - **Path**: `/.*`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
      - Add the `httpbin` Service on port `8000`.

#### **kubectl**
Run the following command:

```bash
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
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  rules:
  - accessStrategies:
    - handler: no_auth
    methods:
    - GET
    path: /.*
  service:
    name: httpbin
    port: 8000
EOF
```
<!-- tabs:end -->

### Use APIRule `v2alpha1`

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  labels:
    app.kubernetes.io/name: httpbin-mtls
  name: httpbin-mtls
  namespace: default
spec:
  hosts:
    - httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  gateway: default/kyma-mtls-gateway
  rules:
    - path: /.*
      methods: ["GET"]
      noAuth: true
  service:
  name: httpbin
  port: 8000
EOF
```

This configuration uses the newly created Gateway `kyma-mtls-gateway` and exposes all workloads behind mTLS.

### Verify the Connection

<!-- tabs:start -->
#### **Postman**
Try to access the secured workload without credentials:

1. Enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/status/418`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. 
2. Send a `GET` request to the HTTPBin Service.

You get an SSL-related error.

Now, access the secured workload using the correct JWT:
1. Go to **Settings > Certificates** and select **Add Certificate**. Use your `cacert.crt` and `client.key` files.
2. Create a new request and enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/status/418`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. 
3. Go to the **Headers** tab and add the header:
    - **Content-Type**: `application/x-www-form-urlencoded`
4. To call the endpoint, send a `GET` request to the HTTPBin Service.

If successful, you get the code `418` response.


#### **curl**

1. Issue the curl command without providing the generated client certificate:
    ```bash
    curl -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/418
    ```
    You get:
    ```bash
    curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0
    ```

2. Provide the client and Root CA certificates in the command:
    ```bash
    curl --cert client.crt --key client.key --cacert cacert.crt -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/418
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
<!-- tabs:end -->
