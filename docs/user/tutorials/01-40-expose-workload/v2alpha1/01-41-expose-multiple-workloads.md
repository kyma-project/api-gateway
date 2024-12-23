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
Follow the instructions to expose two workloads on different paths at the `spec.rules` level without a root Service defined. You can also define a Service at the root level. Such a definition is applied to all the paths specified at `spec.rules` that do not have their own Services defined.

> [!NOTE]
>Services defined at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

## Steps

- To define multiple services on different paths, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: {APIRULE_NAME}
      namespace: {APIRULE_NAMESPACE}
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      hosts:
        - multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
      - path: /headers
        methods: ["GET"]
        noAuth: true
        service:
          name: $FIRST_SERVICE
          port: 8000
      - path: /get
        methods: ["GET"]
        noAuth: true
        service:
          name: $SECOND_SERVICE
          port: 8000
    EOF
    ```
    
- To define a service at the root level, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      hosts:
        - multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      service:
        name: $FIRST_SERVICE
        port: 8000
      rules:
        - path: /headers
          methods: ["GET"]
          noAuth: true
        - path: /get
          methods: ["GET"]
          noAuth: true
          service:
            name: $SECOND_SERVICE
            port: 8000
    EOF
    ```

## Access Your Workloads

To call the endpoints, send `GET` requests to the exposed workloads:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers

  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/get
  ```
If successful, the calls return the `200 OK` response code.