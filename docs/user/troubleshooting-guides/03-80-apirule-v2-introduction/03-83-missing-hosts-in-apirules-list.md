# Missing Hosts in APIRules List

## Symptoms

When you run the following command, you see no hosts listed for the APIRule on list:
```bash
kubectl get apirules.gateway.kyma-project.io -n {NAMESPACE}
NAME                       STATUS   HOSTS
httpbin-v1-beta-1          Ready    
httpbin-v2-shorthost       Ready    ["httpbin-shorthost"]
httpbin-v2                 Ready    ["httpbin-v2.local.kyma.dev"]
```

## Causes
From time of release of Api Gateway v3.0.0, the `hosts` field is not displayed in the list of APIRules for default version v2, when resources are APIRules v1beta1 on the list and are not convertible or updated to v2 version.

## Solution
To get the list of host, run the following command:
```bash
kubectl get apirules.v1beta1.gateway.kyma-project.io -n {NAMESPACE}
NAME                       STATUS   HOST
httpbin-v1-beta-1          OK       httpbin
httpbin-v2-shorthost       OK       httpbin-shorthost
httpbin-v2                 OK       httpbin-v2.local.kyma.dev
```
This command lists the APIRules in the v1beta1 version. Calling `kubectl get apirules.v1beta1.gateway.kyma-project.io -n {NAMESPACE}` with specified version of resource, returns the list of APIRules converted to that version which display original host from `v1beta1` version of APIRule.