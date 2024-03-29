# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Set up your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and select **Create API Rule**. 
2. Provide the following configuration details.
  - **Name**: `httpbin`
  - In the `Service` section, select:
    - **Service Name**: `httpbin`
    - **Port**: `8000`
  - Depending on whether you're using your custom domain or a Kyma domain, follow the respective instructions to fill in the `Gateway` section.
    <!-- tabs:start -->
    #### **Custom Domain**
    - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. 
    - **Name** is Gateway's name, for example `httpbin-gateway`. 
    - In the **Host** field, enter `httpbin.{YOUR_DOMAIN}`. Replace the placeholder with the name of your custom domain.

    #### **Kyma Domain**
    - **Namespace**: `kyma-system`
    - **Name** is the Gateway's name, for example `kyma-gateway`. 
    - In the **Host** field, enter `httpbin.{YOUR_DOMAIN}`. Replace the placeholder with the name of your Kyma domain.
    <!-- tabs:end -->
  - In the `Rules` section, select:
    - **Path**: `/.*`
    - **Handler**: `no_auth`
    - **Methods**: `GET` and `POST`
  
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

2. To expose an instance of the HTTPBin Service, create the following APIRule:

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
        name: $SERVICE_NAME
        namespace: $NAMESPACE
        port: 8000
      gateway: $GATEWAY
      rules:
        - path: /.*
          methods: ["GET"]
          accessStrategies:
            - handler: no_auth
        - path: /post
          methods: ["POST"]
          accessStrategies:
            - handler: no_auth
    EOF
    ```
  
    > [!NOTE]
    > If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file. 

    > [!NOTE]
    > If you don't specify a namespace for your Service, the default namespace is used.

<!-- tabs:end -->


To access your HTTPBin Service, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

1. Enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/ip` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. Call the endpoint by sending a `GET` request to the HTTPBin Service. If successful, the call returns the code `200 OK` response.

2. Enter the URL `https://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/post` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. Call the endpoint by sending a `POST` request to the HTTPBin Service. If successful, the call returns the code `200 OK` response.

#### **curl**

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
    ```
    If successful, the call returns the code `200 OK` response.

2. To call the endpoint, send a `POST` request to the HTTPBin Service.

    ```bash
    curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
    ```
    If successful, the call returns the code `200 OK` response.

<!-- tabs:end -->