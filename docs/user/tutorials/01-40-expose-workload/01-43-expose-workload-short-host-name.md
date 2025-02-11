# Expose a Workload with Short Host Name

This tutorial demonstrates how to expose an unsecured instance of the HTTPBin Service using a short host name instead of the full domain name. Using a short host makes it simpler to apply APIRules because the domain name is automatically retrieved from the referenced Gateway, and you don’t have to manually set it in each APIRule. This might be particularly useful when reconfiguring resources in a new cluster, as it reduces the chance of errors and streamlines the process.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).
* [Set up your custom domain](../../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.

## Steps

### Expose Your Workload

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
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      hosts:
        - httpbin
      service:
        name: $SERVICE_NAME
        namespace: $NAMESPACE
        port: 8000
      gateway: $GATEWAY
      rules:
        - path: /*
          methods: ["GET"]
          noAuth: true
        - path: /post
          methods: ["POST"]
          noAuth: true
    EOF
    ```
    
    The host domain name will be obtained from the referenced Gateway.
> [!NOTE]
> If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

> [!NOTE]
> If you don't specify a namespace for your Service, the default namespace is used.

### Access Your Workload

To access your HTTPBin Service, use [curl](https://curl.se).

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.
