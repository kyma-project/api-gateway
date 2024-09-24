
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

2. Deploy `oauth2-proxy` with configuration for your authorization server. This tutorial uses [oauth2-proxy helm chart](https://github.com/oauth2-proxy/manifests) 
for this purpose. Sample configuration is provided in [this directory](../v2alpha1/resources/oauth2-proxy-helm).

To make sure that `oauth2-proxy` works with your provider, you need to at least adapt the following configuration:

```yaml
config:
  clientID: "<your application aud claim>"
  # Create a new secret with the following command
  # openssl rand -base64 32 | head -c 32 | base64
  clientSecret: "random-secret"

extraArgs: 
  provider: oidc
  whitelist-domain: "*.<your domain>:*"
  oidc-issuer-url: "<your jwt token issuer>"
```

You may want to adapt this configuration to better suit your needs.
Additional configuration parameters are listed [here](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview/#config-options).

3. To expose and secure the Service, create the following APIRule:

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

