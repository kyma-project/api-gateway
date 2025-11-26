# Expose and Secure a Workload with exAuth Using the Authorization Code Flow
Learn how to expose and secure a Kyma workload using OAuth2 Proxy external authorizer and the OAuth 2.0 Authorization Code flow.

## Prerequisites

- You have an SAP BTP, Kyma runtime instance with the Istio and API Gateway modules added. The Istio and API Gateway modules are added to your Kyma cluster by default.
- You have an SAP Cloud Identity Services tenant. See [Initial Setup](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/initial-setup?locale=en-US&version=Cloud&q=open+id+connect).

## Context

Follow these steps:
1. [Create and Configure an OpenID Connect Application](#create-and-configure-an-openid-connect-application)
2. [Deploy OAuth2 Proxy as an External Authorizer](#deploy-oauth2-proxy-as-an-external-authorizer)
3. [Expose Your Workload Using APIRule with extAuth](#expose-your-workload-using-apirule-with-extauth)

## Procedure

### Create and Configure an OpenID Connect Application

oauth2-proxy needs a trusted OIDC client registered with the IdP. The redirect URI is where the IdP sends the user back after a successful login.

1. Sign in to the administration console for SAP Cloud Identity Services. See [Access Admin Console](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/accessing-administration-console?locale=en-US&version=Cloud).

2. Create an OpenID Connect Application.

   1. Go to **Application Resources** > **Application**.
   2. Choose **Create**, provide the application name, and select the OpenID Connect radio button. 
      For more configuration options, see [Create OpenID Connect Application](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/create-openid-connect-application-1a87534329494d48ae5f246c4842e11a?locale=en-US&version=Cloud).
   3. Choose **+Create**.

3. Configure OpenID Connect Application for the Authorization Code flow.
   
   1. In the **Trust > Single Sign-On** section of your created application, choose **OpenID Connect Configuration**.
   2. Provide the name.
   3. Add the `https://oauth2-proxy.{YOUR_DOMAIN}/oauth2/callback` redirect URI. 
      The redirect URI is where the IdP sends the user back after a successful login. In this case, replace `{YOUR_DOMAIN}` with the name of the host on which you expose your service in Kyma.
   3. In the **Grant types** section, check **Authorization Code**.
      For more configuration options, see [Configure OpenID Connect Application for Authorization Code Flow](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/auth-code-configure-openid-connect-application-for-authorization-code-flow?locale=en-US&version=Cloud).
   4. Choose **Save**.

4. Configure a secret for API authentication.

   1. In the **Application API** section of your created application, choose **Client Authentication**.
   2. In the **Secrets** section, choose **Add**.
   3. Choose the OpenID API access and provide other configuration as needed.
      For more configuration options, see [Configure Secrets for API Authentication](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/dev-configure-secrets-for-api-authentication?version=Cloud&locale=en-US).
   4. Choose **Save**.
      Your client ID and secret appear in a pop-up window. Save the secret, as you cannot retrieve it again after closing this window.

### Deploy OAuth2 Proxy as an External Authorizer
OAuth2 Proxy handles the OAuth2/OIDC Authorization Code flow. It redirects unauthenticated users to Cloud Identity Services, processes the callback, sets session cookies, and can pass user info and tokens to your workload via headers.

1. Export the following values as environment variables:

    ```bash
    TENANT_URL="https://my-example-tenant.accounts.ondemand.com"
    CLIENT_ID="{YOUR-CLIENT-ID}"
    CLIENT_SECRET="{YOUR-CLIENT-SECRET}"
    EXPOSE_DOMAIN="{YOUR-DOMAIN}"
    ``` 

2. Create a namespace for deploying the OAuth2 Proxy.
    
    ```bash
    kubectl create ns oauth2-proxy
    kubectl label namespace oauth2-proxy istio-injection=enabled --overwrite
    ```

3. Add the [oauth2-proxy helm chart](https://github.com/oauth2-proxy/manifests) to your local Helm registry.

    ```bash
    helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
    helm repo update oauth2-proxy
    ```

3. Define the `oauth2-proxy` configuration for your authorization server and deploy it to your Kyma cluster.
    You can adjust this configuration as needed. See the [additional configuration parameters](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview/#config-options).

    ```bash
    helm upgrade --install oauth2-proxy --namespace oauth2-proxy oauth2-proxy/oauth2-proxy \
    --set config.clientID="$CLIENT_ID" \
    --set config.clientSecret="$CLIENT_SECRET" \
    --set config.cookieSecret="$(openssl rand -base64 32 | head -c 32 | base64)" \
    --set extraArgs.provider=oidc \
    --set extraArgs.cookie-secure="false" \
    --set extraArgs.cookie-domain="$EXPOSE_DOMAIN" \
    --set extraArgs.cookie-samesite="lax" \
    --set extraArgs.cookie-refresh="1h" \
    --set extraArgs.cookie-expire="4h" \
    --set extraArgs.set-xauthrequest="true" \
    --set extraArgs.whitelist-domain="*.$EXPOSE_DOMAIN:*" \
    --set extraArgs.reverse-proxy="true" \
    --set extraArgs.pass-access-token="true" \
    --set extraArgs.set-authorization-header="true" \
    --set extraArgs.pass-authorization-header="true" \
    --set extraArgs.pass-user-headers="true" \
    --set extraArgs.pass-host-header="true" \
    --set extraArgs.scope="openid email" \
    --set extraArgs.upstream="static://200" \
    --set extraArgs.skip-provider-button="true" \
    --set extraArgs.redirect-url="https://oauth2-proxy.$EXPOSE_DOMAIN/oauth2/callback" \
    --set extraArgs.skip-oidc-discovery="false" \
    --set extraArgs.oidc-issuer-url="$TENANT_URL" \
    --set extraArgs.standard-logging="true" \
    --set extraArgs.auth-logging="true" \
    --set extraArgs.request-logging="true" \
    --set extraArgs.code-challenge-method="S256" \
    --wait
    ```
    
    Check if the Oauth2 Proxy's Pods are running:
    
    ```bash
    kubectl --namespace=oauth2-proxy get pods -l "app=oauth2-proxy"
    ```

4. Deploy APIRule exposing OAuth2 Proxy. 
    Exposing the OAuth2 Proxy is required, as its endpoints must be accessible from the outside for the Authorization Code flow (required for access to redirects)

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: oauth2-proxy
      namespace: oauth2-proxy
    spec:
      gateway: ${GATEWAY:-"kyma-system/kyma-gateway"}
      hosts:
        - oauth2-proxy.$EXPOSE_DOMAIN
      service:
        name: oauth2-proxy
        port: 80
      rules:
        - path: /*
          methods: ["GET", "POST", "PATCH", "DELETE", "OPTIONS"]
          noAuth: true
    EOF
    ```

    Check if the APIRule's status is ready:
    
    ```bash
    kubectl --namespace=oauth2-proxy get apirules oauth2-proxy
    ```

2. Allow internal traffic to the oauth2proxy (required for ext-auth request to reach the provider) to ensure that tokens coming from Authorization Code flow will not authorize this endpoint.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
      name: oauth2-proxy-allow-internal
      namespace: oauth2-proxy
    spec:
      rules:
      - from:
        - source:
            principals:
            - cluster.local/ns/*
        to:
        - operation:
            notPaths:
            - /ping
            - /ready
      selector:
        matchLabels:
          app.kubernetes.io/instance: oauth2-proxy
          app.kubernetes.io/name: oauth2-proxy
    EOF
    ```

3. Configure the OAuth2 Proxy external authorizer in the Istio custom resource.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: operator.kyma-project.io/v1alpha2
    kind: Istio
    metadata:
      name: default
      namespace: kyma-system
    spec:
      config:
        authorizers:
        - name: oauth2-proxy
          service: "oauth2-proxy.oauth2-proxy.svc.cluster.local"
          port: 80
          pathPrefix: "/verify?original="
          headers:
            inCheck:
              include:
                - "authorization"
                - "cookie"
                - "x-forwarded-for"
              add:
                # Sets uri to redirect user after login
                x-forwarded-uri: "https://%REQ(:authority)%%REQ(x-envoy-original-path?:path)%"
            toUpstream:
              onAllow:
              - "authorization"
              - "cookie"
              - "path"
              - "x-forwarded-id-token"
              - "x-forwarded-access-token"
              - "x-forwarded-refresh-token"
              - "x-forwarded-for"
              - "x-forwarded-user"
              - "x-forwarded-groups"
              - "x-forwarded-email"
              - "x-forwarded-preferred-username"
            toDownstream:
              onDeny:
              - "content-type"
              - "set-cookie"
              onAllow:
              - "x-forwarded-id-token"
              - "x-forwarded-access-token"
              - "x-forwarded-refresh-token"
              - "x-forwarded-for"
              - "x-forwarded-user"
              - "x-forwarded-groups"
              - "x-forwarded-email"
              - "x-forwarded-preferred-username"
    EOF
    ```

### Expose Your Workload Using APIRule with extAuth 

To configure OAuth2 Proxy, expose your workload using APIRule custom resource (CR). Configure **extAuth** as the access strategy:

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to **Discovery and Network > API Rules** and choose **Create**. 
2. Provide all the required configuration details.
3. Add a rule with the following configuration.
    - **Access Strategy**: `extAuth`
    - Add allowed methods and the request path.
4. Choose **Create**.  

#### **kubectl**
To expose and secure your Service, create the APIRule custom resource. In the rules section, define the **jwt** field and specify the **issuer** and **jwksUri**.

```bash
...
  rules:
    - path: /*
      methods: ["GET"]
      extAuth:
        authorizers:
          - oauth2-proxy
...
```
<!-- tabs:end -->

See the following example of a sample HTTPBin Deployment exposed by an APIRule with extAuth authorizer:

1. Create the `httpbin-system` namespace and deploy a sample HTTPBin Deployment.

    ```bash
    kubectl create ns httpbin-system
    kubectl label namespace httpbin-system istio-injection=enabled
    kubectl apply -f \
    https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml \
    -n httpbin-system-2
    ```

2. Expose the workload with an APIRule using the extAuth access strategy. 

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin
      namespace: httpbin-system
    spec:
      gateway: ${GATEWAY:-"kyma-system/kyma-gateway"}
      hosts:
        - httpbin.$EXPOSE_DOMAIN
      service:
        name: httpbin
        port: 8000
      rules:
        - path: /*
          methods: ["GET"]
          extAuth:
            authorizers:
              - oauth2-proxy
    EOF
    ```
    
    Check if the APIRule's status is ready:
    
    ```bash
    kubectl --namespace=oauth2-proxy get apirules httpbin
    ```

#### Results
To access your workload go to `httpbin.$EXPOSE_DOMAIN`. You're redirected to Cloud Identity Services first where you need to sign in. If the log in is successful, you can access your application (in the sample scenario, you see HTTPBin endpoints).