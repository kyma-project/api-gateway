# Expose and Secure a Workload with JWT

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Obtain a JSON Web Token (JWT)](./01-51-get-jwt.md).
* [Set up your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.

## Steps

<!-- tabs:start -->
#### **Kyma dashboard**

1. In the **Discovery and Network** section, select **APIRules**, and then **Create**. 
2. Provide the following configuration details:
    - **Name**: `httpbin`
    - **Service Name**: `httpbin`
    - **Port**: `8000`
    - Depending on whether you're using your custom domain or a Kyma domain, follow the relevant instructions to fill in the `Gateway` section.
      <!-- tabs:start -->
      #### **Custom Domain**
      Select a `kyma-system` namespace and choose the gateway's name, for example `kyma-gateway`. Use `httpbin.{KYMA_DOMAIN}` as a host, where `{KYMA_DOMAIN}` is the name of your Kyma domain.

      #### **Kyma Domain**
      Select the namespace in which you deployed an instance of the HTTPBin service and choose the gateway's name, for example `httpbin-gateway`. Enter the name of your custom domain in the **Host** field.
      <!-- tabs:end -->
    - Add an access strategy with the following configuration:
      - **Handler**: `jwt`
      - In the `jwks_uri` section, add your JSON Web Key Set URIs.
      - **Method**: `GET`
      - **Path**: `/.*`

3. To create the APIRule, select `Create`.  
4. Replace the placeholder in the link and access the secured HTTPBin Service at `https://httpbin.{YOUR_DOMAIN}`.

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

2. To expose and secure the Service, create the follwing APIRule:
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS   
      service:
        name: httpbin
        port: 8000
      gateway: $GATEWAY
      rules:
        - accessStrategies:
          - handler: jwt
            config:
              jwks_urls:
              - $JWKS_URI
          methods:
            - GET
          path: /.*
    EOF
    ```

3. To access the secured Service, call it using the JWT access token:

    ```bash
    curl -ik https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN"
    ```

    If successful, the call returns the code `200 OK` response.

<!-- tabs:end -->