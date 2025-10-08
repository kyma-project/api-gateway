# Setting Up a Custom Domain for a Workload

This tutorial shows how to set up a custom domain and prepare a certificate required for exposing a workload. It uses the Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components.

## Context
To expose your workload to the internet, first you need a domain name under which the workload is accessble. In SAP BTP, Kyma runtime, you can use one of the following approaches:

- Use the default domain of your Kyma runtime cluster.
  
    When you create a SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.

    Morover, you can use the default domain to set up an Ingress gateway, and exposed your applications under this host. By default, a simple TLS Gateway `kyma-gateway` is configured under the default wildcard domain of your Kyma cluster. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or fetch the domain name from the default simple TLS Gateway:
    
    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
    ```
    
    You can request any subdomain of the assigned default domain and use it to create an mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more.

    To learn how to do this, follow [](#set-up-an-external-dns-provider). The continue with [](#use-gardener-managed-certificates)

- Use the domain local.kyma.dev for local Kyma installations on k3d.

- Use your custom domain registered in an external DNS provider.
  
    If you want to expose workloads under a custom domain that is not managed by the default provider but by a custom one, you must first register this DNS provider in your Kyma runtime cluster.

If you decide to use the first or the second approach, no additional setup is required. If you decide to use the third aproach and set up your custom domain, you must first register your domain and DNS provider in Kyma. Follow the procedure to learn how to do this.

## Prerequisites

* You have a custom domain registered in one of DNS providers supported by Gardener.

## Procedure

1. In a namespace of your choice, create a Secret containing credentials for your DNS cloud service provider.
        
    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/cert-management?tab=readme-ov-file#using-commonname-and-optional-dnsnames).

    See an example Secret for AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

    ```bash
    apiVersion: v1
    kind: Secret
    metadata:
      name: aws-credentials
      namespace: default
    type: Opaque
    data:
      AWS_ACCESS_KEY_ID: ...
      AWS_SECRET_ACCESS_KEY: ...
      # optionally specify the region
      #AWS_REGION: ... 
      # optionally specify the token
      #AWS_SESSION_TOKEN: ...
    EOF
    ```

2. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.
   
   See an example Secret for AWS Route 53 DNS provider and the domain `my.domain.com`:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws-provider
      namespace: default
    annotations:
      dns.gardener.cloud/class: garden
    spec:
      type: aws-route53
      secretRef:
        name: aws-credentials
      domains:
        include:
        - my.domain.com
    EOF
    ```

3. Get the external access point of the `istio-ingressgateway` Service.

    ```bash
    export INGRESS=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'); [ -z "$INGRESS" ] && export INGRESS=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    ```
    For GCP, you use an IP adress, and for AWS, you use the hostname.

4. Create a DNSEntry resource.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
      name: dns-entry
      namespace: default
    annotations:
        dns.gardener.cloud/class: garden
    spec:
      dnsName: "my.domain.com"
      ttl: 600
      targets:
        - $INGRESS
    EOF
    ```

### Next Steps
[Set up a TLS Gateway](./01-20-set-up-tls-gateway.md) or [set up an mTLS Gateway](./01-30-set-up-mtls-gateway.md).

For more examples of CRs for Services and Ingresses, see the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples).
