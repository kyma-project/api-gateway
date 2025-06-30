# Blocked In-Cluster Communication

## Symptom
After switching from APIRule `v1beta1` to version `v2`, in-cluster communication is blocked. Attempts to connect internally to the workload result the `error RBAC: access denied`.

## Cause

By default, the access to the workload from internal traffic is blocked if APIRule CR in versions `v2` or `v2alpha1` is applied.
This approach aligns with Kyma's "secure by default" principle.

## Solution

If an APIRule is applied, internal traffic is blocked by default. To allow it, you must create an ALLOW-type AuthorizationPolicy.

See the following example of an AuthorizationPolicy that allows internal traffic to the given workload. Note that it excludes traffic coming from `istio-ingressgateway` not to interfere with policies applied by APIRule to external traffic.

To use this code sample, replace `{NAMESPACE}`, `{LABEL_KEY}`, and `{LABEL_VALUE}` with the values appropriate for your environment.

| Option  | Description  |
|---|---|
|`{NAMESPACE}`   | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow internal traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy. |
|`{LABEL_KEY}`:`{LABEL_VALUE}`  | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workloads. Replace these placeholders with the actual key and value of the label.  The label indicates a specific set of Pods to which a policy should be applied. The scope of the label search is restricted to the configuration namespace in which the AuthorizationPolicy is applied. <br>For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/).|

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: {NAMESPACE}
spec:
  selector:
    matchLabels:
      {LABEL_KEY}: {LABEL_VALUE}
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

For example, to apply this policy to a workload labeled with `app: httpbin` deployed in the `default` namespace, you must set `{NAMESPACE}` to `default`, `{LABEL_KEY}` to `app`, and `{LABEL_VALUE}` to `httpbin`.

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