# Expose and Secure a Workload with OAuth2

This tutorial shows how to expose and secure Services using APIGateway Controller. The controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR.

## Prerequisites

* Deploy [a sample HTTPBin Service](../01-00-create-workload.md).
* Set up [your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
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

* Configure your client ID and client Secret using an OAuth2-compliant provider. Then, export the following values as environment variables:
  
  ```shell
    export CLIENT_ID={CLIENT_ID}
    export CLIENT_SECRET={CLIENT_SECRET}
    export TOKEN_URL={TOKEN_URL}
    export INTROSPECTION_URL={INTROSPECTION_URL}
   ```

## Get the Tokens

1. Encode the client's credentials and export them as environment variables:
   
    ```shell
    export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
    ```

2. Get tokens to interact with secured resources using the client credentials flow:

    <!-- tabs:start -->
    #### **Token with `read` scope**
  
    * Export the following value as an environment variable:
      ```shell
      export KYMA_DOMAIN={KYMA_DOMAIN_NAME}
      ```  
    * Get the opaque token:
      ```shell
      curl --location --request POST "$TOKEN_URL?grant_type=client_credentials" --header "Content-Type: application/x-www-form-urlencoded" --header "Authorization: Basic $ENCODED_CREDENTIALS"
      ```
    * Export the issued token as an environment variable:
      ```shell
      export ACCESS_TOKEN_READ={ISSUED_READ_TOKEN}
      ```
    #### **Token with `write` scope**
  
    * Export the following value as an environment variable:
      ```shell
      export KYMA_DOMAIN={KYMA_DOMAIN_NAME}
      ```  
    * Get the opaque token:
      ```shell
      curl --location --request POST "$TOKEN_URL?grant_type=client_credentials" --header "Content-Type: application/x-www-form-urlencoded" --header "Authorization: Basic $ENCODED_CREDENTIALS"
      ```
    * Export the issued token as an environment variable:
      ```shell
      export ACCESS_TOKEN_WRITE={ISSUED_WRITE_TOKEN}
      ```
    <!-- tabs:end -->


## Expose and Secure Your Workload

Expose an instance of the HTTPBin Service, and secure it with OAuth2 scopes by creating an APIRule CR in your namespace. Run:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: httpbin
  namespace: $NAMESPACE
spec:
  gateway: $GATEWAY
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: $SERVICE_NAME
    port: 8000
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
            introspection_url: "$INTROSPECTION_URL"
            introspection_request_headers:
              Authorization: "Basic $ENCODED_CREDENTIALS"
    - path: /post
      methods: ["POST"]
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["write"]
            introspection_url: "$INTROSPECTION_URL"
            introspection_request_headers:
              Authorization: "Basic $ENCODED_CREDENTIALS"
EOF
```

> [!NOTE]
>  If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

The exposed Service requires tokens with `read` scope for `GET` requests in the entire Service, and tokens with `write` scope for `POST` requests to the `/post` endpoint of the Service.

  
> [!WARNING]
>  When you secure a workload, don't create overlapping Access Rules for paths. Doing so can cause unexpected behavior and reduce the security of your implementation.

## Access the Secured Resources

Follow the instructions to call the secured Service using the tokens issued for the client you registered.

1. Send a `GET` request with a token that has the `read` scope to the HTTPBin service:

    ```shell
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN_READ"
    ```

2. Send a `POST` request with a token that has the `write` scope to the HTTPBin's `/post` endpoint:

    ```shell
    curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data" -H "Authorization: bearer $ACCESS_TOKEN_WRITE"
    ```

If successful, the call returns the code `200 OK` response. If you call the Service without a token, you get the code `401` response. If you call the Service or its secured endpoint with a token with the wrong scope, you get the code `403` response.

To learn more about the security options, read the document describing [authorization configuration](../../custom-resources/apirule/04-50-apirule-authorizations.md).
