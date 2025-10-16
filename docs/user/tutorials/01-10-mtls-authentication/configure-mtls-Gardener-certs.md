# Configure mTLS Authentication Using Gardener-Managed Certificates

## Context

tbd

## Prerequisites

tbd

## Procedure
1. Create a namespace with enabled Istio sidecar proxy injection.
   
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as enviroment variables:

    ```bash
    PARENT_DOMAIN="my-own-domain.kyma.ondemand.com"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `my-own-domain.kyma.ondemand.com` | The domain name available in the public DNS zone. You can either use your custom domain
    **SUBDOMAIN** | `mtls.my-own-domain.kyma.ondemand.com` | A dedicated subdomain created under the parent domain, specifically for the mTLS Gateway. Choosing a subdomain is required if you use the default domain of your Kyma cluster, as the parent domain name is already assigned to the TLS Gateway kyma-gateway installed in your cluster by default.
    **GATEWAY_DOMAIN** | `*.mtls.my-own-domain.kyma.ondemand.com` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-own-domain.kyma.ondemand.com`, `test.httpbin.mtls.my-own-domain.kyma.ondemand.com`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.my-own-domain.kyma.ondemand.com` | The specific domain assigned to your sample workload (HTTPBin service) in this tutorial.

3. Create a Secret containing credentials for your DNS cloud service provider.
        
    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/cert-management?tab=readme-ov-file#using-commonname-and-optional-dnsnames).

    See an example Secret for AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

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
      #AWS_REGION: {YOUR_SECRET_ACCESS_KEY
      # Optionally, specify the token
      #AWS_SESSION_TOKEN: ...
    EOF
    ```

4. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.
   
   See an example Secret for AWS Route 53 DNS provider and the domain `my.domain.com`:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws
      namespace: default
    spec:
      type: aws-route53
      ecretRef:
        name: aws-credentials
      domains:
        include:
        - "${PARENT_DOMAIN}"
    EOF
    ```

5. Get the external access point of the `istio-ingressgateway` Service.

    ```bash
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ "$LOAD_BALANCER_ADDRESS" == "" ]; then
    echo "Load Balancer IP address not found, get the host name instead"
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    if [ "$LOAD_BALANCER_ADDRESS" == "" ]; then
    echo "Can't get Load Balancer address!"
    exit 1
    fi
    ```
    For GCP, the command gets the load balancer's IP adress. For AWS, the command gets the load balancer's hostname.

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

7. Create the server's certificate.
    
    You use a Certificate resource to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you own the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret `{GATEWAY_SECRET}`, as specified in the Certificate resource.

    ```bash
    export GATEWAY_SECRET=kyma-mtls
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: "istio-system"
    spec:
      secretName: "${GATEWAY_SECRET}"
       commonName: "${GATEWAY_DOMAIN}"
      issuerRef:
        name: garden
    EOF
    ```
    To verify that the Scret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system "${GATEWAY_SECRET}"
    ```

8. Prepare the client's certificates.

   To illustrate the process, this procedure uses self-signed client certificates. In production, ...
   
   1. Create the client's root CA.
    ```bash
    CLIENT_ROOT_CA_CN="ML Client Root CA"
    CLIENT_ROOT_CA_ORG="ML Client Org"
    CLIENT_ROOT_CA_KEY_FILE="${CLIENT_ROOT_CA_CN}.key"
    CLIENT_ROOT_CA_CRT_FILE="${CLIENT_ROOT_CA_CN}.crt"
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
    ```

    2. Create the client's certificate.
    ```bash
    CLIENT_CERT_CN="ML Client Curl"
    CLIENT_CERT_ORG="ML Client Org"
    CLIENT_CERT_CRT_FILE="${CLIENT_CERT_CN}.crt"
    CLIENT_CERT_CSR_FILE="${CLIENT_CERT_CN}.csr"
    CLIENT_CERT_KEY_FILE="${CLIENT_CERT_CN}.key"
    openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
    ``` 

    3. Sign the client's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"  
    ``` 

   2. Generate the P12 file.
       ```bash
       CLIENT_CERT_P12_FILE="${CLIENT_CERT_CN}.p12"
       openssl pkcs12 -export -out "${CLIENT_CERT_P12_FILE}" -inkey "${CLIENT_CERT_KEY_FILE}" -in "${CLIENT_CERT_CRT_FILE}" -certfile "${CLIENT_ROOT_CA_CRT_FILE}" -passout pass:{SPECIFY_A_PASSWORD}
       ``` 

9.  Create a Secret with Client CA Cert for mTLS Gateway. For more information on the convention that the Secret must use, see [Key Convention](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).

    ```bash
    kubectl create secret generic -n istio-system "${GATEWAY_SECRET}-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
    ```

10. Create an mTLS Gateway.
 
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
            credentialName: "${GATEWAY_SECRET}"
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
11. Create a sample Deployment.

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
EOF
    ```

12. Create an APIRule CR.

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
    - "${WORKLOAD_DOMAIN}"
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
EOF
    ```

13. Connect to the workload.
    
    ```bash
    curl --fail --verbose \
      --key "${CLIENT_CERT_KEY_FILE}" \
      --cert "${CLIENT_CERT_CRT_FILE}" \
      "https://${WORKLOAD_DOMAIN}/headers?show_env==true"
    ```

