# Network Policies

Learn about the network policies for the API Gateway module and how to manage them.

To increase security, you can enable network policy support in the API Gateway module. When enabled, the API Gateway module creates the `kyma-project.io--api-gateway-allow` network policy that controls traffic to and from the API Gateway module's Pods. It allows the following connections:
- Egress connection to the `kube-dns` and `local-dns` workloads to allow name resolution for in-cluster configurations.
- Egress connection to the Kubernetes API server on port `443`.
- Ingress connection for webhook (`9443`) port.
- Ingress connection for metrics (`8080`) ports from external workloads labeled with `networking.kyma-project.io/metrics-scraping=allowed` label, or from other Kyma modules.

The network policy support is disabled by default. The network policies are applied only when you enable this setting.

> ### Caution:
> We do not cover ORY Oathkeeper connectivity with network policies. Oathkeeper support is deprecated, and we do not plan to extend this functionality.
> Because Ory Oathkeeper support is deprecated, network policies do not cover Ory Oathkeeper connectivity.
> It's not recommended to use network policies when Ory Oathkeeper support is enabled. If you still use Ory Oathkeeper and wish to enable network policies, expose your workloads with APIRules in version `v2`.

## Enable Network Policies

To enable network policies, set the **networkPoliciesEnabled** field to `true` in the APIGateway custom resource:

```sh
kubectl patch apigateways.operator.kyma-project.io default --type=merge --patch='{"spec": {"networkPoliciesEnabled": true}}'
```

## Disable Network Policies

To disable network policies, set the **networkPoliciesEnabled** field to `false` in the APIGateway custom resource:

```sh
kubectl patch apigateways.operator.kyma-project.io default --type=merge --patch='{"spec": {"networkPoliciesEnabled": false}}'
```

> ### Caution:
> If you disable network policies in supported clusters, the traffic to the API server or other critical workloads may be disrupted.
> Always consult your cluster configuration to choose the best solution.

## Verify Status

To check if the network policies are active, run:

```
kubectl get networkpolicies -n kyma-system -l api-gateway.kyma-project.io/managed-by
```
