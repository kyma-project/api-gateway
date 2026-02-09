# Expose Multiple Workloads on the Same Host

APIRule allows you to expose multiple backend workloads under a single host by routing traffic to different Services based on path patterns. This capability enables you to consolidate multiple microservices behind a unified domain.

> [!WARNING]
>  Exposing a workload to the outside world is always a potential security vulnerability, so be careful. In a production environment, remember to secure the workload you expose with [JWT](./01-40-expose-workload-jwt.md).

There are two primary patterns for configuring multiple workloads on the same host: [Path-Level Service Definition](#path-level-service-definition) and [Root-Level Service with Overrides](#root-level-service-with-overrides).

## Path-Level Service Definition

In this pattern, each path explicitly declares which Service it routes to. You define all Service configurations at the `spec.rules` level, with no default Service defined at the root.

- Explicit routing - each path clearly states its target Service
- No default fallback behavior
- Maximum clarity and predictability
- Ideal when Services are unrelated or follow different architectural patterns

```yaml
spec:
  rules:
    - path: /headers
      service:
        name: service-a
    - path: /get
      service:
        name: service-b
```

## Root-Level Service with Overrides

This pattern defines a default Service at the root level (**spec.service**), which applies to all paths unless explicitly overridden. Individual rules can specify alternative Services that take precedence over the root definition.

- Default Service handles most traffic (with exceptions)
- Selective overrides for specific paths
- Reduces configuration verbosity when most paths use the same Service

> [!NOTE]
> When both root-level and rule-level Services are defined, the rule-level Service takes precedence.

```yaml
spec:
  service:
    name: primary-service
  rules:
    - path: /headers
    - path: /admin
      service:
        name: admin-service
```

## Cross-Namespace Service Exposure

When exposing Services, each Service referenced in the rules can be in a different namespace. To expose a Service deployed in a different namespace from the APIRule resource itself, specify its namespace in the field **service.namespace**.

```yaml
spec:
  service:
    name: primary-service
  rules:
    - path: /headers
      service:
        name: service-a
        namespace: ns-a
    - path: /get
      service:
        name: service-b
        namespace: ns-b
```

## Example

This example demonstrates all configuration patterns in a single APIRule, showing how to expose multiple workloads on one host.

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    kubectl create ns httpbin
    kubectl label namespace httpbin istio-injection=enabled --overwrite
    kubectl create ns nginx
    kubectl label namespace nginx istio-injection=enabled --overwrite
    ```

2. Get the default domain of your Kyma cluster.

    ```bash
    PARENT_DOMAIN=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.domain}")
    ```

3. Deploy a sample HTTPBin Service in the `httpbin` namespace.

    ```bash
    kubectl apply -n httpbin -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
    ```

4. Deploy a sample nginx Service in the `nginx` namespace.

    ```bash
    kubectl run nginx --image=nginx --port=80 -n nginx
    kubectl expose pod nginx --port=80 -n nginx
    ```

5. Create an APIRule:

    ```bash
    cat <<EOF | kubectl -n "${NAMESPACE}" apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: multi-workload
      namespace: httpbin
    spec:
      hosts:
        - api.${PARENT_DOMAIN}
      gateway: kyma-system/kyma-gateway
      service:
        name: httpbin
        namespace: httpbin
        port: 8000
      rules:
        - path: /headers    # Default: routes to httpbin service
          methods: ["GET"]
          noAuth: true       
        - path: /ip     # Override: routes to nginx service
          methods: ["GET"]
          noAuth: true
          service:
            name: nginx
            namespace: nginx
            port: 80
    EOF
    ```

5. Test the connection:

  - Test default service (httpbin)

    ```bash
    curl -ik https://api.${PARENT_DOMAIN}/headers
    ```
 
  - Test the override service (nginx)
    
    ```bash
    curl -ik https://api.${PARENT_DOMAIN}/ip
    ``` 

  If successful, each request returns a `200 OK` response from its respective service.
