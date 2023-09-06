# API Gateway Operator parameters 

You can configure the [API Gateway Controller](../../00-10-overview-api-gateway-controller.md) and [API Rule Controller](../../00-20-overview-api-rule-controller.md) using various parameters. All options are listed in this document.

## Reconciliation interval

### APIGateway
Kyma API Gateway Operator reconciles the APIGateway custom resource every 10 hours or whenever it is changed.

### APIRule
By default, Kyma API Gateway Operator reconciles APIRules every 60 minutes or whenever the APIRule is changed. You can adjust this interval by modifying the operator's parameters. For example, you can set the **-reconciliation-interval** parameter to `120s`.

## All configuration parameters

| Name                          | Required | Description                                                                                                           | Example values                                   |
|-------------------------------|:--------:|-----------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|
| **metrics-bind-address**      |    NO    | The address the metric endpoint binds to.                                                                             | `:8080`                                          |
| **health-probe-bind-address** |    NO    | The address the probe endpoint binds to.                                                                              | `:8081`                                          |
| **leader-elect**              |    NO    | Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager. | `true`                                           |
| **rate-limiter-burst**        |    NO    | Indicates the burst value for the controller bucket rate limiter.                                                     | 200                                              |
| **rate-limiter-frequency**    |    NO    | Indicates the controller bucket rate limiter frequency, signifying no. of events per second.                          | 30                                               |
| **failure-base-delay**        |    NO    | Indicates the failure base delay for rate limiter.                                                                    | `1s`                                             |
| **failure-max-delay**         |    NO    | Indicates the failure max delay for rate limiter.                                                                     | `1000s`                                          |
| **service-blocklist**         |    NO    | List of services to be blocklisted.                                                                                   | `kubernetes.default` <br> `kube-dns.kube-system` |
| **domain-allowlist**          |    NO    | List of domains that can be exposed. All domains are allowed if empty                                                 | `kyma.local` <br> `foo.bar`                      |
| **generated-objects-labels**  |    NO    | Comma-separated list of key-value pairs used to label generated objects.                                              | `managed-by=api-gateway`                         |
| **reconciliation-interval**   |    NO    | Indicates the time based reconciliation interval of APIRule.                                                          | `1h`                                             |