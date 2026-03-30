# APIGateway Kyma Gateway

Kyma Gateway is an [Istio Gateway CR](https://istio.io/latest/docs/reference/config/networking/gateway/) named
`kyma-gateway` that is located in the `kyma-system` namespace. Istio Gateway describes which ports and protocols
should be exposed for a particular domain.
The configuration of Kyma Gateway varies depending on whether you use a managed SAP BTP, Kyma runtime cluster, or
an open-source Kyma cluster.

## Gateway Configuration

The `kyma-gateway` is configured to:
- Listen on **port 443** (HTTPS) using TLS mode `SIMPLE`, with a TLS credential supplied from the `kyma-gateway-certs`
  secret in the `istio-system` namespace.
- Listen on **port 80** (HTTP) and automatically redirect all HTTP requests to HTTPS (HTTP 301).
- Serve all hosts matching the wildcard `*.{domain}`, where `{domain}` is the cluster domain resolved at reconciliation
  time.

The gateway selector targets the default Istio ingress gateway (`app: istio-ingressgateway`, `istio: ingressgateway`).
Furthermore, a `VirtualService` named `istio-healthz` is reconciled in the `istio-system` namespace. It exposes the
Istio readiness endpoint at `healthz.{domain}/healthz/ready` through the `kyma-gateway`.

## SAP BTP, Kyma Runtime
In a managed SAP BTP, Kyma runtime cluster, Kyma Gateway uses the Gardener Shoot domain. For this domain, an Istio
Gateway CR exposes the HTTPS port (`443`) and the HTTP port (`80`) with a redirect to port `443`.
Istio Gateway uses a certificate managed by a [Gardener Certificate CR](https://gardener.cloud/docs/guides/networking/certificate-extension#using-the-custom-certificate-resource).
The Gardener [DNSEntry CR](https://gardener.cloud/docs/guides/networking/DNS-extension#creating-a-dnsentry-resource-explicitly)
creates a DNS record for the specified domain with the Istio Ingress Gateway Load Balancer Service as the target.
Furthermore, an Istio Virtual Service is created, which exposes the Istio readiness endpoint at
`healthz.{GARDENER_SHOOT_DOMAIN}/healthz/ready`.

![Kyma Gateway Resources Gardener](../../../assets/kyma-gateway-resources-gardener.svg)

### DNS Resolution
The cluster domain is resolved from the Gardener `shoot-info` ConfigMap. The operator creates and manages a
[DNSEntry CR](https://gardener.cloud/docs/guides/networking/DNS-extension#creating-a-dnsentry-resource-explicitly)
named `kyma-gateway` in the `kyma-system` namespace. The `DNSEntry` points to the external IP addresses or hostnames
of the `istio-ingressgateway` LoadBalancer Service in the `istio-system` namespace.

The operator detects the IP stack of the `istio-ingressgateway` Service by inspecting the `spec.ipFamilies` field:

| IP families detected | IP stack type | Annotation set on DNSEntry                |
|----------------------|---------------|-------------------------------------------|
| Single entry: `IPv4` | IPv4          | _(none)_                                  |
| Single entry: `IPv6` | IPv6          | `dns.gardener.cloud/ip-stack: ipv6`       |
| Two entries          | Dual-stack    | `dns.gardener.cloud/ip-stack: dual-stack` |

### Certificate Management
The operator creates and manages a
[Gardener Certificate CR](https://gardener.cloud/docs/guides/networking/certificate-extension#using-the-custom-certificate-resource)
named `kyma-tls-cert` in the `istio-system` namespace. The certificate covers all subdomains of the cluster domain
(`*.{domain}`) and stores the resulting TLS data in the `kyma-gateway-certs` secret automatically.

## Open-Source Kyma
In an open-source Kyma cluster, Kyma Gateway uses the domain `local.kyma.dev`. For this domain, an Istio Gateway CR
exposes the HTTPS port (`443`) and the HTTP port (`80`) with a redirect to port `443`.
Istio Gateway uses a default certificate for the domain `local.kyma.dev` that is valid until July 2030.
Furthermore, an Istio Virtual Service is created, which exposes the Istio readiness endpoint at
`healthz.local.kyma.dev/healthz/ready`.

![Kyma Gateway Resources Open Source](../../../assets/kyma-gateway-resources-os.svg)

### DNS Resolution
No `DNSEntry` is created. DNS resolution must be configured externally.

### Certificate Management
The operator creates a pre-populated Kubernetes `Secret` named `kyma-gateway-certs` in the `istio-system` namespace.
This secret contains a self-signed certificate valid until July 2030 and is intended for local development use with
the `local.kyma.dev` domain.

## Disable or Enable Kyma Gateway
By default, Kyma Gateway is enabled. You can disable it by removing the `enableKymaGateway` field or setting it to
`false` in the [APIGateway CR](./04-00-apigateway-custom-resource.md):

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: APIGateway
metadata:
  name: default
spec:
  enableKymaGateway: false
```

When disabled or when the `APIGateway` CR is deleted, the operator removes all managed resources: the Gateway,
DNSEntry, Certificate or certificate Secret, and VirtualService.
Kyma Gateway can be disabled only if no `APIRule` or `VirtualService` resources in the cluster reference
`kyma-system/kyma-gateway`. If such resources exist, the operator surfaces a warning listing up to five of them.
You must remove or migrate those resources to use a different Gateway before the Kyma Gateway can be disabled.
