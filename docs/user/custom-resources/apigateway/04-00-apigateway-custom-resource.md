# APIGateway Custom Resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes 
the kind and the format of data that APIGateway Controller uses to configure the 
API Gateway resources. Applying the custom resource (CR) triggers the installation 
of API Gateway resources, and deleting it triggers the uninstallation of those resources. 
The default CR has the name `default`.

```bash
kubectl get crd apigateways.operator.kyma-project.io -o yaml
```

You are only allowed to have one APIGateway CR. If there are multiple APIGateway CRs 
in the cluster, the oldest one reconciles the module. Any additional APIGateway CR 
is placed in the `Warning` state.

## Sample Custom Resource
This is a sample APIGateway CR:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: APIGateway
metadata:
  labels:
    operator.kyma-project.io/managed-by: kyma
  name: default
spec:
  enableKymaGateway: true
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
- operator.kyma-project.io/v1alpha1

### Resource Types
- [APIGateway](#apigateway)

### APIGateway

APIGateway is the Schema for APIGateway APIs.

| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br /> string | `operator.kyma-project.io/v1alpha1` | Optional |
| **kind** <br /> string | `APIGateway` | Optional |
| **metadata** <br /> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | Optional |
| **spec** <br /> [APIGatewaySpec](#apigatewayspec) | Defines the desired state of APIGateway CR. | Optional |
| **status** <br /> [APIGatewayStatus](#apigatewaystatus) | Defines the observed status of APIGateway CR. | Optional |

### APIGatewaySpec

Defines the desired state of APIGateway CR.

Appears in:
- [APIGateway](#apigateway)

| Field | Description | Validation |
| --- | --- | --- |
| **enableKymaGateway** <br /> boolean | Specifies whether the default Kyma Gateway `kyma-gateway` in `kyma-system` namespace is created. | Optional |
| **networkPoliciesEnabled** <br /> boolean | Enables network policy reconciliation support for the API Gateway module. | Optional <br /> |

### APIGatewayStatus

Defines the observed state of APIGateway CR.

Appears in:
- [APIGateway](#apigateway)

| Field | Description | Validation |
| --- | --- | --- |
| **state** <br /> [State](#state) | State signifies current state of APIGateway. The possible values are `Ready`, `Processing`, `Error`, `Deleting`, `Warning`. | Enum: [Processing Deleting Ready Error Warning] <br />Required <br /> |
| **description** <br /> string | Contains the description of the APIGateway's state. | Optional |
| **conditions** <br /> [Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#condition-v1-meta) array | Contains conditions associated with the APIGateway's status. | Optional |

### Conditions

The APIGateway CR exposes per-subsystem conditions in `status.conditions` alongside the aggregate `Ready` condition. Each condition follows the standard Kubernetes condition format with `type`, `status`, `reason`, and `message` fields.

Use conditions to observe the state of individual subsystems without parsing `status.description`. For example:

```bash
kubectl wait --for=condition=KymaGatewayReady apigateway/default
```

#### Ready

The top-level aggregate condition. When `status` is `True`, all subsystems reconciled successfully.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `ReconcileSucceeded` | All subsystems reconciled successfully. |
| `Unknown` | `ReconcileProcessing` | Reconciliation is in progress (initial install or spec change). |
| `False` | `ReconcileFailed` | A hard failure occurred with no more specific reason. |
| `False` | `OlderCRExists` | This CR is not the oldest in the cluster and does not reconcile the module. |
| `False` | `DeletionBlockedExistingResources` | Deletion is blocked by existing APIRules, ORY Rules, or RateLimits. |
| `False` | `OathkeeperReconcileFailed` | Ory Oathkeeper reconciliation failed. |

#### KymaGatewayReady

Reports the state of the `kyma-gateway` Istio Gateway and its supporting resources.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `KymaGatewayReconcileSucceeded` | Gateway reconciled successfully. |
| `False` | `KymaGatewayReconcileDisabled` | Gateway reconciliation is skipped because `spec.enableKymaGateway` is `false`. |
| `Unknown` | `KymaGatewayReconcileProcessing` | Certificate is still being issued on initial install. |
| `False` | `KymaGatewayReconcileFailed` | Gateway reconciliation failed, or certificate renewal is pending while the gateway was previously ready. |
| `False` | `KymaGatewayDeletionBlocked` | Custom resources (APIRules or VirtualServices) block gateway deletion. |

#### DNSEntryReady

Reports the state of the Gardener-managed DNS entry for the Kyma Gateway domain. Only present on Gardener clusters.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `DNSEntryReconcileSucceeded` | DNS entry provisioned successfully. |
| `Unknown` | `DNSEntryReconcilePending` | DNS entry applied but not yet provisioned. |
| `False` | `DNSEntryReconcileFailed` | DNS entry provisioning failed. |

#### CertificateReady

Reports the state of the Gardener-managed TLS certificate for the Kyma Gateway domain. Only present on Gardener clusters.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `CertificateReconcileSucceeded` | Certificate issued successfully. |
| `Unknown` | `CertificateReconcilePending` | Certificate applied but not yet issued. |
| `False` | `CertificateReconcileFailed` | Certificate issuance failed. |

#### NetworkPolicyReady

Reports the state of the NetworkPolicy resources managed by API Gateway.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `NetworkPolicyReconcileSucceeded` | NetworkPolicy reconciled successfully. |
| `False` | `NetworkPolicyReconcileDisabled` | NetworkPolicy reconciliation is skipped because `spec.networkPoliciesEnabled` is `false`. |
| `False` | `NetworkPolicyReconcileFailed` | NetworkPolicy reconciliation failed. |

#### DependenciesReady

Reports whether all required CRDs (for example, VirtualService) are present on the cluster.

| Status | Reason | Description |
| --- | --- | --- |
| `True` | `DependenciesAvailable` | All required CRDs are present. |
| `False` | `DependenciesMissing` | One or more required CRDs are not present. |
| `False` | `DependenciesCheckFailed` | An error occurred during dependency discovery. |

### State

Underlying type: string

Appears in:
- [APIGatewayStatus](#apigatewaystatus)
 
| Field | Description |
| --- | --- |
| **Ready** | APIGateway Controller finished reconciliation.<br /> |
| **Processing** | APIGateway Controller is reconciling resources.<br /> |
| **Error** | An error occurred during the reconciliation.<br />The error is rather related to the API Gateway module than the configuration of your resources.<br /> |
| **Deleting** | APIGateway Controller is deleting resources.<br /> |
| **Warning** | An issue occurred during reconciliation that requires your attention.<br />Check the **status.description** message to identify the issue and make the necessary corrections<br />to the APIGateway CR or any related resources.<br /> |

