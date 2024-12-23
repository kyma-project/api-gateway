# Expose Workloads in Multiple Namespaces With a Single APIRule Definition

This tutorial shows how to expose Service endpoints in multiple namespaces.

> [!WARNING]
>  Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).


##  Prerequisites

* You have deployed two workloads in separate namespaces.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: To learn what is the default domain of your Kyma cluster, run  `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

## Steps

1. Create a saprate namespace for the APIRule CR:

Expose the HTTPBin services in their respective namespaces by creating an APIRule custom resource (CR) in its own namespace. Run:

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

### Access Your Workloads
To access your HTTPBin Services, use [curl](https://curl.se).

To call the endpoints, send `GET` requests to the HTTPBin Services:

  ```bash
  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get
  ```
If successful, the calls return the `200 OK` response code.
