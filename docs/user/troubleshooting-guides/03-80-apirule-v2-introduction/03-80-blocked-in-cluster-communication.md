# Blocked In-Cluster Communication

## Symptom
After switching from APIRule `v1beta1` to version `v2`, in-cluster communication is blocked.

## Cause

By default, the access to the workload from internal traffic is blocked if APIRule CR in versions `v2` or `v2alpha1` is applied.
This approach aligns with Kyma's "secure by default" principle.

## Solution
To allow internal traffic, you must create an **AuthorizationPolicy**.
If APIRule is applied, internal traffic is blocked by default. To allow it, you need to create an ALLOW-type AuthorizationPolicy.

See the following example of an **AuthorizationPolicy** that allows internal traffic to the given workload. 
Note that it excludes traffic coming from `istio-ingressgateway` not to interfere with policies applied by APIRule to external traffic.
  
> [!NOTE] Replace `${NAMESPACE}`, `${KEY}`, and `${TARGET_WORKLOAD}` with the appropriate values for your environment.
```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: ${NAMESPACE}
spec:
  selector:
    matchLabels:
      ${KEY}: ${TARGET_WORKLOAD}
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```