# Blocked In-Cluster Communication

## Symptom
After switching from APIRule `v1beta1` to version `v2`, in-cluster communication is blocked.

## Cause

By default, the access to the workload from internal traffic is blocked if APIRule CR in versions `v2` or `v2alpha1` is applied.
This approach aligns with Kyma's "secure by default" principle.

## Solution

If an APIRule is applied, internal traffic is blocked by default. To allow it, you must create an ALLOW-type AuthorizationPolicy.

See the following example of an AuthorizationPolicy that allows internal traffic to the given workload. Note that it excludes traffic coming from `istio-ingressgateway` not to interfere with policies applied by APIRule to external traffic.

To use this code sample, replace `{NAMESPACE}`, `{KEY}`, and `{TARGET_WORKLOAD}` with the values appropriate for your environment.

| Option  | Description  |
|---|---|
|`{NAMESPACE}`   | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow internal traffic. |
|`{KEY}`:`{TARGET_WORKLOAD}`  | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workloads. Replace these placeholders with the actual key and value of the label. For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/).  |

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: {NAMESPACE}
spec:
  selector:
    matchLabels:
      {KEY}: {TARGET_WORKLOAD}
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

For example, to apply this policy to a workload labeled with `app: httpbin` deployed in the `test` namespace, you must set `{NAMESPACE}` to `default`, `{KEY}` to `app`, and `{TARGET_WORKLOAD}` to `httpbin`.

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: default
spec:
  selector:
    matchLabels:
      app: httpbin
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```