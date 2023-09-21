# Expose workloads in multiple Namespaces with a single APIRule definition

This tutorial shows how to expose Service endpoints in multiple Namespaces using APIGateway Controller.

   > **CAUTION:** Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).


##  Prerequisites

1. Create three Namespaces. Deploy two instances of the HttpBin Service, each in a separate Namespace. To learn how to do it, follow the [Create a workload](../01-00-create-workload.md) tutorial. Reserve the third Namespace for creating an APIRule.

  >**NOTE:** Remember to [enable the Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/02-operation-guides/operations/02-20-enable-sidecar-injection) in each Namespace.

1. Export the Namespaces' and Services' names as environment variables:

  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  export NAMESPACE_FIRST_SERVICE={NAMESPACE_NAME}
  export NAMESPACE_SECOND_SERVICE={NAMESPACE_NAME}
  export NAMESPACE_APIRULE={NAMESPACE_NAME}
  ```
  
3. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <div tabs name="export-values">

    <details>
    <summary>
    Custom domain
    </summary>
      
    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE_APIRULE/httpbin-gateway
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

## Expose and access your workloads in multiple Namespaces

1. Expose the HttpBin Services in their respective Namespaces by creating an APIRule CR in its own Namespace. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin-services
     namespace: $NAMESPACE_APIRULE
   spec:
     host: httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS
     gateway: $GATEWAY
     rules:
       - path: /headers
         methods: ["GET"]
         service:
           name: $FIRST_SERVICE
           namespace: $NAMESPACE_FIRST_SERVICE
           port: 8000
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
       - path: /get
         methods: ["GET"]
         service:
           name: $SECOND_SERVICE
           namespace: $NAMESPACE_SECOND_SERVICE
           port: 8000
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
   EOF
   ```

   >**NOTE:** If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

2. Call the HttpBin endpoints by sending a `GET` request to the HttpBin Services:

   ```bash
   curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```
   ```bash
   curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get
   ```

  If successful, the calls return the code `200 OK` response.