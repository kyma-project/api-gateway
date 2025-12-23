# Choosing a Gateway

This document explains what Gateways are in Kubernetes, why you might need a custom Gateway, and how to choose the right Gateway configuration for your use case in Kyma.

## What is a Gateway?

A Gateway manages inbound and outbound traffic for the service mesh. It acts as a load balancer operating at the edge of the mesh, receiving incoming HTTP/HTTPS connections.

The Gateway performs the following functions:
- Decrypts incoming HTTPS traffic and encrypts responses
- Directs requests to the appropriate services based on traffic redirection rules
- Applies authentication, authorization, and traffic management rules in a centralized location
- Acts as the entry point for traffic from outside the cluster to reach services inside

## Choosing the Domain for a Gateway

A Gateway host is the domain name or hostname that the Gateway listens on to accept incoming traffic. It defines which domain names the Gateway responds to. As for the domain name, you can choose from the following options:

- Use your custom domain.
    
    To use a custom domain, you must own the DNS zone and supply credentials for a provider supported by Gardener so the ACME DNS challenge can be completed. For this, you must first register this DNS provider in your Kyma runtime cluster and create a DNS entry resource.

- Use the default domain of your Kyma cluster.
    
    When you create an SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.
    
    By default, the default Ingress Gateway `kyma-gateway` is configured under this domain. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or get the domain name from the default simple TLS Gateway: 
    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts[0]}'
    ```

    You can request any subdomain of the assigned default domain and use it to create a TLS or mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more. If you use the Kyma runtime default domain, Gardener’s issuer can issue certificates for subdomains of that domain without additional DNS delegation.

## Choosing the TLS Mode

### Simple TLS Options

In Simple TLS mode, the server presents a certificate to prove its identity to clients, but clients don't need to provide certificates. This is the most common configuration for public-facing websites and APIs.

| Option | Domain | When to Use |
|--------|--------|-------------|
| Default Kyma Gateway<br/>`kyma-system/kyma-gateway` | Kyma default | Pre-configured and ready to use immediately, recommended for development |
| Custom TLS on Kyma Domain | Kyma subdomain | Gateway isolation, no DNS setup |
| Custom TLS on Custom Domain | Your domain | Production, full control over domain name |

See [TLS Gateway Tutorial](./tutorials/01-05-configure-tls.md).

### Mutual TLS Options
In Mutual TLS (mTLS) mode, both the server and client present certificates to verify each other's identity. This provides stronger authentication by ensuring only clients with valid certificates can connect.

| Option | Domain | Setup Complexity | When to Use |
|--------|--------|------------------|-------------|
| Custom mTLS on Kyma Domain | Kyma subdomain | B2B APIs, no custom domain |
| Custom mTLS on Custom Domain | Your domain | Highest security, custom domain |

See [mTLS Gateway](./tutorials/01-10-mtls-authentication/configure-mtls-Gardener-certs.md).