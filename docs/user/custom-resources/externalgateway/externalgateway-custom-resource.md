# ExternalGateway Custom Resource

The `externalgateways.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data used to configure an ExternalGateway custom resource (CR). To get the up-to-date CRD in the yaml format, run the following command:

```bash
kubectl get crd externalgateways.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource
This is a sample ExternalGateway CR:

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: ExternalGateway
metadata:
  name: my-app
  namespace: my-namespace
spec:
  externalDomain: api.customer.com
  internalDomain:
    kymaSubdomain: external-myapp
  region: eu10
  regionsConfigMap: external-gateway-regions
  caSecretRef:
    name: ca-certificate
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
- gateway.kyma-project.io/v1alpha1

### Resource Types
- [ExternalGateway](#externalgateway)

### ExternalGateway

ExternalGateway defines the Schema for the ExternalGateway API.

| Field | Description | Validation |
| --- | --- | --- |
| **apiVersion** <br /> string | `gateway.kyma-project.io/v1alpha1` | Optional |
| **kind** <br /> string | `ExternalGateway` | Optional |
| **metadata** <br /> [ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta) | For more information on the metadata fields, see Kubernetes API documentation. | Optional |
| **spec** <br /> [ExternalGatewaySpec](#externalgatewayspec) | ExternalGatewaySpec defines the desired state of ExternalGateway. | Optional |
| **status** <br /> [ExternalGatewayStatus](#externalgatewaystatus) | ExternalGatewayStatus defines the observed state of an ExternalGateway. | Optional |

### ExternalGatewaySpec

ExternalGatewaySpec defines the desired state of ExternalGateway.

Appears in:
- [ExternalGateway](#externalgateway)

| Field | Description | Validation |
| --- | --- | --- |
| **externalDomain** <br /> string | ExternalDomain is the customer-facing domain, for example, `api.customer.com` or `*.api.customer.com`.<br />It uses the Istio Gateway host format and can include an optional wildcard prefix, for example, `*.example.com`. | MaxLength: 255 <br />Pattern: `^(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$` <br />Required <br /> |
| **internalDomain** <br /> [InternalDomainConfig](#internaldomainconfig) | InternalDomain defines the domain configuration for Kyma-internal access. | Required <br /> |
| **region** <br /> string | Region contains a region identifier, for example, `eu10` or `us10`.<br />It must match a region defined in the RegionsConfigMap. | Required <br /> |
| **regionsConfigMap** <br /> string | RegionsConfigMap specifies the name of the ConfigMap that contains region metadata.<br />The ConfigMap must be in the same namespace as the ExternalGateway.<br />If no key is specified in the ConfigMap, it auto-detects a single key or looks for `regions.yaml`. | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br />Required <br /> |
| **caSecretRef** <br /> [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#secretreference-v1-core) | CASecretRef references the Secret containing the CA certificate.<br />This CA is used to validate client certificates during the mTLS handshake.<br />If the namespace is not specified, it defaults to the ExternalGateway's namespace.<br />If the Secret key is not specified in the Secret, it auto-detects a single key or looks for `ca.crt`. | Required <br /> |
| **includeExtGatewayClientCert** <br /> boolean | IncludeExtGatewayClientCert controls whether the ExternalGateway client certificate is included in HTTP headers.<br />By default, this option is disabled, so the client certificate is not included. | Optional <br /> |

### ExternalGatewayStatus

ExternalGatewayStatus defines the observed state of an ExternalGateway.

Appears in:
- [ExternalGateway](#externalgateway)

| Field | Description | Validation |
| --- | --- | --- |
| **lastProcessedTime** <br /> [Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta) | LastProcessedTime represents the last time the ExternalGateway status was processed. | Optional |
| **state** <br /> [State](#state) | State defines the reconciliation state of the ExternalGateway. | Enum: [Processing Ready Error] <br /> |
| **description** <br /> string | Description provides details about the ExternalGateway status. | Optional |

### InternalDomainConfig

InternalDomainConfig defines the configuration for the Kyma-internal domain.

Appears in:
- [ExternalGatewaySpec](#externalgatewayspec)

| Field | Description | Validation |
| --- | --- | --- |
| **kymaSubdomain** <br /> string | KymaSubdomain specifies the subdomain prefix, for example, `external-myapp`.<br />The full internal domain follows the pattern `{kymaSubdomain}.{KYMA_DOMAIN}`. | MaxLength: 63 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |

### State

State defines the reconciliation state of the ExternalGateway.

Underlying type: string

Appears in:
- [ExternalGatewayStatus](#externalgatewaystatus)
 
| Field | Description |
| --- | --- |
| **Processing** | Processing indicates that the ExternalGateway is being created or updated.<br /> |
| **Ready** | Ready indicates that reconciliation of the ExternalGateway has finished.<br /> |
| **Error** | Error indicates that an error occurred during reconciliation.<br /> |

