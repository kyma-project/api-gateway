# Configure mTLS Authentication Using Gardener-Managed Certificates

## Context

tbd

## Prerequisites

tbd

## Procedure
1. Create a namespace with enabled Istio sidecar proxy injection.
   
    ```bash
    kubectl create ns mtls-test
    kubectl label namespace mtls-test istio-injection=enabled --overwrite
    ```

2. Create a Secret containing credentials for your DNS cloud service provider.
        
    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/cert-management?tab=readme-ov-file#using-commonname-and-optional-dnsnames).

    See an example Secret for AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

    ```bash
    apiVersion: v1
    kind: Secret
    metadata:
      name: aws-credentials
      namespace: mtls-test
    type: Opaque
    data:
      AWS_ACCESS_KEY_ID: ...
      AWS_SECRET_ACCESS_KEY: ...
      # optionally specify the region
      #AWS_REGION: {YOUR_SECRET_ACCESS_KEY
      # optionally specify the token
      #AWS_SESSION_TOKEN: ...
    EOF
    ```

3. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.
   
   See an example Secret for AWS Route 53 DNS provider and the domain `my.domain.com`:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws-provider
      namespace: mtls-test
    annotations:
      dns.gardener.cloud/class: garden
    spec:
      type: aws-route53
      secretRef:
        name: aws-credentials
      domains:
        include:
        - my-domain.com
    EOF
    ```

4. Get the external access point of the `istio-ingressgateway` Service.

    ```bash
    export LOAD_BALANCER_ADDRESS==$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'); [ -z "$LOAD_BALANCER_ADDRESS=" ] && export LOAD_BALANCER_ADDRESS==$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    ```
    For GCP, the command gets the load balancer's IP adress. For AWS, the command gets the load balancer's hostname.

5. Create a DNSEntry resource.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
    name: dns-entry
    namespace: mtls-test
    annotations:
        dns.gardener.cloud/class: garden
    spec:
    dnsName: "${GATEWAY_DOMAIN}"
    ttl: 600
    targets:
        - "${LOAD_BALANCER_ADDRESS}"
    EOF
    ```

5. Create the server's certificate.
    
    You use a Certificate resource to request and manage certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you control the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret, as specified in the Certificate resource.

    Option | Description
    ---------|----------
    {GATEWAY_SECRET} | The name of the Secret that Gardener creates. It contains your certificate for the domain specified in the Certificate resource.
    {DOMAIN_NAME} | The domain name for which you request the certificate. For example, `my-domain.c1234.kyma.ondemand.com`.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: istio-system
    spec:
      secretName: {GATEWAY_SECRET}
      commonName: {GATEWAY_DOMAIN}
      issuerRef:
        name: garden
    EOF
    ```

6. Verify that the Scret with Gateway certificates is created:
   
    ```bash
    kubectl get secret -n istio-system "${GATEWAY_SECRET}"
    ```

7. Generate Client certificates

8. Create a Secret with Client CA Cert for mTLS Gateway. For more information on the convention that the Secret must use, see [Key Convention](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).

    ```bash
    kubectl create secret generic -n istio-system "${GATEWAY_SECRET}-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
    ```

9. Create an mTLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
    name: kyma-mtls-gateway
    namespace: test
    spec:
    selector:
        app: istio-ingressgateway
        istio: ingressgateway
    servers:
        - port:
            number: 443
            name: mtls
            protocol: HTTPS
        tls:
            mode: MUTUAL
            credentialName: {GATEWAY_SECRET}
        hosts:
            - {GATEWAY_DOMAIN}
    EOF
    ```
10. Create a sample Deployment.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: httpbin
    namespace: test
    ---
    apiVersion: v1
    kind: Service
    metadata:
    name: httpbin
    namespace: test
    labels:
        app: httpbin
        service: httpbin
    spec:
    ports:
    - name: http
        port: 8000
        targetPort: 80
    selector:
        app: httpbin
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: httpbin
    namespace: test
    spec:
    replicas: 1
    selector:
        matchLabels:
        app: httpbin
        version: v1
    template:
        metadata:
        labels:
            app: httpbin
            version: v1
        spec:
        serviceAccountName: httpbin
        containers:
        - image: docker.io/kennethreitz/httpbin
            imagePullPolicy: IfNotPresent
            name: httpbin
            ports:
            - containerPort: 80
    ```

11. Create an APIRule CR.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
    name: httpbin-mtls
    namespace: test
    spec:
    gateway: test/kyma-mtls-gateway
    hosts:
        - {WORKLOAD_DOMAIN}
    rules:
        - methods:
            - GET
        noAuth: true
        path: /*
        timeout: 300
        request:
            headers:
            X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'
            X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'
            X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'
    service:
        name: httpbin
        port: 8000
    ```

12. Connect to the workload.
    
    ```bash
    curl --fail --verbose \ --key "${CLIENT_CERT_KEY_FILE}" \ --cert "${CLIENT_CERT_CRT_FILE}" \ "https://${WORKLOAD_DOMAIN}/headers?show_env==true"
    ```

