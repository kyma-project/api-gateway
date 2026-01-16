# Expose and Secure a Workload with a JWT Using SAP Cloud Identity Services

This procedure explains how to expose a workload on a custom domain and secure it with JSON Web Tokens (JWTs) issued by SAP Cloud Identity Services using the Client Credentials grant.

## Prerequisites

* You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud).
* You have an SAP Cloud Identity Services tenant. See [Initial Setup](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/initial-setup?locale=en-US&version=Cloud&q=open+id+connect).

## Context
Use this procedure to secure your workload with a short-lived JWT. To get the JWT, you must first register an OpenID Connect (OIDC) application in SAP Cloud Identity Services and enable the Client Credentials grant. This generates a client ID (public identifier) and a client secret (confidential credential). A calling system sends these credentials to the OIDC token endpoint over TLS, receiving a signed JWT.

When the client calls your exposed API, it includes the token in the Authorization header using the Bearer scheme. The API Gateway module validates the token based on the configuration you include in the APIRule custom resource (CR). If the validation fails, the Gateway returns `HTTP/2 403 RBAC: access denied` without forwarding the request to the backend Service.

If the validation is successful, the request proceeds to the Service behind the Gateway. At that point, you can implement optional, deeper authorization (examining scopes, audience, or custom claims) inside your application code.

## Configure a TLS Gateway for Your Custom Domain

To configure the flow in Kyma, you must first provide credentials for a supported DNS provider so Gardener can create and verify the necessary DNS records for your custom wildcard domain. After this, Let’s Encrypt issues a trusted TLS certificate. The issued certificate is stored in a Kubernetes Secret referenced by an Istio Gateway, which terminates HTTPS at the cluster edge, so all downstream traffic enters encrypted.

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as environment variables. Replace `my-own-domain.example.com` with the name of your domain. You can adjust these values as needed.

    ```bash
    PARENT_DOMAIN="my-own-domain.example.com"
    SUBDOMAIN="tls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    | Placeholder | Example domain name | Description |
    |---------|----------|---------|
    | **PARENT_DOMAIN** | `my-own-domain.example.com` | The domain name available in the public DNS zone. |
    | **SUBDOMAIN** | `tls.my-own-domain.example.com` | A subdomain created under the parent domain, specifically for the TLS Gateway. |
    | **GATEWAY_DOMAIN** | `*.tls.my-own-domain.example.com` | A wildcard domain covering all possible subdomains under the TLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.tls.my-own-domain.example.com`, `test.httpbin.tls.my-own-domain.example.com`) without creating separate Gateway rules for each one.|
    | **WORKLOAD_DOMAIN** | `httpbin.tls.my-own-domain.example.com` | The specific domain assigned to your workload. |

3. Create a Secret containing credentials for your DNS cloud service provider.

    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management).
    See an example Secret for the AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

    ```bash
    apiVersion: v1
    kind: Secret
    metadata:
      name: aws-credentials
      namespace: test
    type: Opaque
    data:
      AWS_ACCESS_KEY_ID: ...
      AWS_SECRET_ACCESS_KEY: ...
      # Optionally, specify the region
      #AWS_REGION: {YOUR_SECRET_ACCESS_KEY}
      # Optionally, specify the token
      #AWS_SESSION_TOKEN: ...
    ```

    To verify that the Secret is created, run:
   
    ```bash
    kubectl get secret -n test {SECRET_NAME}
    ```

4. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.

   See an example Secret for AWS Route 53 DNS provider:

    ```yaml
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws
      namespace: test
    spec:
      type: aws-route53
    secretRef:
        name: aws-credentials
      domains:
        include:
        - "${PARENT_DOMAIN}"
    ```

    To verify that the DNSProvider is created, run:
   
    ```bash
    kubectl get DNSProvider -n test {DNSPROVIDER_NAME}
    ```

5. Get the external access point of the `istio-ingressgateway` Service. The external access point is either stored in the ingress Gateway's **ip** field (for example, on GCP) or in the ingress Gateway's **hostname** field (for example, on AWS).

    ```bash
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [[ -z $LOAD_BALANCER_ADDRESS ]]; then
        LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    ```

6. Create a DNSEntry resource.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
      name: dns-entry
      namespace: test
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      dnsName: "${GATEWAY_DOMAIN}"
      ttl: 600
      targets:
        - "${LOAD_BALANCER_ADDRESS}"
    EOF
    ```

    To verify that the DNSEntry is created, run:
   
    ```bash
    kubectl get DNSEntry -n test dns-entry
    ```

7. Create the server's certificate.
    
   To request and manage Let's Encrypt certificates from your Kyma cluster, you use a Certificate CR. When you create a Certificate CR, one of Gardener's operators detects it and creates an [ACME](https://letsencrypt.org/how-it-works/) request to Let's Encrypt, requesting a certificate for the specified domain names. The issued certificate is stored in an automatically created Kubernetes Secret, whose name you specify in the Certificate's **secretName** field. For more information, see [Manage certificates with Gardener for public domain](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/request_cert/).

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: "istio-system"
    spec:
      secretName: custom-tls-secret
      commonName: "${GATEWAY_DOMAIN}"
      issuerRef:
        name: garden
    EOF
    ```
  
    To verify that the Secret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system custom-tls-secret
    ```

9.  Create a TLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: custom-tls-gateway
      namespace: test
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: tls
            protocol: HTTPS
          tls:
            mode: SIMPLE
            credentialName: custom-tls-secret
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
    
    To verify that the TLS Gateway is created, run:
   
    ```bash
    kubectl get gateway -n test custom-tls-gateway
    ```

### Create and Configure OpenID Connect Application
You need an identity provider to issue JWTs. Creating an OpenID Connect application allows SAP Cloud Identity Services to act as your issuer and manage authentication for your workloads.

1. Sign in to the administration console for SAP Cloud Identity Services. See [Access Admin Console](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/accessing-administration-console?locale=en-US&version=Cloud).

2. Create an OpenID Connect Application.

   1. Go to **Application Resources** > **Application**.
   2. Choose **Create**, provide the application name, and select the OpenID Connect radio button. 
      For more configuration options, see [Create OpenID Connect Application](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/create-openid-connect-application-299ae2f07a6646768cbc881c4d368dac?locale=en-US&version=Cloud).
   3. Choose **+Create**.

3. Configure OpenID Connect Application for the Client Credentials flow.
   
   1. In the **Trust > Single Sign-On** section of your created application, choose **OpenID Connect Configuration**.
   2. Provide the name.
   3. In the **Grant types** section, check **Client Credentials**.
      For more configuration options, see [Configure OpenID Connect Application for Client Credentials Flow](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/client-cred-configure-openid-connect-application-for-client-credentials-flow?locale=en-US&version=Cloud).
   4. Choose **Save**.

4. Configure a secret for API authentication.

   1. In the **Application API** section of your created application, choose **Client Authentication**.
   2. In the **Secrets** section, choose **Add**.
   3. Choose the OpenID API access and provide other configuration as needed.
      For more configuration options, see [Configure Secrets for API Authentication](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/dev-configure-secrets-for-api-authentication?version=Cloud&locale=en-US).
   4. Choose **Save**.
      Your client ID and secret appear in a pop-up window. Save the secret, as you will not be able to retrieve it from the system later.

### Get a JWT

1. Export the following values as environment variables:

    The name of your Cloud Identity Services instance in the URL of the administration console. For example, if your URL is `https://abc123.trial-accounts.ondemand.com/admin/`, the name of the instance is `abc123.trial-accounts.ondemand.com`.

    ```bash
    CLOUD_IDENTITY_SERVICES_INSTANCE="my-example-tenant.accounts.ondemand.com"
    CLIENT_ID="${YOUR-CLIENT-ID}"
    CLIENT_SECRET="${YOUR-CLIENT-SECRET}"
    ``` 

2. Export base 64 encoded client ID and client secret.
    
    ```bash
    export ENCODED_CREDENTIALS=$(echo -n "${CLIENT_ID}:${CLIENT_SECRET}" | base64)
    ```
3. Get **token_endpoint**, **jwks_uri**, and **issuer** from your OpenID application, and save these values as environment variables:

    ```bash
    TOKEN_ENDPOINT=$(curl -s https://${CLOUD_IDENTITY_SERVICES_INSTANCE}/.well-known/openid-configuration | jq -r '.token_endpoint')
    echo token_endpoint: ${TOKEN_ENDPOINT}
    JWKS_URI=$(curl -s https://${CLOUD_IDENTITY_SERVICES_INSTANCE}/.well-known/openid-configuration | jq -r '.jwks_uri')
    echo jwks_uri: ${JWKS_URI}
    ISSUER=$(curl -s https://${CLOUD_IDENTITY_SERVICES_INSTANCE}/.well-known/openid-configuration | jq -r '.issuer')
    echo issuer: ${ISSUER}
    ```
4. Get the JWT access token:

    ```bash
    ACCESS_TOKEN=$(curl -s -X POST "${TOKEN_ENDPOINT}" \
        -d "grant_type=client_credentials" \
        -d "client_id=${CLIENT_ID}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Authorization: Basic ${ENCODED_CREDENTIALS}" |  jq -r '.access_token')
    echo ${ACCESS_TOKEN}
    ```

### Configure JWT Authentication in Kyma

To configure JWT authentication, expose your workload using APIRule custom resource (CR). Configure **jwt** as the access strategy:
<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to **Discovery and Network > API Rules** and choose **Create**. 
2. Provide all the required configuration details.
3. Add a rule with the following configuration.
    - **Access Strategy**: `jwt`
    - In the `JWT` section, add an authentication with your issuer and JSON Web Key Set URIs.
    - Add allowed methods and the request path.
4. Choose **Create**.  

#### **kubectl**
To expose and secure your Service, create the APIRule custom resource. In the rules section, define the **jwt** field and specify the **issuer** and **jwksUri**.

```bash
...
  rules:
    - jwt:
        authentications:
          - issuer: ${ISSUER}
            jwksUri: ${JWKS_URI}
...
```
<!-- tabs:end -->
See the following example of a sample HTTPBin Deployment exposed by an APIRule with JWT authentication:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: httpbin
  namespace: test
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  namespace: test
  labels:
    app: httpbin
    service: httpbin
spec:
  ports:
  - name: http
    port: 8000
    targetPort: 80
  selector:
    app: httpbin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
  namespace: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
      version: v1
  template:
    metadata:
      labels:
        app: httpbin
        version: v1
    spec:
      serviceAccountName: httpbin
      containers:
      - image: docker.io/kennethreitz/httpbin
        imagePullPolicy: IfNotPresent
        name: httpbin
        ports:
        - containerPort: 80
```
```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin-tls
  namespace: test
spec:
  gateway: test/custom-tls-gateway
  hosts:
    - "${WORKLOAD_DOMAIN}"
  rules:
    - jwt:
        authentications:
          - issuer: ${ISSUER}
            jwksUri: ${JWKS_URI}
      methods:
        - GET
      path: /*
  service:
    name: httpbin
    namespace: test
    port: 8000
```

1. To test the connection, first, do not provide the JWT.
   
    ```bash
    curl -ik -X GET https://${WORKLOAD_DOMAIN}/headers
    ```
    You get the error `HTTP/2 403 RBAC: access denied`.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://${WORKLOAD_DOMAIN}/headers --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
    You get the `200 OK` response code.
