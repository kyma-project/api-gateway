# Missing Hosts in APIRules List

## Symptoms

When you run the following command, you see no hosts listed for an APIRule:
```bash
kubectl get apirules.gateway.kyma-project.io -n {NAMESPACE}
NAME                       STATUS   HOSTS
httpbin-v1-beta-1          Ready    
httpbin-v2-shorthost       Ready    ["httpbin-shorthost"]
httpbin-v2                 Ready    ["httpbin-v2.local.kyma.dev"]
```

## Causes
After the release of API Gateway 3.0.0, running the command `kubectl get apirules.gateway.kyma-project.io -n {NAMESPACE}` without specifying a version returns a list of APIRules for the default version `v2`. However, for resources that were originally created in version `v1beta1` and are not convertible or updated to version `v2`, the **hosts** field is not displayed in the APIRules list.

## Solution
To get the list of hosts, run the following command:
```bash
kubectl get apirules.v1beta1.gateway.kyma-project.io -n {NAMESPACE}
NAME                       STATUS   HOST
httpbin-v1-beta-1          OK       httpbin
httpbin-v2-shorthost       OK       httpbin-shorthost
httpbin-v2                 OK       httpbin-v2.local.kyma.dev
```
When you run `kubectl get apirules.v1beta1.gateway.kyma-project.io -n {NAMESPACE}` with the specified version of the resource, the command returns the list of APIRules in version `v1beta1`. The original host used in APIRule `v1beta1` is displayed in the list.