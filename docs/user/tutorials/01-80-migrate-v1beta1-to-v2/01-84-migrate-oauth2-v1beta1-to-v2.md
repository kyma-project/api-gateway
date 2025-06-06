# Migrate APIRule `v1beta1` of type oauth2_introspection to version `v2`

This tutorial explains how to migrate an APIRule created with version `v1beta1` using the **oauth2_introspection** handler, as it is the most popular ORY Oathkeeper-based handler, to version `v2` with the **extAuth** handler.

## Context

APIRule version `v1beta1` is deprecated and scheduled for removal. Once the APIRule custom resource definition (CRD) stops serving version `v1beta1`, the API server will no longer respond to requests for APIRules in this version. Consequently, you will not be able to create, update, delete, or view APIRules in `v1beta1`. Therefore, migrating to version `v2` is required.

## Prerequisites

* You have a deployed workload with the Istio and API Gateway modules enabled.
* To use the CLI instructions, you must have [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/) installed.
* You have obtained the configuration of the APIRule in version `v1beta1` to be migrated. See [Retrieve the **spec** of APIRule in version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).
* The workload exposed by the APIRule in version `v2` must be part of the Istio service mesh.
todo: jwt do przjesci a ias 
## Steps

> [!NOTE] In this example, the APIRule `v1beta1` was created with the **oauth2_introspection** handler, so the migration targets an APIRule `v2` using the **extAuth** handler. To illustrate the migration, the HTTPBin service is used, exposing the `/anything` endpoint. The HTTPBin service is deployed in its own namespace, with Istio enabled, ensuring the workload is part of the Istio service mesh.

1. Obtain a configuration of the APIRule in version `v1beta1` and save it for future modifying. For instructions, see [Retrieve the **spec** of APIRule in version `v1beta1`](./01-81-retrieve-v1beta1-spec.md). Below is a sample of the retrieved **spec** in YAML format for an APIRule in `v1beta1`:
```yaml
host: httpbin.example.com
service:
  name: httpbin
  namespace: test
  port: 8000
gateway: kyma-system/kyma-gateway
rules:
  - path: /anything
    methods:
      - GET
    accessStrategies:
      - handler: oauth2_introspection
        config:
          introspection_request_headers:
            Authorization: 'Basic '
          introspection_url: https://{IAS_TENANT}.accounts.ondemand.com/oauth2/introspect
          required_scope:
            - read
```
Above configuration uses the **auth2_introspection** handler to expose HTTPBin `/anything` endpoint.

2. In order for the `extAuth` handler in APIRule `v2` to work you must deploy a service that will act as external authorizer for Istio. This tutorial uses [`oauth2-proxy`](https://oauth2-proxy.github.io/oauth2-proxy/) with an OAuth2.0 complaint authorization server supporting OIDC discovery with following configuration:
```yaml
cat <<EOF > values.yaml
config:
  clientID: {CLIENT_ID}
  clientSecret: {CLIENT_SECRET}
  cookieName: ""
  cookieSecret: {COOKIE_SECRET}

extraArgs:
  auth-logging: true
  cookie-domain: "{DOMAIN_TO_EXPOSE_WORKLOADS}"
  cookie-samesite: lax
  cookie-secure: false
  force-json-errors: true
  login-url: static://401
  oidc-issuer-url: {OIDC_ISSUER_URL}
  pass-access-token: true
  pass-authorization-header: true
  pass-host-header: true
  pass-user-headers: true
  provider: oidc
  request-logging: true
  reverse-proxy: true
  scope: "{TOKEN_SCOPES}"
  set-authorization-header: true
  set-xauthrequest: true
  skip-jwt-bearer-tokens: true
  skip-oidc-discovery: false
  skip-provider-button: true
  standard-logging: true
  upstream: static://200
  whitelist-domain: "*.{DOMAIN_TO_EXPOSE_WORKLOADS}:*"
EOF
```

3. This is just a subset of configuration. Refer to oauth2-proxy documentation to view all the options. Once you have the configuration prepared, install oauth2-proxy using the following commands.
```bash
kubectl create namespace oauth2-proxy
helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
helm upgrade --install oauth2-proxy oauth2-proxy/oauth2-proxy -f values.yaml -n oauth2-proxy
```
4. Register `oauth2-proxy` as an authorization provider in the Istio module:
```bash
kubectl patch istio -n kyma-system default --type merge --patch '{"spec":{"config":{"authorizers":[{"name":"oauth2-proxy","port":80,"service":"oauth2-proxy.oauth2-proxy.svc.cluster.local","headers":{"inCheck":{"include":["x-forwarded-for", "cookie", "authorization"]}}}]}}}'

```
5. Adjust configuration of the APIRule to version `v2` with the **extAuth** handler. Below is an example of the adjusted APIRule configuration for version `v2`:

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin
  namespace: test
spec:
  hosts:
    - httpbin
  service:
    name: httpbin
    namespace: test
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - extAuth:
        authorizers:
          - oauth2-proxy
      methods:
        - GET
      path: /anything
```
> [!NOTE]
> Notice that the **hosts** field can accept a short host name (without a domain). For more information about the changes introduced in APIRule `v2`, see the [APIRule v2 Changes](../../custom-resources/apirule/04-70-changes-in-apirule-v2.md) document. **Read this document before applying the new APIRule in `v2`.**

The above example of APIRule delegates checking a token to previously configured oauth2-proxy. Active tokens will continue to work during migration, and the migration process will not disrupt any exposed or secured workloads.

6. Update the APIRule to version `v2` by applying the adjusted configuration. To verify the version of the applied APIRule, check the value of the `gateway.kyma-project.io/original-version` annotation in the APIRule spec. A value of `v2` indicates that the APIRule has been successfully migrated. You can use the following command:
```bash 
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
```
```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  annotations:
    gateway.kyma-project.io/original-version: v2
...
```
Above APIRule has been successfully migrated to version `v2`.
> [!WARNING] Do not manually change the `gateway.kyma-project.io/original-version` annotation. This annotation is automatically updated when you apply your APIRule in version `v2`.

7. To preserve the internal traffic policy from the APIRule `v1beta1`, apply the following AuthorizationPolicy. Make sure to update the selector label so that it matches the target workload:
```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: ${NAMESPACE}
spec:
  selector:
    matchLabels:
      ${LABEL_KEY}: ${LABEL_VALUE} 
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

8. Additionally, to retain the CORS configuration from the APIRule `v1beta1`, update the APIRule in version `v2` to include the same CORS settings. For preflight requests work correctly, you must explicitly add the `"OPTIONS"` method to the **rules.methods** field of your APIRule `v2`. For guidance, refer to the available [APIRUle `v2` samples](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/04-10-apirule-custom-resource?id=sample-custom-resource).

### Access Your Workload todo czy to nie ma byc lekko zmienione?

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip --header "Authorization:Bearer $ACCESS_TOKEN"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything -d "test data" --header "Authorization:Bearer $ACCESS_TOKEN"
  ```
  If successful, the call returns the `200 OK` response code.


