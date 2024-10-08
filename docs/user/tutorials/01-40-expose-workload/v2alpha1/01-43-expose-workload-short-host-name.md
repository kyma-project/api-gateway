# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service using short host name without domain and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).

## Steps

### Expose Your Workload

#### **kubectl**

1. Export the necessary values as environment variables:

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```

2. Create custom Gateway specifying host domain:

<!-- tabs:start -->
#### **Kyma Dashboard**

* Go to **Istio > Gateways** and select **Create**.
* Provide the following configuration details:
    - **Name**: `httpbin-gateway`
    - **Namespace**: `{NAMESPACE_NAME}`
    - In the `Servers` section, select **Add**. Then, use these values:
      - **Port Number**: `80`
      - **Name**: `http`
      - **Protocol**: `HTTP`
    - Use `httpbin.{KYMA_DOMAIN_NAME}` as **Host**.

* Select **Create**.

#### **kubectl**

* To create a custom Gateway, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1beta1
    kind: Gateway
    metadata:
      name: httpbin-gateway
      namespace: $NAMESPACE
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - hosts:
            - '*.$DOMAIN_TO_EXPOSE_WORKLOADS'
          port:
            name: http
            number: 80
            protocol: HTTP
    EOF
    ```

<!-- tabs:end -->

3. To expose an instance of the HTTPBin Service, create the following APIRule referring created custom Gateway:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
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

> [!NOTE]
> If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

> [!NOTE]
> If you don't specify a namespace for your Service, the default namespace is used.

### Access Your Workload

To access your HTTPBin Service, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

- Enter the URL `http://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/ip` and replace the placeholder with the name of your domain. Call the endpoint by sending a `GET` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

- Enter the URL `http://httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}/post` and replace the placeholder with the name of your domain. Call the endpoint by sending a `POST` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

#### **curl**

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET http://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST http://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.

<!-- tabs:end -->
