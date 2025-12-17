# Understanding Gateways and Domains in Kyma

This document explains what Gateways are in Kubernetes, why you might need a custom Gateway, and how to choose the right Gateway configuration for your use case in Kyma.

## What is a Gateway?

In Kyma runtime, you can create a Gateway that manages inbound and outbound traffic for the service mesh. It acts as a load balancer operating at the edge of the mesh, receiving incoming HTTP/HTTPS connections.

Think of a Gateway as the front door to your cluster:
- It terminates TLS/SSL connections
- It handles domain name routing
- It provides a centralized point for traffic management and security policies
- It connects external clients to internal services

Gateways are implemented by Istio's ingress gateway controller, which is typically deployed as a LoadBalancer Service that receives a public IP address or hostname from your cloud provider.

## What Is a Gateway Host?
A Gateway host is the domain name or hostname that the Gateway listens on to accept incoming traffic. It defines which domain names the Gateway will respond to.

When a request arrives at your cluster's IP address, the Gateway needs to know if it should handle that request. The host field tells the Gateway: "Only handle requests for these specific domain names."

You can have multiple Gateways on the same cluster, each listening on different domains:

```yaml
# Gateway 1: Public APIs
hosts:
  - "*.api.mycompany.com"

# Gateway 2: Admin Portal
hosts:
  - "admin.mycompany.com"

# Gateway 3: Partner APIs
hosts:
  - "*.partners.mycompany.com"
```

## Why Do You Need a Custom Gateway?

While Kyma provides a default Gateway (`kyma-gateway`) that works for many use cases, you may need a custom Gateway for:

### Security Requirements
- **Mutual TLS (mTLS) authentication**: When you need to verify client identities using certificates, you must create a custom mTLS Gateway. The default Gateway only supports simple TLS (server authentication only).
- **Custom certificate management**: When you need specific certificate configurations or want to use certificates from your own Certificate Authority (CA).
- **Isolation**: When you want to separate traffic for different security zones or compliance requirements.

### Domain Management
- **Custom domains**: When you want to use your own domain name instead of the Kyma default domain.
- **Wildcard certificates**: When you need to expose multiple subdomains under a single certificate.
- **Multi-domain support**: When you need to serve different domains with different certificates on the same cluster.

### Traffic Management
- **Custom port configurations**: When you need to expose services on non-standard ports.
- **Protocol support**: When you need to handle specific protocols (HTTP/2, gRPC, TCP, etc.) with custom settings.
- **Multiple ingress points**: When you want separate ingress gateways for different types of traffic (public vs internal, different geographic regions, etc.).

## Gateway Options in Kyma

Kyma provides several Gateway configuration options depending on your domain and security requirements:

1. Default Kyma Gateway (Simple TLS)

**Location**: `kyma-system/kyma-gateway`

**Domain**: Kyma default domain (for example, `*.c-a1b2c3d.kyma.ondemand.com`)

**Security**: Simple TLS (server authentication only)

**Certificate**: Automatically managed by Kyma/Gardener

**Use when:**
- You're getting started with Kyma and don't have a custom domain
- You need basic TLS encryption without custom certificates
- You don't require client authentication (mTLS)
- You want the simplest setup with no certificate management

**Characteristics:**
- Pre-configured and ready to use immediately
- Automatic certificate renewal
- Shared across all workloads in the cluster
- No DNS configuration required

**Example APIRule reference:**
```yaml
gateway: kyma-system/kyma-gateway
```

---

2. Custom TLS Gateway on Kyma Default Domain

**Domain**: Subdomain of Kyma default domain (for example, `*.custom.c-a1b2c3d.kyma.ondemand.com`)

**Security**: Simple TLS (server authentication only)

**Certificate**: Automatically managed by Gardener using Let's Encrypt

**Use when:**
- You want to use the Kyma domain but need a separate Gateway for isolation
- You need custom Gateway configurations (ports, protocols) but don't have your own domain
- You want to experiment with custom Gateways without DNS setup

**Characteristics:**
- Uses a subdomain of the Kyma default domain
- Automatic certificate issuance via Gardener (no DNS provider credentials needed)
- Requires creating Gateway and Certificate resources
- Traffic is isolated from the default `kyma-gateway`

**Example domain setup:**
```bash
# Default Kyma domain: *.c-a1b2c3d.kyma.ondemand.com (used by kyma-gateway)
# Your custom subdomain: *.custom.c-a1b2c3d.kyma.ondemand.com
```

---

3. Custom TLS Gateway on Custom Domain

**Domain**: Your own domain (for example, `*.api.mycompany.com`)

**Security**: Simple TLS (server authentication only)

**Certificate**: Automatically managed by Gardener using Let's Encrypt (requires DNS provider credentials)

**Use when:**
- You want to use your own branded domain name
- You need public-facing APIs with your company domain
- You require control over DNS records

**Characteristics:**
- Full control over domain name
- Requires DNS provider credentials (AWS Route 53, Google Cloud DNS, Azure DNS, etc.)
- Automatic certificate issuance and renewal via ACME/Let's Encrypt
- Must create DNSProvider, DNSEntry, Certificate, and Gateway resources

**Setup complexity**: Medium - requires DNS provider integration

**Example domain setup:**
```bash
# Your owned domain: mycompany.com
# Gateway wildcard: *.api.mycompany.com
# Workload endpoint: httpbin.api.mycompany.com
```

**Learn more**: See [Configure TLS Gateway](./tutorials/01-05-configure-tls.md).

---

4. Custom mTLS Gateway on Kyma Default Domain

**Domain**: Subdomain of Kyma default domain (for example, `*.mtls.c-a1b2c3d.kyma.ondemand.com`)

**Security**: Mutual TLS (both server and client authentication)

**Certificate**: 
- Server certificate: Automatically managed by Gardener using Let's Encrypt
- Client certificates: You provide and manage

**Use when:**
- You need mutual authentication (verify client identities with certificates)
- You want to restrict access to clients with valid certificates
- You're working with B2B APIs or internal services requiring strong authentication
- You don't have your own domain but need mTLS security

**Characteristics:**
- Server certificates automatically issued by Let's Encrypt
- Client certificates are self-signed or from your own CA
- Must configure client CA trust in the Gateway
- Both client and server verify each other's identity
- No DNS provider credentials needed (uses Kyma subdomain)

**Example client verification flow:**
```
1. Client presents certificate → 2. Gateway verifies against trusted CA → 3. If valid, connection allowed
```

---

5. Custom mTLS Gateway on Custom Domain

**Domain**: Your own domain (for example, `*.secure.mycompany.com`)

**Security**: Mutual TLS (both server and client authentication)

**Certificate**: 
- Server certificate: Automatically managed by Gardener using Let's Encrypt (requires DNS provider credentials)
- Client certificates: You provide and manage

**Use when:**
- You need mutual authentication with your own branded domain
- You have strict security and compliance requirements
- You're building partner-facing APIs with certificate-based authentication
- You want full control over both domain and security

**Characteristics:**
- Full domain control with mTLS security
- Requires DNS provider credentials for server certificate issuance
- Client certificates managed separately (self-signed or from your CA)
- Highest level of control and security
- Most complex setup

**Setup complexity**: High - requires DNS provider integration + certificate management

**Learn more**: See [Configure mTLS Authentication](./tutorials/01-10-mtls-authentication/configure-mtls-Gardener-certs.md).

---

## Domain Ownership and Management

### Using Kyma Default Domain
- **No DNS setup required**: Kyma manages DNS automatically
- **Certificate issuance**: Automatic via Gardener, no provider credentials needed
- **Limitation**: You don't control the domain name
- **Best for**: Development, testing, internal tools

### Using Custom Domain
- **DNS setup required**: You must configure DNS records
- **Certificate issuance**: Automatic via Gardener + ACME, requires DNS provider credentials
- **Advantage**: Full control over domain name and branding
- **Best for**: Production, public-facing APIs, customer-facing services

## Certificate Management

### Server Certificates
In all Kyma Gateway configurations (except local k3d), server certificates are automatically managed:
- **Issuer**: Let's Encrypt (publicly trusted CA)
- **Renewal**: Automatic via Gardener
- **Storage**: Kubernetes Secret (referenced in Gateway configuration)
- **Trust**: Trusted by all modern browsers and HTTP clients

### Client Certificates (mTLS only)
When using mTLS, you manage client certificates:
- **For production**: Use certificates from a trusted CA
- **For development/testing**: Self-signed certificates are acceptable
- **Distribution**: You provide certificates to clients
- **Trust configuration**: You configure the Gateway to trust your client CA