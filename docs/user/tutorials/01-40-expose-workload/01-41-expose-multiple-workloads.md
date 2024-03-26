# Expose Multiple Workloads on the Same Host

This tutorial shows how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING] Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, remember to secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* Deploy two instances of a sample [HTTPBin Service](../01-00-create-workload.md). Export their names as environment variables:
  
  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  ```

* Set up [your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 

## Define Multiple Services on Different Paths

Follow the instructions to expose the instances of the HTTPBin Service on different paths at the `spec.rules` level without a root Service defined.

<!-- tabs:start -->
#### **Kyma dashboard**

1. In the **Discovery and Network** section, select **APIRules**, and then **Create**. 
2. Provide the following configuration details:
    - **Name**: `multiple-service`
    - Depending on whether you're using your custom domain or a Kyma domain, follow the relevant instructions to fill in the `Gateway` section.
      <!-- tabs:start -->
      #### **Custom Domain**
      Select a `kyma-system` namespace and choose the gateway's name, for example `httpbin-gateway`. Use `httpbin.{KYMA_DOMAIN}` as a host, where `{KYMA_DOMAIN}` is the name of your Kyma domain.

      #### **Kyma Domain**
      Select the namespace in which you deployed an instance of the HTTPBin service and choose the gateway's name, for example `httpbin-gateway`. Use `httpbin.{CUSTOM_DOMAIN}` as a host, where `{CUSTOM_DOMAIN}` is the name of your custom domain.
      <!-- tabs:end -->
    
    - To expose the first service, add a Rule with the following configuration:
      - **Path**: `/headers`
      - Use the predefined access strategy with the no_auth handler and the `GET` method.
      - In the `Service` section, add specify the name and port of the first service you deployed.
    - To expose the second service, add a Rule with the following configuration:
      - **Path**: `/get`
      - Use the predefined access strategy with the `no_auth` handler and the `GET` method.
      - In the `Service` section, add specify the name and port of the second service you deployed.

      <!-- tabs:end -->

3. To create the APIRule, select `Create`.  
4. Replace the placeholder in the link and access the exposed HTTPBin Service at `https://httpbin.{YOUR_DOMAIN}`.

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

2. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: multiple-service
      namespace: $NAMESPACE
      labels:
        app: multiple-service
        example: multiple-service
    spec:
      host: multiple-service.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
      rules:
      - path: /headers
        methods: ["GET"]
        accessStrategies:
          - handler: no_auth
        service:
          name: $FIRST_SERVICE
          port: 8000
      - path: /get
        methods: ["GET"]
        accessStrategies:
          - handler: no_auth
        service:
          name: $SECOND_SERVICE
          port: 8000
    EOF
    ```

3. To call the endpoints, send `GET` requests to the HTTPBin Services:

    ```bash
    curl -ik -X GET https://multiple-service.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
    ```
    If successful, the calls return the code `200 OK` response.

<!-- tabs:end -->

## Define a Service at the Root Level

You can also define a Service at the root level. Such a definition is applied to all the paths specified at `spec.rules` that do not have their own Services defined.
 
> [!NOTE] 
>Services defined at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. In the **Discovery and Network** section, select **APIRules**, and then **Create**. Provide the following configuration details:
    - **Name**: `multiple-service`
    - In the `Service` section, select the name of the first service you deployed and the port `8000`. 
    - Depending on whether you're using your custom domain or a Kyma domain, follow the relevant instructions to fill in the `Gateway` section.
      <!-- tabs:start -->
      #### **Custom Domain**
      Select a `kyma-system` namespace and choose the gateway's name, for example `httpbin-gateway`. Use `httpbin.{KYMA_DOMAIN}` as a host, where `{KYMA_DOMAIN}` is the name of your Kyma domain.

      #### **Kyma Domain**
      Select the namespace in which you deployed an instance of the HTTPBin service and choose the gateway's name, for example `httpbin-gateway`. Use `httpbin.{CUSTOM_DOMAIN}` as a host, where `{CUSTOM_DOMAIN}` is the name of your Kyma domain.
      <!-- tabs:end -->
    
    - Add the first Rule with the following configuration:
      - **Path**: `/headers`
      - Use the predefined access strategy with the `no_auth` handler and the `GET` method.
      - Keep the `Service` section empty.
    - Add another Rule with the following configuration:
      - **Path**: `/get`
      - Use the predefined access strategy with the `no_auth` handler and the `GET` method.
      - In the `Service` section, add specify the name and port of the second service you deployed and use the port `8000`.

      <!-- tabs:end -->
  
    To create the APIRule, select `Create`.


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


2. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```shell
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: multiple-service
      namespace: $NAMESPACE
      labels:
        app: multiple-service
        example: multiple-service
    spec:
      host: multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      service:
        name: $FIRST_SERVICE
        port: 8000
      rules:
        - path: /headers
          methods: ["GET"]
          accessStrategies:
            - handler: no_auth
        - path: /get
          methods: ["GET"]
          accessStrategies:
            - handler: no_auth
          service:
            name: $SECOND_SERVICE
            port: 8000
    EOF
    ```

3. To call the endpoints, send `GET` requests to the HTTPBin Services:

    ```bash
    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
    ```
    If successful, the calls return the code `200 OK` response.

<!-- tabs:end -->