# Kyma Gateway
Kyma Gateway is a preconfigured [Istio Gateway CR](https://istio.io/latest/docs/reference/config/networking/gateway/) named `kyma-gateway` located in the `kyma-system` namespace. It describes the ports and protocols exposed for a particular domain.

> [!WARNING]
> Kyma Gateway is not recommended for production environments. For production use, set up a custom gateway with your
> own domain and certificate. For more information, see [Istio Gateways](./README.md).
> For instructions on how to set up a custom gateway, see the following topics:
> - For TLS, see [Configure a TLS Gateway in SAP BTP, Kyma Runtime](./set-up-tls-gateway.md).
> - For mutual TLS (mTLS), see [Mutual TLS Authentication](./mtls-context.md) and
>   [Configure mTLS Authentication in SAP BTP, Kyma Runtime](./configure-mtls-Gardener-certs.md).


## Gateway Configuration

Kyma Gateway is configured with the following settings in both SAP BTP, Kyma Runtime and open-source Kyma clusters:
- It listens on port `443` (HTTPS) using TLS mode `SIMPLE`, with a TLS credential supplied from the `kyma-gateway-certs` Secret in the `istio-system` namespace.
- It listens on port `80` (HTTP) and automatically redirects all HTTP requests to HTTPS (responds with a `301` status code).
- It serves all hosts matching the wildcard `*.{domain}`, where `{domain}` is the cluster domain resolved at reconciliation
  time.
- The gateway selector targets the default Istio ingress gateway (`app: istio-ingressgateway`, `istio: ingressgateway`).
- The `istio-healthz` VirtualService is reconciled in the `istio-system` namespace. It exposes the
Istio readiness endpoint at `healthz.{domain}/healthz/ready` through `kyma-gateway`.

The following table summarises how the Kyma Gateway configuration differs between environments:

| | SAP BTP, Kyma Runtime | Open-Source Kyma |
|---|---|---|
| **Domain** | Gardener Shoot domain | `local.kyma.dev` |
| **TLS certificate** | Managed by a Gardener Certificate CR | Pre-populated self-signed cert (valid until July 2030) |
| **DNS** | Managed by a Gardener DNSEntry CR | Must be configured externally |

## SAP BTP, Kyma Runtime

In a managed SAP BTP, Kyma runtime cluster, Kyma Gateway uses the Gardener Shoot domain. The operator automatically manages the TLS certificate and DNS record using Gardener resources. An Istio VirtualService exposes the Istio readiness endpoint at `healthz.{GARDENER_SHOOT_DOMAIN}/healthz/ready`.

![Kyma Gateway Resources Gardener](../../assets/kyma-gateway-resources-gardener.svg)

### DNS Resolution
The cluster domain is resolved from the Gardener `shoot-info` ConfigMap. The operator creates and manages a
[DNSEntry CR](https://gardener.cloud/docs/guides/networking/DNS-extension#creating-a-dnsentry-resource-explicitly)
named `kyma-gateway` in the `kyma-system` namespace. The `DNSEntry` points to the external IP addresses or hostnames
of the `istio-ingressgateway` LoadBalancer Service in the `istio-system` namespace.

The operator detects the IP stack of the `istio-ingressgateway` Service by inspecting the **spec.ipFamilies** field:

| IP families detected | IP stack type | Annotation set on DNSEntry                |
|----------------------|---------------|-------------------------------------------|
| Single entry: `IPv4` | IPv4          | _(none)_                                  |
| Single entry: `IPv6` | IPv6          | `dns.gardener.cloud/ip-stack: ipv6`       |
| Two entries          | Dual-stack    | `dns.gardener.cloud/ip-stack: dual-stack` |

### Certificate Management
The operator creates and manages a
[Gardener Certificate CR](https://gardener.cloud/docs/guides/networking/certificate-extension#using-the-custom-certificate-resource)
named `kyma-tls-cert` in the `istio-system` namespace. The certificate covers all subdomains of the cluster domain
(`*.{domain}`) and automatically stores the TLS data in the `kyma-gateway-certs` Secret.

## Open-Source Kyma

In an open-source Kyma cluster, Kyma Gateway uses the `local.kyma.dev` domain and a pre-populated self-signed certificate. It is intended for local development only. An Istio VirtualService exposes the Istio readiness endpoint at `healthz.local.kyma.dev/healthz/ready`.

![Kyma Gateway Resources Open Source](../../assets/kyma-gateway-resources-os.svg)

### DNS Resolution
No `DNSEntry` is created. DNS resolution must be configured externally.

### Certificate Management
The operator creates a pre-populated Kubernetes Secret named `kyma-gateway-certs` in the `istio-system` namespace.
This Secret contains a self-signed certificate valid until July 2030 and is intended for local development use with
the `local.kyma.dev` domain.

## Enable or Disable Kyma Gateway
By default, Kyma Gateway is enabled. To disable it, remove the **enableKymaGateway** field from the [APIGateway CR](../custom-resources/apigateway/04-00-apigateway-custom-resource.md) or set **enableKymaGateway** to `false`:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: APIGateway
metadata:
  name: default
spec:
  enableKymaGateway: false
```

When you disable Kyma Gateway or delete the APIGateway CR, the operator removes all managed resources: Gateway, DNSEntry, Certificate or certificate Secret, and VirtualService.

You can only disable Kyma Gateway if no APIRule or VirtualService resources in the cluster reference `kyma-system/kyma-gateway`. If such resources exist, the operator returns a warning listing up to five of them. Remove or migrate those resources to a different gateway before disabling Kyma Gateway.