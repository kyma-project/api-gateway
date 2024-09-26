
# Expose and Secure a Workload with ExtAuth

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the 
APIRule custom resource (CR) and creates an Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/), 
[Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) with action type `CUSTOM`.

In this tutorial we will use a [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) with a OAuth2.0 complaint authorization server supporting oidc discovery 
to secure a workload using OAuth2.0 Client Credentials flow.

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).
* [Set up your custom domain](../../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.
* [Obtain a JSON Web Token (JWT)](../01-51-get-jwt.md).

## Steps

### Expose and Secure Your Workload

#### **kubectl**

1. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:

<!-- tabs:start -->
#### **Custom Domain**

```bash
export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
export GATEWAY=$NAMESPACE/httpbin-gateway
```
#### **Kyma Domain**

```bash
export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
export GATEWAY=kyma-system/kyma-gateway
```
<!-- tabs:end -->

2. Deploy `oauth2-proxy` with configuration for your authorization server.
This tutorial uses [oauth2-proxy helm chart](https://github.com/oauth2-proxy/manifests) for this purpose.

Export the following environment variables before deploying oauth2-proxy. 
You can generate the CLIENT and COOKIE secret with `openssl rand -base64 32 | head -c 32 | base64`.

You may want to adapt this configuration to better suit your needs.
Additional configuration parameters are listed [here](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview/#config-options).

```
export CLIENT_ID={CLIENT_ID/APPLICATION_AUDIENCE}
export CLIENT_SECRET={CLIENT_SECRET} # Generate with "openssl rand -base64 32 | head -c 32 | base64"
export COOKIE_SECRET={COOKIE_SECRET} # Generate with "openssl rand -base64 32 | head -c 32 | base64"
export OIDC_ISSUER_URL={OIDC_ISSUER_URL} # e.g "https://issuer.com"
export TOKEN_SCOPES={TOKEN_SCOPES} # e.g. "read write"
```

```
cat <<EOF > values.yaml
config:
  clientID: $CLIENT_ID
  clientSecret: $CLIENT_SECRET
  cookieName: ""
  cookieSecret: $COOKIE_SECRET

extraArgs: 
  auth-logging: true
  cookie-domain: "$DOMAIN_TO_EXPOSE_WORKLOADS"
  cookie-samesite: lax
  cookie-secure: false
  force-json-errors: true
  login-url: static://401
  oidc-issuer-url: $OIDC_ISSUER_URL
  pass-access-token: true
  pass-authorization-header: true
  pass-host-header: true 
  pass-user-headers: true
  provider: oidc
  request-logging: true
  reverse-proxy: true
  scope: "$TOKEN_SCOPES"
  set-authorization-header: true
  set-xauthrequest: true
  skip-jwt-bearer-tokens: true
  skip-oidc-discovery: false
  skip-provider-button: true
  standard-logging: true
  upstream: static://200
  whitelist-domain: "*.$DOMAIN_TO_EXPOSE_WORKLOADS:*"
EOF
```

Install oauth2-proxy with your configuration using Helm:

```
helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
helm upgrade --install oauth2-proxy oauth2-proxy/oauth2-proxy -f values.yaml -n oauth2-proxy
```

3. Register `oauth2-proxy` as authorization provider in Istio module:

```
kubectl patch istio -n kyma-system default --type merge --patch '{"spec":{"config":{"authorizers":[{"name":"oauth2-proxy","port":80,"service":"oauth2-proxy.oauth2-proxy.svc.cluster.local","headers":{"inCheck":{"include":["x-forwarded-for", "cookie", "authorization"]}}}]}}}'
```

4. To expose and secure the Service, create the following APIRule:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: httpbin
  namespace: $NAMESPACE
spec:
  hosts: 
    - httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: httpbin
    port: 8000
  gateway: $GATEWAY
  rules:
    - extAuth:
        authorizers:
          - oauth2-proxy
      methods:
        - GET
      path: /.*
EOF
```

### Access the Secured Resources

To access your HTTPBin Service use [curl](https://curl.se).

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

```bash
curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
```
You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.

```bash
curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers --header "Authorization:Bearer $ACCESS_TOKEN"
```
You get the `200 OK` response code.

