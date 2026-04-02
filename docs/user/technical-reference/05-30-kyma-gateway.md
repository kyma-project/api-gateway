# Kyma Gateway

> [!WARNING]
> Kyma Gateway is not recommended for production environments. For production use, set up a custom gateway with your
> own domain and certificate. For more information, see [Istio Gateways](../istio-gateways/README.md).
> For specific instructions on how to set up a custom gateway, see the following topics:
> - For TLS, see [Configure a TLS Gateway in SAP BTP, Kyma Runtime](../istio-gateways/set-up-tls-gateway.md).
> - For mutual TLS (mTLS), see [Mutual TLS Authentication](../istio-gateways/mtls-context.md) and
>   [Configure mTLS Authentication in SAP BTP, Kyma Runtime](../istio-gateways/configure-mtls-Gardener-certs.md).

Kyma Gateway is an [Istio Gateway CR](https://istio.io/latest/docs/reference/config/networking/gateway/) named
`kyma-gateway` that is located in the `kyma-system` namespace. Istio Gateway describes ports and protocols exposed for a particular domain.
The configuration of Kyma Gateway varies depending on whether you use a managed SAP BTP, Kyma runtime cluster, or
an open-source Kyma cluster.

## Gateway Configuration

Kyma Gateway is configured with the following settings:
- It listens on port `443` (HTTPS) using TLS mode `SIMPLE`, with a TLS credential supplied from the `kyma-gateway-certs`.
  Secret in the `istio-system` namespace.
- It listens on port `80` (HTTP) and automatically redirects all HTTP requests to HTTPS (responds with a `301` status code).
- It serves all hosts matching the wildcard `*.{domain}`, where `{domain}` is the cluster domain resolved at reconciliation
  time.

The gateway selector targets the default Istio ingress gateway (`app: istio-ingressgateway`, `istio: ingressgateway`).
Furthermore, a VirtualService named `istio-healthz` is reconciled in the `istio-system` namespace. It exposes the
Istio readiness endpoint at `healthz.{domain}/healthz/ready` through `kyma-gateway`.

## SAP BTP, Kyma Runtime
In a managed SAP BTP, Kyma runtime cluster, Kyma Gateway uses the Gardener Shoot domain. For this domain, an Istio
Gateway CR exposes the HTTPS port (`443`) and the HTTP port (`80`) with a redirect to port `443`.
Istio Gateway uses a certificate managed by a [Gardener Certificate CR](https://gardener.cloud/docs/guides/networking/certificate-extension#using-the-custom-certificate-resource).
The Gardener [DNSEntry CR](https://gardener.cloud/docs/guides/networking/DNS-extension#creating-a-dnsentry-resource-explicitly)
creates a DNS record for the specified domain with the Istio Ingress Gateway LoadBalancer Service as the target.
Furthermore, an Istio VirtualService is created, which exposes the Istio readiness endpoint at
`healthz.{GARDENER_SHOOT_DOMAIN}/healthz/ready`.

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
The operator creates and manages a
[Gardener Certificate CR](https://gardener.cloud/docs/guides/networking/certificate-extension#using-the-custom-certificate-resource)
named `kyma-tls-cert` in the `istio-system` namespace. The certificate covers all subdomains of the cluster domain
(`*.{domain}`) and automatically stores the TLS data in the `kyma-gateway-certs` Secret.

## Open-Source Kyma
In an open-source Kyma cluster, Kyma Gateway uses the domain `local.kyma.dev`. For this domain, an Istio Gateway CR
exposes the HTTPS port (`443`) and the HTTP port (`80`) with a redirect to port `443`.
Istio Gateway uses a default certificate for the domain `local.kyma.dev` that is valid until July 2030.
Furthermore, Istio VirtualService exposes the Istio readiness endpoint at
`healthz.local.kyma.dev/healthz/ready`.

![Kyma Gateway Resources Open Source](../../assets/kyma-gateway-resources-os.svg)

### DNS Resolution
No `DNSEntry` is created. DNS resolution must be configured externally.

### Certificate Management
The operator creates a pre-populated Kubernetes Secret named `kyma-gateway-certs` in the `istio-system` namespace.
This Secret contains a self-signed certificate valid until July 2030 and is intended for local development use with
the `local.kyma.dev` domain.

## Disable or Enable Kyma Gateway
By default, Kyma Gateway is enabled. To disable it, remove the **enableKymaGateway** field from the [APIGateway CR](../custom-resources/apigateway/04-00-apigateway-custom-resource.md) or set **enableKymaGateway** to `false`:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: APIGateway
metadata:
  name: default
spec:
  enableKymaGateway: false
```

When you disable Kyma Gateway or delete the APIGateway CR, the operator removes all the managed resources: Gateway,
DNSEntry, Certificate or certificate Secret, and VirtualService.
Kyma Gateway can be disabled only if no `APIRule` or `VirtualService` resources in the cluster reference
`kyma-system/kyma-gateway`. If such resources exist, the operator surfaces a warning listing up to five of them.
You can disable Kyma Gateway only if there are no APIRules or VirtualServices in the cluster that reference
`kyma-system/kyma-gateway`. If such resources exist, you get a warning listing up to 5 of them.
Before disabling Kyma Gateway, remove these resources or migrate them to use a different Gateway.
