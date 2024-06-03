# Issues When Creating an APIRule Custom Resource in Version `v1beta2`

## Symptom

When you create an APIRule custom resource (CR), a validation error appears, or the APIRule CR has the `ERROR` status, for example:

```bash
kubectl get apirules httpbin

NAME      STATUS   HOST
httpbin   ERROR    httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

The error may result in an inconsistent state of the APIRule CR. Your Service then cannot be properly exposed.
To check the error message of the APIRule CR, run:


```bash
kubectl get apirules -n <namespace> <api-rule-name> -o=jsonpath='{.status.APIRuleStatus}'
```

---
## Invalid **issuer** for the **jwt** Access Strategy
### Cause

Here's an example of an APIRule with an invalid **issuer** URL configured:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /.*
      jwt:
        authentications:
          - issuer: http://unsecured.or.not.valid.url
            jwksUri: https://example.com/.well-known/jwks.json
```

If the **issuer** URL is an unsecured HTTP URL, or the **issuer** URL is not valid, you get the following error, and the APIRule resource is not created:

```
TBD
```

### Remedy

The JWT **issuer** must be a valid HTTPS URL, for example:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /.*
      jwt:
        authentications:
          - issuer: https://dev.kyma.local
            jwksUri: https://dev.kyma.local/.well-known/jwks.json
```

---
## Both **noAuth** and **jwt** Access Strategies Defined on the Same Path
### Cause

The following APIRule has both **noAuth** and **jwt** access strategies defined on the same path:

```yaml
spec:
  ...
  rules:
    - path: /.*
      noAuth: true
      jwt:
        authentications:
          - issuer: https://dev.kyma.local
            jwksUri: https://dev.kyma.local/.well-known/jwks.json
```

If you set **noAuth** access strategy to `true` and define the **jwt** configuration on the same path, you get the following `APIRuleStatus` error appears:

```
TBD
```

### Remedy

Decide on one configuration you want to use. You can either **noAuth** access to the specific path or restrict it using a JWT security token.

---
## Occupied Host
### Cause

The following APIRule CRs use the same host:

```yaml
spec:
  ...
spec:
  host: httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

If your APIRule CR specifies a host that is already used by another APIRule or Virtual Service, the following `APIRuleStatus` error appears:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.host\": This host is occupied by another Virtual Service"}
```

### Remedy

Use a different host for the second APIRule CR, for example:

```yaml
spec:
  ...
  host: httpbin-new.xxx.shoot.canary.k8s-hana.ondemand.com
```

---
## Additional configuration for **noAuth** Access Strategy

### Cause

In the following APIRule CR, the **noAuth** access strategy has the **issuer** field configured:

```yaml
spec:
  ...
  rules:
    - path: /.*
      noAuth: true
        authentications:
          - issuer: https://dev.kyma.local
            jwksUri: https://dev.kyma.local/.well-known/jwks.json
```

If your APIRule CR uses the **noAuth** access strategy and has some further configuration defined, you get the following `APIRuleStatus` error:

```
TBD
```


### Remedy

Use the **noAuth** access strategy without any further configuration:

```yaml
spec:
  ...
  rules:
    - path: /.*
      noAuth: true
```
