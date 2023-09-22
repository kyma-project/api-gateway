# APIRule custom resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the APIGateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Specification of APIRule custom resource

This table lists all parameters of APIRule CRD together with their descriptions:

**Spec:**

| Field                         | Mandatory | Description                                                                                                                                                                                                                                                                                            |
|-------------------------------|:---------:|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| **gateway**                   |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                           |
| **host**                      |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used.                                                                                                                                              |
| **service.name**              |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                             |
| **service.namespace**         |  **NO**   | Specifies the Namespace of the exposed Service.                                                                                                                                                                                                                                                        |
| **service.port**              |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                               |
| **timeout**                   |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Oathkeeper access rules but can be overridden for each rule. The maximum timeout is limited to 3900 seconds (65 minutes). </br> If no timeout is specified, the default timeout of 180 seconds applies.                                    |
| **rules**                     |  **YES**  | Specifies the list of Oathkeeper access rules.                                                                                                                                                                                                                                                         |
| **rules.service**             |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                   |
| **rules.path**                |  **YES**  | Specifies the path of the exposed Service.                                                                                                                                                                                                                                                             |
| **rules.methods**             |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**.                                                                                                                                                                                                                          |
| **rules.mutators**            |  **NO**   | Specifies the list of the [Oathkeeper](https://www.ory.sh/docs/next/oathkeeper/pipeline/mutator) or Istio mutators.                                                                                                                                                                                        |
| **rules.accessStrategies**    |  **YES**  | Specifies the list of access strategies. Supported are the [Oathkeeper's](https://www.ory.sh/docs/next/oathkeeper/pipeline/authn) `oauth2_introspection`, `jwt`, `noop` and `allow`. We also support `jwt` as [Istio](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy. |
| **rules.timeout**             |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. The maximum timeout is limited to 3900 seconds (65 minutes). Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                    |

>**CAUTION:** If `service` is not defined at **spec.service** level, all defined Rules must have `service` defined at **spec.rules.service** level. Otherwise, the validation fails.

>**CAUTION:** Having both the Oathkeeper and Istio `jwt` access strategies defined is not supported. Access strategies `noop` or `allow` cannot be used with any other access strategy on the same **spec.rules.path**.

**Status:**

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the VirtualService and the Oathkeeper Access Rule created for this CR. The following table lists the fields of the **status** section.

| Field                                  | Description                                  |
|:---------------------------------------|:---------------------------------------------|
| **status.apiRuleStatus**               | Status code describing the APIRule CR.       |
| **status.virtualServiceStatus.code**   | Status code describing the VirtualService.   |
| **status.virtualService.desc**         | Current state of the VirtualService.         |
| **status.accessRuleStatus.code**       | Status code describing the Oathkeeper Rule.  |
| **status.accessRuleStatus.desc**       | Current state of the Oathkeeper Rule.        |

**Status codes:**

The following status codes describe VirtualServices and Oathkeeper Access Rules:

| Code          | Description                    |
|---------------|--------------------------------|
| **OK**        | Resource created.              |
| **SKIPPED**   | Skipped creating a resource.   |
| **ERROR**     | Resource not created.          |

## Sample custom resource

This is a sample custom resource (CR) that the APIGateway Controller listens for to expose a Service. The following example has the **rules** section specified, which makes APIGateway Controller create an Oathkeeper Access Rule for the Service.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360  
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```