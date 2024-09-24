# Expose Workloads in Multiple Namespaces With a Single APIRule Definition

This tutorial shows how to expose Service endpoints in multiple namespaces using APIGateway Controller.

> [!WARNING]
>  Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).


##  Prerequisites

Create three namespaces. Deploy two instances of the HTTPBin Service, each in a separate namespace. To learn how to do it, follow the [Create a Workload](../../01-00-create-workload.md) tutorial. Reserve the third namespace for creating an APIRule.

  > [!NOTE]
  > Remember to [Enable Automatic Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/operation-guides/02-20-enable-sidecar-injection) in each namespace.

## Steps

### Expose Your Workloads

#### **kubectl**

1. Export the namespaces' and Services' names as environment variables:

    ```bash
    export FIRST_SERVICE={SERVICE_NAME}
    export SECOND_SERVICE={SERVICE_NAME}
    export NAMESPACE_FIRST_SERVICE={NAMESPACE_NAME}
    export NAMESPACE_SECOND_SERVICE={NAMESPACE_NAME}
    export NAMESPACE_APIRULE={NAMESPACE_NAME}
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

3. Expose the HTTPBin Services in their respective namespaces by creating an APIRule custom resource (CR) in its own namespace. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: httpbin-services
      namespace: $NAMESPACE_APIRULE
    spec:
      hosts:
        - httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
        - path: /headers
          methods: ["GET"]
          service:
            name: $FIRST_SERVICE
            namespace: $NAMESPACE_FIRST_SERVICE
            port: 8000
          noAuth: true
        - path: /get
          methods: ["GET"]
          service:
            name: $SECOND_SERVICE
            namespace: $NAMESPACE_SECOND_SERVICE
            port: 8000
          noAuth: true
    EOF
    ```

    > [!NOTE]
    > If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

### Access Your Workloads
To access your HTTPBin Services, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

- Enter the URL `https://httpbin-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with your domain name. To call the endpoint, send a `GET` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

- Enter the URL `https://httpbin-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/get` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with your domain name. To call the endpoint, send a `GET` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

#### **curl**

To call the endpoints, send `GET` requests to the HTTPBin Services:

  ```bash
  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get
  ```
If successful, the calls return the `200 OK` response code.

<!-- tabs:end -->
