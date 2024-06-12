# Expose and Secure a Workload with JWT

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/), [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).
* [Obtain a JSON Web Token (JWT)](../01-51-get-jwt.md).
* [Set up your custom domain](../../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.

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

2. To expose and secure the Service, create the following APIRule:
    
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
        - jwt:
            authentications:
              -  issuer: $ISSUER
                 jwksUri: $JWKS_URI
          methods:
            - GET
          path: /.*
    EOF
    ```

### Access the Secured Resources

To access your HTTPBin Service, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

1. Try to access the secured workload without credentials.
    1. Enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. 
    2. To call the endpoint, send a `GET` request to the HTTPBin Service. 

    You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.
    1. Enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. 
    2. Go to the **Headers** tab. 
    3. Add a new header with the key **Authorization** and the value `Bearer {ACCESS_TOKEN}`. Replace `{ACCESS_TOKEN}` with your JWT.
    4. To call the endpoint, send a `GET` request to the HTTPBin Service. 

    If successful, you get the `200 OK` response code.


#### **curl**

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
<!-- tabs:end -->
