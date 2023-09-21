# Expose multiple workloads on the same host

This tutorial shows how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, remember to secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* Deploy two instances of a sample [HttpBin Service](../01-00-create-workload.md). Export their names as environment variables:
  
  ```bash
    export FIRST_SERVICE={SERVICE_NAME}
    export SECOND_SERVICE={SERVICE_NAME}
  ```

* Set up [your custom domain](../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <div tabs name="export-values">

    <details>
    <summary>
    Custom domain
    </summary>
    
    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```
    </details>

    <details>
    <summary>
    Kyma domain
    </summary>

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
    export GATEWAY=kyma-system/kyma-gateway
    ```
    </details>
  </div>   

## Define multiple Services on different paths

Follow the instructions to expose the instances of the HttpBin Service on different paths at the `spec.rules` level without a root Service defined.

1. To expose the instances of the HttpBin Service, create an APIRule CR in your Namespace. Run:

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
     host: multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS
     gateway: $GATEWAY
     rules:
     rules:
      - path: /headers
        methods: ["GET"]
        accessStrategies:
          - handler: noop
        service:
          name: $FIRST_SERVICE
          port: 8000
      - path: /get
        methods: ["GET"]
        accessStrategies:
          - handler: noop
        service:
          name: $SECOND_SERVICE
          port: 8000
   EOF
   ```

2. To call the endpoints, send `GET` requests to the HttpBin Services:

    ```bash
    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
    ```
  If successful, the calls return the code `200 OK` response.

## Define a Service at the root level

You can also define a Service at the root level. Such a definition is applied to all the paths specified at the `spec.rules` which do not have their own Services defined.
 
 > **NOTE:** Services definitions at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

1. To expose the instances of the HttpBin Service, create an APIRule CR in your Namespace. Run:

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
          - handler: noop
      - path: /get
        methods: ["GET"]
        accessStrategies:
          - handler: noop
        service:
          name: $SECOND_SERVICE
          port: 8000
  EOF
  ```
  In the above APIRule, the HttpBin Service on port 8000 is defined at the `spec.service` level. This Service definition is applied to the `/headers` path. The `/get` path has the service definition overwritten.

1. To call the endpoints, send `GET` requests to the HttpBin Services:

    ```bash
    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
    ```
  If successful, the calls return the code `200 OK` response.