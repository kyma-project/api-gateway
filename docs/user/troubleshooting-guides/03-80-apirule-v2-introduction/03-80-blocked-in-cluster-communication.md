# Blocked in-cluster communication

## Symptoms
After switching from APIRule `v1beta1` to version `v2`  in-cluster communication is blocked.

## Cause

By default, the access to the workload from internal traffic is blocked.
This approach aligns with Kyma's "secure by default" principle.
In one of the future releases of the API Gateway module, the APIRule custom resource (CR) will contain a new field **internalTraffic** set to `Deny` by default. This field will allow you to permit traffic from the CR. For more information on this topic, see issue [#1632](https://github.com/kyma-project/api-gateway/issues/1632).

## Solution

To retain APIRule `v1beta1` internal traffic policy, apply the following AuthorizationPolicy. Remember to change the selector label to the one pointing to the target workload:
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