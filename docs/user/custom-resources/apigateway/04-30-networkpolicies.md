<!-- loio1eb05cf4196e4a63b60c38417299c746 -->

# Network Policies

Learn about the network policies for the API Gateway module and how to manage them.



To increase security, the API Gateway module creates the following network policy that control traffic to and from the API Gateway module's Pods:
`kyma-project.io--api-gateway-allow` It allows:
- Egress connection to the kube-dns and local-dns workloads to allow name resolution for in-cluster configurations.
- Egress connection to the kubernetes API server on port 443.
- Ingress connection for webhook (9443) and metrics (8080) ports.

> ### Caution:
> We do not cover ORY Oathkeeper connectivity with network policies. Oathkeeper support is deprecated, and we do not plan to extend this functionality.
> We do not recommend using network policies with ORY Oathkeeper support enabled. If you still use it and wish to enable network policies, please use APIRules in v2 version to expose your workloads.

## Enable Network Policies

To enable network policies, set the `networkPoliciesEnabled` field to `true` in the APIGateway custom resource:

```sh
kubectl patch apigateways.operator.kyma-project.io default --type=merge --patch='{"spec": {"networkPoliciesEnabled": true}}'
```

## Disable Network Policies

To disable network policies, set the `networkPoliciesEnabled` field to `false` in the APIGateway custom resource:

```sh
kubectl patch apigateways.operator.kyma-project.io default --type=merge --patch='{"spec": {"networkPoliciesEnabled": false}}'
```

> ### Caution:
> If you disable NetworkPolicies on supported clusters, the traffic to the API server or other critical workloads may be disrupted.
> Always consult your cluster configuration and choose the best solution.

## Verify Status

To check if the network policies are active, run:

```
kubectl get networkpolicies -n kyma-system -l api-gateway.kyma-project.io/managed-by
```
