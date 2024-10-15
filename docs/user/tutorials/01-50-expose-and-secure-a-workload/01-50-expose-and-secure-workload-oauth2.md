# Expose and Secure a Workload with OAuth2

This tutorial shows how to expose and secure Services using APIGateway Controller. The controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR.

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Set up your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.
* Configure your client ID and client Secret using an OAuth2-compliant provider.

## Steps

### Get the Tokens

1. Encode the client's credentials and export them as an environment variable:

    ```bash
    export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
    export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
    ```
2. Export your token endpoint as an environment variable:

    ```bash
    export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
    ```

3. Get a token with the `read` scope.

    1. Get the opaque token:
        ```shell
        curl --location --request POST "$TOKEN_ENDPOINT?grant_type=client_credentials" -F "scope=read" --header "Content-Type: application/x-www-form-urlencoded" --header "Authorization: Basic $ENCODED_CREDENTIALS"
        ```
    2. Export the issued token as an environment variable:
        ```bash
        export ACCESS_TOKEN_READ={ISSUED_READ_TOKEN}
        ```
4. Get a token with the `write` scope.

    1. Get the opaque token:
        ```shell
        curl --location --request POST "$TOKEN_ENDPOINT?grant_type=client_credentials" -F "scope=write" --header "Content-Type: application/x-www-form-urlencoded" --header "Authorization: Basic $ENCODED_CREDENTIALS"
        ```
    2. Export the issued token as an environment variable:
        ```shell
        export ACCESS_TOKEN_WRITE={ISSUED_WRITE_TOKEN}
        ```

### Expose and Secure Your Workload
<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and select **Create**.
2. Provide the following configuration details:
    - **Name**: `httpbin`
    - **Service Name**: `httpbin`
    - **Port**: `8000`
    - To fill in the `Gateway` section, use these values:
      - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. With a Kyma domain, use the `kyma-system` namespace.
      - **Name** is the Gateway's name, for example `httpbin-gateway`.
      - In the **Host** field, enter `httpbin.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
    - Add an access strategy with the following configuration:
      - **Handler**: `oauth2_introspection`
      - **Introspection URL** is your introspection endpoint.
      - Add introspection request header with the key **Authorization** and value `Basic {ENCODED_CREDENTIALS}`. Replace with base 64 encoded `{CLIENT_ID}:{CLIENT_SECRET}`.
      - **Required Scope**: `read`
      - **Method**: `GET`
      - **Path**: `/.*`
    - Add another access strategy with the following configuration:
      - **Handler**: `oauth2_introspection`
      - **Introspection URL** is your introspection endpoint.
      - Add introspection request header with the key **Authorization** and value `Basic {ENCODED_CREDENTIALS}`. Replace with base 64 encoded `{CLIENT_ID}:{CLIENT_SECRET}`.
      - **Required Scope**: `write`
      - **Method**: `POST`
      - **Path**: `/post`

3. To create the APIRule, select **Create**.


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

2. Export your introspection endpoint as an environment variable:

    ```bash
      export INTROSPECTION_ENDPOINT={INTROSPECTION_URL}
    ```

3. To expose an instance of the HTTPBin Service and secure it with OAuth2 scopes, create the following APIRule in your namespace:

    ```bash
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
                introspection_url: "$INTROSPECTION_ENDPOINT"
                introspection_request_headers:
                  Authorization: "Basic $ENCODED_CREDENTIALS"
        - path: /post
          methods: ["POST"]
          accessStrategies:
            - handler: oauth2_introspection
              config:
                required_scope: ["write"]
                introspection_url: "$INTROSPECTION_ENDPOINT"
                introspection_request_headers:
                  Authorization: "Basic $ENCODED_CREDENTIALS"
    EOF
    ```
<!-- tabs:end -->

> [!NOTE]
>  If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

The exposed Service requires tokens with `read` scope for `GET` requests in the entire Service, and tokens with `write` scope for `POST` requests to the `/post` endpoint of the Service.

> [!WARNING]
>  When you secure a workload, don't create overlapping Access Rules for paths. Doing so can cause unexpected behavior and reduce the security of your implementation.

### Access the Secured Resources

Follow the instructions to call the secured Service using the tokens issued for the client you registered.

1. Send a `GET` request with a token that has the `read` scope to the HTTPBin Service:

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN_READ"
    ```

2. Send a `POST` request with a token that has the `write` scope to the HTTPBin's `/post` endpoint:

    ```bash
    curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data" -H "Authorization: Bearer $ACCESS_TOKEN_WRITE"
    ```

If successful, the calls return the code `200 OK` responses. If you call the Service without a token, you get the code `401 Unauthorized` response. If you call the Service or its secured endpoint with a token with the wrong scope, you get the code `403 Forbidden` response.

To learn more about the security options, read the document describing [authorization configuration](../../custom-resources/apirule/04-50-apirule-authorizations.md).
