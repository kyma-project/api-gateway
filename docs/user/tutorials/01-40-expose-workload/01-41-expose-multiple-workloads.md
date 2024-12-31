# Expose Multiple Workloads on the Same Host

This tutorial shows how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING] Exposing a workload to the outside world is always a potential security vulnerability, so be careful. In a production environment, remember to secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Deploy two instances of a sample HTTPBin Service](../01-00-create-workload.md) in one namespace. 
* [Set Up Your Custom Domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

## Define Multiple Services on Different Paths

Follow the instructions to expose the instances of the HTTPBin Service on different paths at the `spec.rules` level without a root Service defined.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > APIRules** and select **Create**. 
2. Provide the following configuration details:
    - **Name**: `multiple-services`
    - Add a Gateway with the following values:
      - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. If you use a Kyma domain, select the `kyma-system` namespace.
      - **Name** is the Gateway's name. If you use a Kyma domain, select `kyma-gateway`. 
      - In the **Host** field, enter `httpbin.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
    - To expose the first Service, add a rule with the following configuration:
      - **Path**: `/headers`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
      - In the `Service` section, select the name of the first Service you deployed and use port `8000`.
    - To expose the second Service, add a rule with the following configuration:
      - **Path**: `/get`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
      - In the `Service` section, select the name of the second Service you deployed and use port `8000`.
      <!-- tabs:end -->

3. To create the APIRule, choose **Create**.  


#### **kubectl**
1. Export the names of two deployed HTTPBin Services as environment variables:
  
  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  ```

2. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
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

3. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      host: multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
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

<!-- tabs:end -->

## Define a Service at the Root Level

You can also define a Service at the root level. Such a definition is applied to all the paths specified at `spec.rules` that do not have their own Services defined.
 
> [!NOTE] 
>Services defined at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > APIRules** and select **Create**. 
2. Provide the following configuration details:
    - **Name**: `httpbin-services`
    - In the `Service` section, select the name of the first service you deployed and port `8000`. 
    - To fill in the `Gateway` section, use these values:
      - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. If you use a Kyma domain, select the `kyma-system` namespace.
      - **Name** is the Gateway's name. If you use a Kyma domain, select `kyma-gateway`. 
      - In the **Host** field, enter `httpbin.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
    - Add a rule with the following configuration:
      - **Path**: `/headers`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
      - Leave the `Service` section empty.
    - Add another rule with the following configuration:
      - **Path**: `/get`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
      - In the `Service` section, select the name of the second Service you deployed and use port `8000`.
  
3. To create the APIRule, select **Create**.

#### **kubectl**

1. Export the names of the two deployed HTTPBin Services as environment variables:
  
  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  ```

2. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
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


3. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      host: multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
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
<!-- tabs:end -->

## Access Your Workloads
To access your HTTPBin Services, use [curl](https://curl.se).

To call the endpoints, send `GET` requests to the HTTPBin Services:

  ```bash
  curl -ik -X GET https://multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

  curl -ik -X GET https://multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
  ```
If successful, the calls return the `200 OK` response code.