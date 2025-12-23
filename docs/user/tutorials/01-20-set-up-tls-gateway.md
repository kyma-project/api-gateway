# Configure TLS Gateway in SAP BTP, Kyma Runtime
Learn how to configure a TLS Gateway in SAP BTP, Kyma runtime using Gardener-managed Let's Encrypt certificates.

## Context

In this procedure, you set up a TLS Gateway that secures communication between clients and your workloads. The server certificate is automatically provisioned and managed through Gardener's Certificate custom resource (CR), which requests a publicly trusted certificate from Let's Encrypt using the ACME protocol.

Simple TLS provides server-side authentication only, meaning clients verify the server's identity using its certificate, but the server does not authenticate clients.

## Prerequisites

- You have Istio and API Gateway modules in your cluster.
- For setting up the mTLS Gateway, you must prepare the domain name available in the public DNS zone.
- You must supply credentials for a DNS provider supported by Gardener so the ACME DNS challenge can be completed during certificate issuance. For the list of supported providers, see [External DNS Management Guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management).

## Procedure

<!-- tabs:start -->
#### **Custom Domain**

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```

2. Export the following domain names as environment variables. Replace `my-own-domain.example.com` with the name of your domain. You can adjust these values as needed.

    ```bash
    PARENT_DOMAIN="my-own-domain.example.com"
    SUBDOMAIN="tls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    echo "Parent Domain: ${PARENT_DOMAIN}"
    echo "Subdomain: ${SUBDOMAIN}"
    echo "Gateway Domain: ${GATEWAY_DOMAIN}"
    echo "Workload Domain: ${WORKLOAD_DOMAIN}"
    ```

    | Placeholder | Example domain name | Description |
    |---------|----------|---------|
    | **PARENT_DOMAIN** | `my-own-domain.example.com` | The domain name available in the public DNS zone. |
    | **SUBDOMAIN** | `tls.my-own-domain.example.com` | A subdomain created under the parent domain, specifically for the TLS Gateway. |
    | **GATEWAY_DOMAIN** | `*.tls.my-own-domain.example.com` | A wildcard domain covering all possible subdomains under the TLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.tls.my-own-domain.example.com`, `test.httpbin.tls.my-own-domain.example.com`) without creating separate Gateway rules for each one.|
    | **WORKLOAD_DOMAIN** | `httpbin.tls.my-own-domain.example.com` | The specific domain assigned to your workload. |

3. Create a Secret containing credentials for your DNS cloud service provider.

    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management).
    See an example Secret for the AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

    ```bash
    apiVersion: v1
    kind: Secret
    metadata:
      name: aws-credentials
      namespace: test
    type: Opaque
    data:
      AWS_ACCESS_KEY_ID: ...
      AWS_SECRET_ACCESS_KEY: ...
      # Optionally, specify the region
      #AWS_REGION: {YOUR_SECRET_ACCESS_KEY}
      # Optionally, specify the token
      #AWS_SESSION_TOKEN: ...
    ```

    To verify that the Secret is created, run:
   
    ```bash
    kubectl get secret -n test {SECRET_NAME}
    ```

4. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.

   See an example Secret for AWS Route 53 DNS provider:

    ```yaml
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws
      namespace: test
    spec:
      type: aws-route53
    secretRef:
        name: aws-credentials
      domains:
        include:
        - "${PARENT_DOMAIN}"
    ```

    To verify that the DNSProvider is created, run:
   
    ```bash
    kubectl get DNSProvider -n test {DNSPROVIDER_NAME}
    ```

5. Get the external access point of the `istio-ingressgateway` Service. The external access point is either stored in the ingress Gateway's **ip** field (for example, on GCP) or in the ingress Gateway's **hostname** field (for example, on AWS).

    ```bash
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [[ -z $LOAD_BALANCER_ADDRESS ]]; then
        LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    ```

6. Create a DNSEntry resource.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
      name: dns-entry
      namespace: test
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      dnsName: "${GATEWAY_DOMAIN}"
      ttl: 600
      targets:
        - "${LOAD_BALANCER_ADDRESS}"
    EOF
    ```

    To verify that the DNSEntry is created, run:
   
    ```bash
    kubectl get DNSEntry -n test dns-entry
    ```

7. Create the server's certificate.
    
   You use a Certificate CR to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate CR, one of Gardener's operators detects it and creates an [ACME](https://letsencrypt.org/how-it-works/) request to Let's Encrypt requesting certificate for the specified domain names. The issued certificate is stored in an automatically created Kubernetes Secret, which name you specify in the Certificate's secretName field. For more information, see [Manage certificates with Gardener for public domain](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/request_cert/).

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: "istio-system"
    spec:
      secretName: custom-tls-secret
      commonName: "${GATEWAY_DOMAIN}"
      issuerRef:
        name: garden
    EOF
    ```
  
    To verify that the Secret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system custom-tls-secret
    ```

9.  Create a TLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: custom-tls-gateway
      namespace: test
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: tls
            protocol: TLS
          tls:
            mode: SIMPLE
            credentialName: custom-tls-secret
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
    
    To verify that the TLS Gateway is created, run:
   
    ```bash
    kubectl get gateway -n test custom-tls-gateway
    ```

<!-- tabs:start -->
#### **Default Domain**

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```

2. Export the following domain names as enviroment variables. You can adjust the prefixes as needed.

    ```bash
    PARENT_DOMAIN=$(kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts[0]}' | sed 's/\*\.//')
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    echo "Parent Domain: ${PARENT_DOMAIN}"
    echo "Subdomain: ${SUBDOMAIN}"
    echo "Gateway Domain: ${GATEWAY_DOMAIN}"
    echo "Workload Domain: ${WORKLOAD_DOMAIN}"
    ```

    | Placeholder         | Example domain name                                | Description                                                                                                                                                                                                                                                                                                                                       |
    |---------------------|----------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **PARENT_DOMAIN**   | `my-default-domain.kyma.ondemand.com`              | The default domain of your Kyma cluster retrieved from the default TLS Gateway `kyma-gateway`.                                                                                                                                                                                                                                                    |
    | **SUBDOMAIN**       | `mtls.my-default-domain.kyma.ondemand.com`         | A subdomain created under the parent domain, specifically for the mTLS Gateway. Having a separate subdomain is required if you use the default domain of your Kyma cluster, as the parent domain name is already assigned to the TLS Gateway `kyma-gateway` installed in your cluster by default.                                                 |
    | **GATEWAY_DOMAIN**  | `*.mtls.my-default-domain.kyma.ondemand.com`       | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-default-domain.kyma.ondemand.com`, `test.httpbin.mtls.my-default-domain.kyma.ondemand.com`) without creating separate Gateway rules for each one. |
    | **WORKLOAD_DOMAIN** | `httpbin.mtls.my-default-domain.kyma.ondemand.com` | The specific domain assigned to your sample workload (HTTPBin Service) in this tutorial.                                                                                                                                                                                                                                                          |

3.  Create a TLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: custom-tls-gateway
      namespace: test
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: tls
            protocol: TLS
          tls:
            mode: SIMPLE
            credentialName: kyma-gateway-certs
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
    
    To verify that the TLS Gateway is created, run:
   
    ```bash
    kubectl get gateway -n test custom-tls-gateway
    ```

<!-- tabs:end -->