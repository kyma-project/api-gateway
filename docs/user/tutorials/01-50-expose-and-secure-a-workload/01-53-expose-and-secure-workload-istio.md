# Expose and Secure a Workload with Istio

This tutorial shows how to expose and secure a workload using Istio's built-in security features. You will expose the workload by creating a [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/). Then, you will secure access to your workload by adding the JWT validation verified by the Istio security configuration with [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/).

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Obtain a JSON Web Token (JWT)](./01-51-get-jwt.md).
* [Set up your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.

## Steps

1. Expose the HTTPBin Service instance.

<!-- tabs:start -->
  #### **Kyma dashboard**

  1. Go to **Istio > Virtual Services** and select **Create Virtual Service**. 
  2. Switch to the `Advanced` tab and provide the following configuration details:
      - **Name**: `httpbin`
      - Go to **HTTP > Matches > Match** and provide URI of the type **prefix** and value `/`.
      - Go to **HTTP > Routes > Route > Destination**. Replace `{NAMESPACE}` with the name of the HTTPBin Service's namespace and add the following fields:
        - **Host**: `httpbin.{NAMESPACE}.svc.cluster.local`
        - **Port Number**: `8000`
  3. To create the VirtualService, select **Create**.

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

  2. To expose your workload, create a VirtualService:

      ```bash
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
<!-- tabs:end --> 

3. Secure a Workload Using a JWT

    To secure the HTTPBin workload using a JWT, create a Request Authentication with Authorization Policy. Workloads with the `matchLabels` parameter specified require a JWT for all requests. Follow the instructions:

<!-- tabs:start -->
    #### **Kyma Dashboard**
    1. Go to **Custom Resources > RequestAuthentications**.
    2. Select **Create RequestAuthentication** and paste the following configuration into the editor:
        ```yaml
        apiVersion: security.istio.io/v1beta1
        kind: RequestAuthentication
        metadata:
          name: jwt-auth-httpbin
          namespace: {NAMESPACE}
        spec:
          selector:
            matchLabels:
              app: httpbin
          jwtRules:
          - issuer: {ISSUER}
            jwksUri: {JWKS_URI}
        ```
    3. Replace the placeholders:
      - `{NAMESPACE}` is the name of the namespace in which you deployed the HTTPBin Service.
      - `{ISSUER}` is
      - `{JWKS_URI}` is 
    3. Select **Create**.
    4. Go to **Istio > Authorization Policies**.
    5. Select **Create Authorization Policy**, switch to the `YAML` tab and paste the following configuration into the editor:
        ```bash
        apiVersion: security.istio.io/v1beta1
          kind: AuthorizationPolicy
          metadata:
            name: httpbin
            namespace: {NAMESPACE}
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
    6. Replace `{NAMESPACE}` with the name of the namespace in which you deployed the HTTPBin Service.
    7. Select **Create**.
    8. verify

    #### **kubectl**

    Create the Request Authentication and Authorization Policy resources:

    ```bash
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
<!-- tabs:end -->

3. Access the workload you secured. You get the code `403 Forbidden` error.

    ```shell
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
    ```

4. Now, access the secured workload using the correct JWT. You get the code `200 OK` response.

    ```shell
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
<!-- tabs:start -->