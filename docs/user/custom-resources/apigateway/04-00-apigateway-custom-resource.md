

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

APIGateway is the Schema for the apigateways API

| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br /> string | `operator.kyma-project.io/v1alpha1` | None |
| **kind** <br /> string | `APIGateway` | None |
| **metadata** <br /> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | Optional |
| **spec** <br /> [APIGatewaySpec](#apigatewayspec) |  | Optional |
| **status** <br /> [APIGatewayStatus](#apigatewaystatus) |  | Optional |

### APIGatewaySpec

APIGatewaySpec defines the desired state of APIGateway

Appears in:
- [APIGateway](#apigateway)

| Field | Description | Validation |
| --- | --- | --- |
| **enableKymaGateway** <br /> boolean | Specifies whether the default Kyma Gateway kyma-gateway in kyma-system Namespace is created. | Optional |

### APIGatewayStatus

APIGatewayStatus defines the observed state of APIGateway

Appears in:
- [APIGateway](#apigateway)

| Field | Description | Validation |
| --- | --- | --- |
| **state** <br /> [State](#state) | State signifies current state of APIGateway. Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning"). | Enum: [Processing Deleting Ready Error Warning] <br />Required: \{\} <br /> |
| **description** <br /> string | Description of APIGateway status | Optional |
| **conditions** <br /> [Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#condition-v1-meta) array | Conditions of APIGateway | Optional |

### State

Underlying type: string

Appears in:
- [APIGatewayStatus](#apigatewaystatus)
 
| Field | Description |
| --- | --- |
| **Ready** |  |
| **Processing** |  |
| **Error** |  |
| **Deleting** |  |
| **Warning** |  |

