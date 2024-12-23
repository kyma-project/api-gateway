# Expose Multiple Workloads on the Same Host

This tutorial shows how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING]
>  Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, remember to secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have deployed two workloads in one namespace.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: To learn what is the default domain of your Kyma cluster, run  `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

## Context
Learn how to expose two workloads on different paths at the `spec.rules` level without a root Service defined. You can also define a Service at the root level. Such a definition is applied to all the paths specified at `spec.rules` that do not have their own Services defined. Services defined at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and select **Create**.
2. Provide the following configuration details.
  - **Name**: `httpbin`
  - In the `Service` section, select:
    - **Service Name**: `httpbin`
    - **Port**: `8000`
  - To fill in the `Gateway` section, use these values:
    - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. If you use a Kyma domain, select the `kyma-system` namespace.
    - **Name** is the Gateway's name. If you use a Kyma domain, select `kyma-gateway`.
    - In the **Host** field, enter `httpbin.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
  - In the `Rules` section, add two Rules. Use the following configuration for the first one:
    - **Path**: `/.*`
    - **Handler**: `no_auth`
    - **Methods**: `GET`
  - Use the following configuration for the second Rule:
    - **Path**: `/post`
    - **Handler**: `no_auth`
    - **Methods**: `POST`

3. To create the APIRule, select **Create**.

#### **kubectl**


<!-- tabs:end -->

- To expose multiple services on different paths, create an APIRule CR and define each of your services on separate **spec.rules** level. See the example:

  ```bash
  ...
  spec:
    hosts:
      - example.your-domain.com
    gateway: 
    rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
      service:
        name: service1
        port: 8000
    - path: /get
      methods: ["GET"]
      noAuth: true
      service:
        name: service2
        port: 8000
  ```
    
- To expose a service at the root level, create an APIRule CR. Define one of your services on the **spec.service** and another on the **spec.rules** level. See the example:

  ```bash
  ...
  spec:
    hosts:
      - example.your-domain.com
    gateway: gateway-namespace/gateway-name
    service:
      name: service1
      port: 8000
    rules:
      - path: /headers
        methods: ["GET"]
        noAuth: true
      - path: /get
        methods: ["GET"]
        noAuth: true
        service:
          name: service2
          port: 8000
  ```