# Choosing a Domain for Exposing Workloads

### Using Your Custom Domain

If you want to expose workloads under a custom domain that is not managed by the default provider but by a custom one, you must first register this DNS provider in your Kyma runtime cluster.
This tutorial shows how to set up a custom domain and prepare a certificate required for exposing a workload. It uses the Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components.

### Using the Default Domain of Your Kyma Cluster
  
When you create a SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.

Morover, you can use the default domain to set up an Ingress gateway, and exposed your applications under this host. By default, a simple TLS Gateway `kyma-gateway` is configured under the default wildcard domain of your Kyma cluster. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or get the domain name from the default simple TLS Gateway:

```bash
kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
```

You can request any subdomain of the assigned default domain and use it to create a TLS or mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more.

To learn how to do this, follow [](#set-up-an-external-dns-provider). The continue with [](#use-gardener-managed-certificates)

- Use the domain local.kyma.dev for local Kyma installations on k3d.

- Use your custom domain registered in an external DNS provider.


If you decide to use the first or the second approach, no additional setup is required. If you decide to use the third aproach and set up your custom domain, you must first register your domain and DNS provider in Kyma. Follow the procedure to learn how to do this.

### Using the Domain local.kyma.dev for Local Installations on k3d


## Prerequisites

* You have a custom domain registered in one of DNS providers supported by Gardener.

## Procedure



### Next Steps
[Set up a TLS Gateway](./01-20-set-up-tls-gateway.md) or [set up an mTLS Gateway](./01-30-set-up-mtls-gateway.md).

For more examples of CRs for Services and Ingresses, see the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples).
