# Configure mTLS Authentication Locally on k3d
Learn how to set up mutual TLS (mTLS) authentication in a local Kyma environment using k3d.

## Context
mTLS (mutual TLS) provides two‑way authentication: the client verifies the server's identity and the server verifies the client's identity. To enforce this authentication, the mTLS Gateway requires three items: the server private key, the server certificate chain (server certificate plus any intermediate CAs), and the client root CA used to validate presented client certificates. Each client connecting through the mTLS Gateway must have a valid client certificate and key and trust the server's root CA.

To better illustrate the process, this procedure uses self-signed certificates. First, you create the server root CA, generate and sign the server certificate, and assemble the certificate chain so the gateway can present a valid chain to clients. Next, you create the client root CA and generate a client certificate that the server can validate.

When using self-signed certificates for mTLS, you act as your own CA and establish trust relationships without relying on a publicly trusted authority. Therefore, this approach is recommended for use in testing or development environments only. 

>[!WARNING]
> For production deployments, use trusted certificate authorities to ensure proper security and automatic certificate management.


## Prerequisites
- [k3d](https://k3d.io/stable/)
- [OpenSSL](https://openssl-library.org/)

## Procedure
1. Create a Kyma cluster with the Istio and API Gateway modules added.
    ```bash
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

3. Create a namespace with Istio sidecar injection enabled.
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
4. Export the following domain names as enviroment variables:

    ```bash
    PARENT_DOMAIN="local.kyma.dev"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `local.kyma.dev` | The main wildcard public domain for your local Kyma installation. The domain is registered in public DNS and points to the local host `127.0.0.1`. By default, this domain is used by the API Gateway module to configure the default TLS Gateway. To avoid conflicts and enable custom gateways, you must use a subdomain of this parent domain for your own Gateway.
    **SUBDOMAIN** | `mtls.local.kyma.dev` | A dedicated subdomain created under the parent domain, specifically for the mTLS Gateway. This isolates mTLS traffic and allows you to manage certificates and routing separately from the default Gateway.
    **GATEWAY_DOMAIN** | `*.mtls.local.kyma.dev` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.local.kyma.dev`, `test.httpbin.mtls.local.kyma.dev`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.local.kyma.dev` | The specific domain assigned to your sample workload (HTTPBin service) in this tutorial.

4. Create the server's root CA.

    ```bash
    SERVER_ROOT_CA_CN="ML Server Root CA"
    SERVER_ROOT_CA_ORG="ML Server Org"
    SERVER_ROOT_CA_KEY_FILE="${SERVER_ROOT_CA_CN}.key"
    SERVER_ROOT_CA_CRT_FILE="${SERVER_ROOT_CA_CN}.crt"
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${SERVER_ROOT_CA_ORG}/CN=${SERVER_ROOT_CA_CN}" -keyout "${SERVER_ROOT_CA_KEY_FILE}" -out "${SERVER_ROOT_CA_CRT_FILE}"
    ```
5. Create the server's certificate.
    
    ```bash
    SERVER_CERT_CN="${GATEWAY_DOMAIN}"
    SERVER_CERT_ORG="ML Server Org"
    SERVER_CERT_CRT_FILE="${SERVER_CERT_CN}.crt"
    SERVER_CERT_CSR_FILE="${SERVER_CERT_CN}.csr"
    SERVER_CERT_KEY_FILE="${SERVER_CERT_CN}.key"
    openssl req -out "${SERVER_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${SERVER_CERT_KEY_FILE}" -subj "/CN=${SERVER_CERT_CN}/O=${SERVER_CERT_ORG}"
    ```
6. Sign the server's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${SERVER_ROOT_CA_CRT_FILE}" -CAkey "${SERVER_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${SERVER_CERT_CSR_FILE}" -out "${SERVER_CERT_CRT_FILE}"
    ```

7. Create the server's certificate chain consisting of the server's certificate and the server's root CA.
   
    ```bash
    SERVER_CERT_CHAIN_FILE="${SERVER_CERT_CN}-chain.pem"
    cat "${SERVER_CERT_CRT_FILE}" "${SERVER_ROOT_CA_CRT_FILE}" > "${SERVER_CERT_CHAIN_FILE}"
    ```
8. Create a Secret for the mTLS Gateway with the server's key and certificate.
    
    ```bash
    GATEWAY_SECRET=kyma-mtls
    kubectl create secret tls -n istio-system "${GATEWAY_SECRET}" --key="${SERVER_CERT_KEY_FILE}" --cert="${SERVER_CERT_CHAIN_FILE}"
    ```

9. Create the client's root CA.
    
    ```bash 
    CLIENT_ROOT_CA_CN="ML Client Root CA"
    CLIENT_ROOT_CA_ORG="ML Client Org"
    CLIENT_ROOT_CA_KEY_FILE="${CLIENT_ROOT_CA_CN}.key"
    CLIENT_ROOT_CA_CRT_FILE="${CLIENT_ROOT_CA_CN}.crt"
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
    ```
10. Create the client's certificate.
    
    ```bash
    CLIENT_CERT_CN="ML Client Curl"
    CLIENT_CERT_ORG="ML Client Org"
    CLIENT_CERT_CRT_FILE="${CLIENT_CERT_CN}.crt"
    CLIENT_CERT_CSR_FILE="${CLIENT_CERT_CN}.csr"
    CLIENT_CERT_KEY_FILE="${CLIENT_CERT_CN}.key"
    openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
    ```

11. Sign the client's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"
    ```
12. Create a Secret for the mTLS Gateway containing the client's CA certificate. 
    The Secret must follow Istio convention. See [Key Formats](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).
    
    ```bash
    kubectl create secret generic -n istio-system "${GATEWAY_SECRET}-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
    ```
13. Create the mTLS Gateway.
    
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
14. Create a HTTPBin Deployment.

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

15. Create an APIRule.
    
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

16. Test the connection.

    1. Run the following curl command:
    
    ```bash
    curl --fail --verbose \
      --key "${CLIENT_CERT_KEY_FILE}" \
      --cert "${CLIENT_CERT_CRT_FILE}" \
      --cacert "${SERVER_ROOT_CA_CRT_FILE}" \
      "https://${WORKLOAD_DOMAIN}/headers?show_env==true"
    ```
    If successful, you get the following response:
    
    ```bash
    {
      "headers": {
        "Accept": "*/*",
        "Host": "httpbin.mtls.local.kyma.dev",
        "User-Agent": "curl/8.7.1",
        "X-Client-Ssl-Cn": "O=ML Client Org,CN=ML Client Curl",
        "X-Client-Ssl-Issuer": "CN=ML Client Root CA,O=ML Client Org",
        "X-Envoy-Attempt-Count": "1",
        "X-Envoy-External-Address": "10.42.0.1",
        "X-Forwarded-Client-Cert": "Hash=32137db958ee1c175cb6892431eb35067fc0a95513e3612d033a573005852fb9;Cert=\"-----BEGIN%20CERTIFICATE-----%0AMIIC2TCCAcECAQAwDQYJKoZIhvcNAQELBQAwNDEWMBQGA1UECgwNTUwgQ2xpZW50%0AIE9yZzEaMBgGA1UEAwwRTUwgQ2xpZW50IFJvb3QgQ0EwHhcNMjUxMDA4MTk1NjM4%0AWhcNMjYxMDA4MTk1NjM4WjAxMRcwFQYDVQQDDA5NTCBDbGllbnQgQ3VybDEWMBQG%0AA1UECgwNTUwgQ2xpZW50IE9yZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC%0AggEBAOQkN4w60Rf%2F8KtyZW9D1rVcxjnwR2GIOo6h0Zm9PFRtbYLT8WvTxO7V2SDK%0AmxfXRQlvuUWEW3XzoGZc0e%2FNTtV3ajswCM9A10wPCYvm%2Bv4BHNk%2FBuCf2jAmGoYS%0AO%2FtmHwiwaZo43pb6kW4wEk2POBYSPB4ekQW1H2X2RzSGXuuOFyr6%2BqL9RgldjNf9%0A3e29agdU6XbJzGCItlXrx5O0aSJaLEVyMhZKV%2BVq58I8mLzO6Nsl5IQ1SgrnmYxx%0Akf6P9myW9tnFaLH39Uv8NLBEd6yE46BGts8%2BdaSBfG0Xkska580NAWjQYetA0kFZ%0A6K%2FoOVqG8jxp9icjC8Cy0WXW6qkCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAqv%2BL%0A%2Bmu%2FGifiRQkNnhzmC4j6NzcfFbDkdxTtTzYHdtbrccWX5cj0mpKQEj1ymdm7uqx0%0A5ZlLe4sEQ91BsfsQFc62N3ucPhVOtka8UnyWRh102GyKJ6xWRNiRmLpRGWNSIfQh%0A0wUz%2Bvm%2BjUvFbm7qG4stzI8NJ75lBbE0So1UpRTLUExLM7oceYTdnznXWa%2F734Iy%0AH7xNFuwVDL6WctixdqtrNnmivXnmIP83427ehi%2B9ta%2Bhgwy4PFWW%2FBth7F%2BieFs%2B%0AQGP6i02nnAPcuYiEieCNTd7R21AEKh1%2FpcColABGXZQguKFZND2FmVjjcdEAOABW%0A%2B2GxZDkusIprH9TXUw%3D%3D%0A-----END%20CERTIFICATE-----%0A\";Subject=\"O=ML Client Org,CN=ML Client Curl\";URI=,By=spiffe://cluster.local/ns/test/sa/httpbin;Hash=5caf3404f9404e4dd4314f2a184d17d89082818f2550f6a24e7de2ba7f400c52;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
        "X-Forwarded-For": "10.42.0.1",
        "X-Forwarded-Host": "httpbin.mtls.local.kyma.dev",
        "X-Forwarded-Proto": "https",
        "X-Request-Id": "6d90f040-14d5-4eee-87b7-db92aa004eee"
      }
    }
    ```
    2. To test the connection in Chrome...