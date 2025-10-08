# Configure mTLS Authentication for Your Workloads

Configure mutual TLS (mTLS) authentication using Gardener-managed or self-signed certificates, configure Gateways and APIRules for mTLS authentication, and verify the mTLS connection.
- [Use SAP BTP, Kyma runtime and Gardener-managed Certificates](#use-gardener-managed-certificates)
- [Use k3d and Self-Signed Certificates](#use-k3d-and-self-signed-certificates)

# Use Gardener-managed Certificates

## Context
tbd

## Prerequisites
- You have an avaliable domain that you can use for setting up an mTLS Gateway and exposing your workload. You can either use the default domain of your Kyma cluster or a custom domain registered in an external DNS provider. If you choose to use the custom domain, first follow the guide [](./01-10-setup-custom-domain-for-workload.md).

## Procedure

1. Create the server's certificate.
    
    You use a Certificate resource to request and manage certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you control the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret, as specified in the Certificate resource.

    Option | Description
    ---------|----------
    {TLS_SECRET} | The name of the Secret that Gardener creates. It contains your certificate for the domain specified in the Certificate resource.
    {DOMAIN_NAME} | The domain name for which you request the certificate. For example, `my-domain.c1234.kyma.ondemand.com`.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: gardener-domain-cert
      namespace: istio-system
    spec:
      secretName: kyma-mtls-certs
      commonName: my.own.domain.kyma.ondemand.com
      issuerRef:
        name: garden
    EOF
    ```
    //Root CA: let's encrypt jest commonly trusted
    // public key servera i private key servera w sekrecie {TLS_SECRET}

2. Create a Secret for the mTLS Gateway.
// sekret ma mieć klucz publiczny servera, prywatny servera i root ca klienta
// czy root ca klienta trzeba dodać manualnie, czy stworzyć jeszcze jeden sekret
// jak to wygląda na prod env -> certyikaty servera są tworzone przez Gardener, certyfikaty klienta user musi mieć swoje??

    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs --from-file=cacert=cacert.crt
    ```
    ```bash
    kubectl patch secret kyma-mtls-certs -n istio-system \ --type='json' \ -p='[{"op": "add", "path": "/data/cacert", "value": "'$(base64 -w0 < cacert.crt)'"}]'
    ```

1. Create an mTLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
    name: kyma-mtls-gateway
    namespace: default
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
            credentialName: kyma-mtls-certs
        hosts:
            - "my.own.domain.kyma.ondemand.com"
    EOF
    ```

2. Create an APIRule CR.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
    name: {APIRULE_NAME}
    namespace: {APIRULE_NAMESPACE}
    spec:
    gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
    hosts:
        - {DOMAIN_NAME}
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
        name: {SERVICE_NAME}
        port: {SERVICE_PORT}
    EOF
    ```

3. Connect to the workload.
    
    ```bash
    curl --key "${CLIENT_CERT_KEY_FILE}" \
        --cert "${CLIENT_CERT_CRT_FILE}" \
        --cacert "${CLIENT_ROOT_CA_CRT_FILE}" \
        -ik -X GET https://{YOUR_DOMAIN}/headers
    ```

### Use k3d and Self-Signed Certificates

## Context
When using self-signed certificates for mTLS, you're creating a certification chain that consists of:
-  A root CA certificates for the client and the server that you create (acting as your own Certificate Authority)
- Server and client certificates that are signed by the respective root CA

This means you're establishing trust relationships without using a publicly trusted authority. Therefore, this approach is recommended for use in testing or development environments only. For production deployments, use trusted certificate authorities to ensure proper security and automatic certificate management (for example, Let's Encrypt used in the previous section).

## Prerequisites
- [k3d](https://k3d.io/stable/)
- OpenSSL

## Procedure
1. Create a Kyma cluster with the Istio and API Gateway modules added.
    ```bash
    k3d cluster delete kyma
    k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer  --image rancher/k3s:v1.31.9-k3s1 --k3s-arg "--disable=traefik@server:*"
    ```

2. Add the Istio and API Gateway modules.
    ```bash
    kubectl create ns kyma-system
    kubectl label namespace kyma-system istio-injection=enabled --overwrite
    kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
    kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
    kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
    kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml
    ```

3. Create a test namespace with Istio sidecar injection enabled.
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```

4. Create the server's root CA.

    ```bash
    export SERVER_ROOT_CA_CN="Example Server Root CA"
    export SERVER_ROOT_CA_ORG="Example Server Root Org"
    export SERVER_ROOT_CA_KEY_FILE=server-ca.key
    export SERVER_ROOT_CA_CRT_FILE=server-ca.crt
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${SERVER_ROOT_CA_ORG}/CN=${SERVER_ROOT_CA_CN}" -keyout "${SERVER_ROOT_CA_KEY_FILE}" -out "${SERVER_ROOT_CA_CRT_FILE}"
    ```
5. Create the server's certificate.
    
    ```bash
    export SERVER_CERT_CN="httpbin.local.kyma.dev"
    export SERVER_CERT_ORG="Example Server Org"
    export SERVER_CERT_CRT_FILE=server.crt
    export SERVER_CERT_CSR_FILE=server.csr
    export SERVER_CERT_KEY_FILE=server.key
    openssl req -out "${SERVER_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${SERVER_CERT_KEY_FILE}" -subj "/CN=${SERVER_CERT_CN}/O=${SERVER_CERT_ORG}"
    ```
6. Sign the server's certificate.
    ```bash
    openssl x509 -req -days 365 -CA "${SERVER_ROOT_CA_CRT_FILE}" -CAkey "${SERVER_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${SERVER_CERT_CSR_FILE}" -out "${SERVER_CERT_CRT_FILE}"
    ```
7. Create client's root CA.
    
    ```bash 
    export CLIENT_ROOT_CA_CN="Example Client Root CA"
    export CLIENT_ROOT_CA_ORG="Example Client Root Org"
    export CLIENT_ROOT_CA_KEY_FILE=client-ca.key
    export CLIENT_ROOT_CA_CRT_FILE=client-ca.crt
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
    ```
8. Create the client's certificate.
    
    ```bash
    export CLIENT_CERT_CN="Example Client Curl"
    export CLIENT_CERT_ORG="Example Client Org"
    export CLIENT_CERT_CRT_FILE=client.crt
    export CLIENT_CERT_CSR_FILE=client.csr
    export CLIENT_CERT_KEY_FILE=client.key
    openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
    ```

9.  Sign the client's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"
    ```

10. Create a Secret for the mTLS Gateway.
    
    ```bash
    kubectl create secret generic -n istio-system local-mtls-certs --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"  --from-file=key="${SERVER_CERT_KEY_FILE}" --from-file=cert="${SERVER_CERT_CRT_FILE}"
    ```
11. Create the mTLS Gateway.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: local-mtls-gateway
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
            credentialName: local-mtls-certs
          hosts:
            - "httpbin.local.kyma.dev"
    EOF
    ```
12. Create a HTTPBin Deployment.

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

13. Create an APIRule.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin-mtls
      namespace: test
    spec:
      gateway: test/local-mtls-gateway
      hosts:
        - "httpbin.local.kyma.dev"
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

14. Test the connection.
    
    ```bash
    curl --verbose  \
        --key "${CLIENT_CERT_KEY_FILE}" \
        --cert "${CLIENT_CERT_CRT_FILE}" \
        --cacert "${SERVER_ROOT_CA_CRT_FILE}" \
        "https://httpbin.local.kyma.dev/headers?show_env=true"
    ```
    If successful, you get code 200 in response. 