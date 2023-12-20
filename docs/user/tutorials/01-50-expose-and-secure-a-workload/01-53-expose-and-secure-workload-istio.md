# Expose and Secure a Workload with Istio

This tutorial shows how to expose and secure a workload using Istio's built-in security features. You will expose the workload by creating a [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/). Then, you will secure access to your workload by adding the JWT validation verified by the Istio security configuration with [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/).

## Prerequisites

* Deploy a [sample HttpBin Service](../01-00-create-workload.md).
* [JSON Web Token (JWT)](./01-51-get-jwt.md).
* Set up [your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
<!-- tabs:start -->
  #### Custom Domain
    
  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
  #### Kyma Domain

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=kyma-system/kyma-gateway
  ```
<!-- tabs:end -->  

## Expose Your Workload Using a VirtualService

Follow the instructions in the tabs to expose the HttpBin workload using a VirtualService.

1. Create a VirtualService:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     hosts:
     - "httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS"
     gateways:
     - $GATEWAY
     http:
     - match:
       - uri:
           prefix: /
       route:
       - destination:
           port:
             number: 8000
           host: httpbin.$NAMESPACE.svc.cluster.local
   EOF
   ```

## Secure a Workload Using a JWT

To secure the HttpBin workload using a JWT, create a Request Authentication with Authorization Policy. Workloads with the `matchLabels` parameter specified require a JWT for all requests. Follow the instructions:

1. Create the Request Authentication and Authorization Policy resources:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: RequestAuthentication
   metadata:
     name: jwt-auth-httpbin
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: httpbin
     jwtRules:
     - issuer: $ISSUER
       jwksUri: $JWKS_URI
   ---
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: httpbin
     rules:
     - from:
       - source:
           requestPrincipals: ["*"]
   EOF
   ```

2. Access the workload you secured. You get the code `403 Forbidden` error.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
   ```

3. Now, access the secured workload using the correct JWT. You get the code `200 OK` response.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
   ```
