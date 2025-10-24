# Configure mTLS Authentication Using Gardener-Managed Certificates
Learn how to configure mutual TLS (mTLS) in SAP BTP, Kyma runtime using Gardener-managed Let's Encrypt server certificates and client certificates that you supply.

## Context
mTLS (mutual TLS) provides two‑way authentication: the client verifies the server's identity and the server verifies the client's identity. To enforce this authentication, the mTLS Gateway requires the following values: 
- The server private key
- The server certificate chain (server certificate plus any intermediate certificate authorities (CAs))
- The client root CA used to validate presented client certificates
Each client connecting through the mTLS Gateway must have a valid client certificate and key and trust the server's root CA.

Specifically, in this procedure, you generate certificates using the following approach:
- Gardener’s Certificate custom resource (CR) requests a publicly trusted server certificate from Let’s Encrypt and creates a Secret that stores the certificate and private key. Therefore, the clients must trust Let's Encrypt, which is the CA that signs the server's certificate. Most modern HTTP clients already trust Let's Encrypt.
- Client certificates are self-signed. For production use, it's strongly advised to use certificates issued by a trusted CA instead.

## Prerequisites

- You have a SAP BTP, Kyma runtime instance with Istio and API Gateway modules added. The Istio and API Gateway modules are added to your Kyma cluster by default.
- For setting up the mTLS Gateway, you must prepare the domain name available in the public DNS zone. You can use one of the following approaches:

  - Use your custom domain.
    
    For a custom domain you must own the DNS zone and supply credentials for a provider supported by Gardener so the ACME DNS challenge can be completed. For this, you must first register this DNS provider in your Kyma runtime cluster and create a DNS entry resource.

  - Use the default domain of your Kyma cluster.
    
    When you create a SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.
    
    By default, the default Ingress Gateway `kyma-gateway` is configured under this domain. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or get the domain name from the default simple TLS Gateway: 
    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
    ```

    You can request any subdomain of the assigned default domain and use it to create a TLS or mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more. If you use the Kyma runtime default domain, Gardener’s issuer can issue certificates for subdomains of that domain without additional DNS delegation.
  
## Procedure

<!-- tabs:start -->
### **Custom Domain**

1. Create a namespace with enabled Istio sidecar proxy injection.
   
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as enviroment variables. Replace `my-own-domain.example.com` with the name of your domain.

    ```bash
    PARENT_DOMAIN="my-own-domain.example.com"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `my-own-domain.example.com` | The domain name available in the public DNS zone.
    **SUBDOMAIN** | `mtls.my-own-domain.example.com` | A subdomain created under the parent domain, specifically for the mTLS Gateway.
    **GATEWAY_DOMAIN** | `*.mtls.my-own-domain.example.com` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-own-domain.example.com`, `test.httpbin.mtls.my-own-domain.example.com`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.my-own-domain.example.com` | The specific domain assigned to your workload.

3. Create a Secret containing credentials for your DNS cloud service provider.
        
    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/external-dns-management).

    See an example Secret for AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.

    <!-- tabs:start -->
    ### **AWS Route 53**

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
      #AWS_REGION: ...
      # Optionally, specify the token
      #AWS_SESSION_TOKEN: ...
    EOF
    ```

    ### **Google**

    apiVersion: v1
    kind: Secret
    metadata:
      name: google-credentials
      namespace: default
    type: Opaque
    data:
      # replace '...' with json key from service account creation (encoded as base64)
      # see https://cloud.google.com/iam/docs/creating-managing-service-accounts
      serviceaccount.json: ...
    <!-- tabs:end -->

4. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.
   
   See an example Secret for AWS Route 53 DNS provider:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws
      namespace: default
    spec:
      type: aws-route53
      secretRef:
        name: aws-credentials
      domains:
        include:
        - "${PARENT_DOMAIN}"
    EOF
    ```

5. Get the external access point of the `istio-ingressgateway` Service. The external access point is either stored in the **ip** field of `istio-ingressgateway` (for example, on GCP) or in the **hostname** field of `istio-ingressgateway` (for example, on AWS).

    ```bash
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [[ -z $LOAD_BALANCER_ADDRESS ]]; then
        LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    echo "The load balancer address is ${LOAD_BALANCER_ADDRESS}"
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

7. Create the server's certificate.
   
    You use a Certificate CR to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate CR, one of Gardener's operators detects it and creates an [ACME](https://www.google.com/url?sa=t&source=web&rct=j&opi=89978449&url=https://letsencrypt.org/how-it-works/&ved=2ahUKEwiRhM_VrruQAxWFPxAIHbePM38QFnoECC0QAQ&usg=AOvVaw25RIWodU2kz362IWS5BbJs) request to Let's Encrypt requesting certificate for the specified domain names. The issued certificate is stored in an automatically created Kubernetes Secret, which name you specify in the Certificate's **secretName** field. For more information, see [Manage certificates with Gardener for public domain](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/request_cert/).

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: "istio-system"
    spec:
      secretName: kyma-mtls
      commonName: "${GATEWAY_DOMAIN}"
      issuerRef:
        name: garden
    EOF
    ```
    To verify that the Secret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system kyma-mtls
    ```

8. Prepare the client's certificates.

    To illustrate the process, this procedure uses self-signed client certificates.
    >[!WARNING]
    > For production deployments, use trusted certificate authorities to ensure proper security and automatic certificate management.
   
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
    
    4. Generate a PKCS#12 file that bundles the client’s private key, client certificate, and the client root CA certificate into a single file. 
   
       ```bash
       CLIENT_CERT_P12_FILE="${CLIENT_CERT_CN}.p12"
       openssl pkcs12 -export -out "${CLIENT_CERT_P12_FILE}" -inkey "${CLIENT_CERT_KEY_FILE}" -in "${CLIENT_CERT_CRT_FILE}" -certfile "${CLIENT_ROOT_CA_CRT_FILE}" -passout pass:{SPECIFY_A_PASSWORD}
       ``` 

9.  Create a Secret with Client CA Cert for mTLS Gateway. For more information on the convention that the Secret must use, see [Key Convention](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).

    ```bash
    kubectl create secret generic -n istio-system "kyma-mtls-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
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
            credentialName: kyma-mtls
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```

11. Create a sample HTTPBin Deployment.

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

12. To expose the sample HTTPBin Deployment, create an APIRule custom resource. 
    The APIRule must include the `X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'`, `X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'`, and `X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'` headings. These headers are necessary to ensure that the backend service receives the authenticated client's identity.

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

13. Test the mTLS connection.
    
    1. Run the following curl command:
    
    ```bash
    curl --fail --verbose \
      --key "${CLIENT_CERT_KEY_FILE}" \
      --cert "${CLIENT_CERT_CRT_FILE}" \
      "https://${WORKLOAD_DOMAIN}/headers?show_env=true"
    ```

    If successful, you get code `200` in response. The configured headers are also populated.
    
    2. To test the connection using your browser, import the client certificates into your operating system or browser. For Chrome, you can use the generated PKCS#12 file. Then, open `https://{WORKLOAD_DOMAIN}`.

### **Default Domain**


1. Create a namespace with enabled Istio sidecar proxy injection.
   
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as enviroment variables. Replace `my-own-domain.kyma.ondemand.com` with the name of your domain.

    ```bash
    PARENT_DOMAIN="my-own-domain.kyma.ondemand.com"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `my-default-domain.kyma.ondemand.com` | The default domain of your Kyma cluster.
    **SUBDOMAIN** | `mtls.my-default-domain.kyma.ondemand.com` | A subdomain created under the parent domain, specifically for the mTLS Gateway. Having a separate subdomain is required if you use the default domain of your Kyma cluster, as the parent domain name is already assigned to the TLS Gateway `kyma-gateway` installed in your cluster by default.
    **GATEWAY_DOMAIN** | `*.mtls.my-default-domain.kyma.ondemand.com` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-default-domain.kyma.ondemand.com`, `test.httpbin.mtls.my-default-domain.kyma.ondemand.com`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.my-default-domain.kyma.ondemand.com` | The specific domain assigned to your sample workload (HTTPBin service) in this tutorial.

3. Create the server's certificate.
    
    You use a Certificate resource to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you own the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret, which name is specified in the Certificate's **secretName** field.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: domain-certificate
      namespace: "istio-system"
    spec:
      secretName: kyma-mtls
      commonName: "${GATEWAY_DOMAIN}"
      issuerRef:
        name: garden
    EOF
    ```
    To verify that the Scret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system kyma-mtls
    ```

4. Prepare the client's certificates.

    To illustrate the process, this procedure uses self-signed client certificates. 
    >[!WARNING]
    > For production deployments, use trusted certificate authorities to ensure proper security and automatic certificate management.
   
   1. Create the client's root CA.
      ```bash
      CLIENT_ROOT_CA_CN="Example Client Root CA"
      CLIENT_ROOT_CA_ORG="Example Client Org"
      CLIENT_ROOT_CA_KEY_FILE="${CLIENT_ROOT_CA_CN}.key"
      CLIENT_ROOT_CA_CRT_FILE="${CLIENT_ROOT_CA_CN}.crt"
      openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
      ```

    2. Create the client's certificate.
        ```bash
        CLIENT_CERT_CN="Example Client Curl"
        CLIENT_CERT_ORG="Example Client Org"
        CLIENT_CERT_CRT_FILE="${CLIENT_CERT_CN}.crt"
        CLIENT_CERT_CSR_FILE="${CLIENT_CERT_CN}.csr"
        CLIENT_CERT_KEY_FILE="${CLIENT_CERT_CN}.key"
        openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
        ``` 

    3. Sign the client's certificate.
    
        ```bash
        openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"  
        ``` 
    
    4. Generate a PKCS#12 file that bundles the client’s private key, client certificate, and the client root CA certificate into a single file. 
   
       ```bash
       CLIENT_CERT_P12_FILE="${CLIENT_CERT_CN}.p12"
       openssl pkcs12 -export -out "${CLIENT_CERT_P12_FILE}" -inkey "${CLIENT_CERT_KEY_FILE}" -in "${CLIENT_CERT_CRT_FILE}" -certfile "${CLIENT_ROOT_CA_CRT_FILE}" -passout pass:{SPECIFY_A_PASSWORD}
       ``` 

5.  Create a Secret with Client CA Cert for mTLS Gateway. For more information on the convention that the Secret must use, see [Key Convention](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).

    ```bash
    kubectl create secret generic -n istio-system "kyma-mtls-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
    ```

6.  Create an mTLS Gateway.
 
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
            credentialName: kyma-mtls
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```

7.  Create a sample HTTPBin Deployment.

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

8.  To expose the sample HTTPBin Deployment, create an APIRule custom resource. 
    The APIRule must include the `X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'`, `X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'`, and `X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'` headings. These headers are necessary to ensure that the backend service receives the authenticated client's identity. Specifically, they provide the client certificate's subject, issuer DN, and SAN values, respectively.

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

9.  Test the mTLS connection.
    
    1. Run the following curl command:
    
        ```bash
        curl --fail --verbose \
          --key "${CLIENT_CERT_KEY_FILE}" \
          --cert "${CLIENT_CERT_CRT_FILE}" \
          "https://${WORKLOAD_DOMAIN}/headers?show_env=true"
        ```

        If successful, you get code `200` in response. The **X-Forwarded-Client-Cert** heading contains your client certificate.
    
    2. To thest the connection using your browser, import the client certificates into your operating system or browser. For Chrome, you can use the generated PKCS#12 file. Then, open `https://{WORKLOAD_DOMAIN}`.

<!-- tabs:end -->