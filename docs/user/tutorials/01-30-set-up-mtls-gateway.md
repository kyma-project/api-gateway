# Set Up an mTLS Gateway and Expose Workloads Behind It

Learn how to set up an mTLS Gateway in Istio and use it to expose a workload.

## Context

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
    > How to perform this step heavily depends on the configuration of a hyperscaler. Always consult the official documentation of each cloud service provider.

    For Gardener shoot clusters, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a self-signed Root CA and a client certificate.

    This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

    For a detailed step-by-step guide on how to generate a self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

<!-- tabs:start -->
#### **Kyma Dashboard**

3. Set up Istio Gateway in mutual mode. 
    1. Go to **Istio > Gateways** and choose **Create**. 
    2. Add the name `kyma-mtls-gateway`.
    3. Add a server with the following configuration:
      - **Port Number**: `443`
      - **Name**: `mtls`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `MUTUAL`
      - **Credential Name**: `kyma-mtls-certs`
      - Add a host `*.{DOMAIN_NAME}`. Replace `{DOMAIN_NAME}` with the name of your custom domain.
    4. Choose **Create**.

    > [!NOTE]
    >  The `kyma-mtls-certs` Secret must contain a valid certificate for your custom domain.

4. Create a Secret containing the Root CA certificate.

    In order for the `MUTUAL` mode to work correctly, you must apply a Root CA in a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace.

    1. Go to **Configuration > Secrets** and choose **Create**. 
    2. Provide the following configuration details:
      - **Name**: `kyma-mtls-certs-cacert`
      - **Type**: `Opaque`
      - In the `Data` section, choose **Read value from file**. Select the file that contains your Root CA certificate.

#### **kubectl**
3. To set up Istio Gateway in mutual mode, apply the Gateway custom resource.

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
            - "*.{DOMAIN_NAME}"
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

## Expose Workloads Behind Your mTLS Gateway

To expose a custom workload, create an APIRule. You can either use version `v2`, `v2alpha1`, or `v1beta1`.

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed on May 12, 2025. Version `v2alpha1`, introduced for testing purposes, will become deprecated after the stable APIRule `v2` is promoted to the regular channel. The promotion of the APIRule `v2` to the regular channel has been postponed. We will keep you posted on the coming dates and changes.
>  
> To migrate your APIRule CRs to version `v2`, follow the prcedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same.

### Use APIRule `v2`

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to the `default` namespace.
2. Go to **Discovery and Network > API Rules** and choose **Create**.
3. Add the name `apirule-mtls`.
4. Add a Gateway with the following values:
   - **Namespace**: `default`
   - **Gateway**: `kyma-mtls-gateway`
5. Add the host `{SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}`.
6. Add a rule with the following values:
   - **Path**: `/.*`
   - **Handler**: `noAuth`
   - **Methods**: `GET`
   - Select the name and port of your Service.

#### **kubectl**
Run the following command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  labels:
    app.kubernetes.io/name: apirule-mtls
  name: apirule-mtls
  namespace: default
spec:
  hosts:
    - {SUBDOMAIN}.{DOMAIN_NAME}
  gateway: default/kyma-mtls-gateway
  rules:
    - path: /.*
      methods: ["GET"]
      noAuth: true
  service:
  name: {SERVICE_NAME}
  port: {SERVICE_PORT}
EOF
```
<!-- tabs:end -->

### Use APIRule `v2alpha1`

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to the `default` namespace.
2. Go to **Discovery and Network > API Rules v2alpha1** and choose **Create**.
3. Add the name `apirule-mtls`.
4. Add a Gateway with the following values:
   - **Namespace**: `default`
   - **Gateway**: `kyma-mtls-gateway`
5. Add the host `{SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}`.
6. Add a rule with the following values:
   - **Path**: `/.*`
   - **Handler**: `noAuth`
   - **Methods**: `GET`
   - Select the name and port of your Service.

#### **kubectl**
Run the following command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  labels:
    app.kubernetes.io/name: apirule-mtls
  name: apirule-mtls
  namespace: default
spec:
  hosts:
    - {SUBDOMAIN}.{DOMAIN_NAME}
  gateway: default/kyma-mtls-gateway
  rules:
    - path: /.*
      methods: ["GET"]
      noAuth: true
  service:
  name: {SERVICE_NAME}
  port: {SERVICE_PORT}
EOF
```
<!-- tabs:end -->

### Use APIRule `v1beta1`

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to the `default` namespace.
2. Go to **Discovery and Network > API Rules** and select **Create**.
3. Add the name `apirule-mtls`.
4. Add a Gateway with the following values:
   - **Namespace**: `default`
   - **Gateway**: `kyma-mtls-gateway`
5. Add the host `{SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}`.
6. Add a rule with the following values:
   - **Path**: `/.*`
   - **Handler**: `no_auth`
   - **Methods**: `GET`
   - Select the name and port of your Service.

#### **kubectl**
Run the following command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  labels:
    app.kubernetes.io/name: apirule-mtls
  name: apirule-mtls
  namespace: default
spec:
  gateway: default/kyma-mtls-gateway
  host: {SUBDOMAIN}.{DOMAIN_NAME}
  rules:
  - accessStrategies:
    - handler: no_auth
    methods:
    - GET
    path: /.*
  service:
    name: {SERVICE_NAME}
    port: {SERVICE_PORT}
EOF
```
<!-- tabs:end -->

### Verify the Connection

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