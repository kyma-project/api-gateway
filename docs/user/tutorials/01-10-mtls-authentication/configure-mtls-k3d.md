# Configure mTLS Authentication Locally on k3d
Learn how to set up mutual TLS (mTLS) authentication in a local Kyma environment using k3d.

## Context
mTLS (mutual TLS) provides two‑way authentication: the client verifies the server's identity and the server verifies the client's identity. To enforce this authentication, the mTLS Gateway requires the following values: 
- the server private key
- the server certificate chain (server certificate plus any intermediate CAs)
- the client root CA used to validate presented client certificates. 
Each client connecting through the mTLS Gateway must have a valid client certificate and key and trust the server's root CA.

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
    k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer --k3s-arg "--disable=traefik@server:*"
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

4. Export the following domain names as environment variables, you might want to adapt them to your use case:

    ```bash
    PARENT_DOMAIN="local.kyma.dev"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    echo "Parent Domain: ${PARENT_DOMAIN}"
    echo "Subdomain: ${SUBDOMAIN}"
    echo "Gateway Domain: ${GATEWAY_DOMAIN}"
    echo "Workload Domain: ${WORKLOAD_DOMAIN}"
    ```

   | Placeholder         | Example domain name           | Description                                                                                                                                                                                                                                                                                                                                                            |
   |---------------------|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
   | **PARENT_DOMAIN**   | `local.kyma.dev`              | The main wildcard public domain for your local Kyma installation. The domain is registered in public DNS and points to the local host `127.0.0.1`. By default, this domain is used by the API Gateway module to configure the default TLS Gateway. To avoid conflicts and enable custom gateways, you must use a subdomain of this parent domain for your own Gateway. |
   | **SUBDOMAIN**       | `mtls.local.kyma.dev`         | A dedicated subdomain created under the parent domain, specifically for the mTLS Gateway. This isolates mTLS traffic and allows you to manage certificates and routing separately from the default Gateway.                                                                                                                                                            |
   | **GATEWAY_DOMAIN**  | `*.mtls.local.kyma.dev`       | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.local.kyma.dev`, `test.httpbin.mtls.local.kyma.dev`) without creating separate Gateway rules for each one.                                                                |
   | **WORKLOAD_DOMAIN** | `httpbin.mtls.local.kyma.dev` | The specific domain assigned to your sample workload (HTTPBin service) in this tutorial.                                                                                                                                                                                                                                                                               |

5. Create the server's root CA.

    ```bash
    SERVER_ROOT_CA_KEY_FILE="server_root_ca_cn.key"
    SERVER_ROOT_CA_CRT_FILE="server_root_ca_cn.crt"
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=Example Server Root CA ORG/CN=Example Client Root CA CN" -keyout "${SERVER_ROOT_CA_KEY_FILE}" -out "${SERVER_ROOT_CA_CRT_FILE}"
    ```

6. Create the server's certificate.
    
    ```bash
    SERVER_CERT_CRT_FILE=""${GATEWAY_DOMAIN}".crt"
    SERVER_CERT_CSR_FILE=""${GATEWAY_DOMAIN}".csr"
    SERVER_CERT_KEY_FILE=""${GATEWAY_DOMAIN}".key"
    openssl req -out "${SERVER_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${SERVER_CERT_KEY_FILE}" -subj "/CN=${GATEWAY_DOMAIN}/O=Example Server Cert Org"
    ```

7. Sign the server's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${SERVER_ROOT_CA_CRT_FILE}" -CAkey "${SERVER_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${SERVER_CERT_CSR_FILE}" -out "${SERVER_CERT_CRT_FILE}"
    ```

8. Create the server's certificate chain consisting of the server's certificate and the server's root CA.
   
    ```bash
    SERVER_CERT_CHAIN_FILE="${SERVER_CERT_CN}-chain.pem"
    cat "${SERVER_CERT_CRT_FILE}" "${SERVER_ROOT_CA_CRT_FILE}" > "${SERVER_CERT_CHAIN_FILE}"
    ```
9. Create a Secret for the mTLS Gateway with the server's key and certificate.
    
    ```bash
    kubectl create secret tls -n istio-system kyma-mtls --key="${SERVER_CERT_KEY_FILE}" --cert="${SERVER_CERT_CHAIN_FILE}"
    ```

10. Create the client's root CA.

    ```bash
    CLIENT_ROOT_CA_KEY_FILE="client_root_ca_cn.key"
    CLIENT_ROOT_CA_CRT_FILE="client_root_ca_cn.crt"
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=Example Client Root CA ORG/CN=Example Client Root CA CN" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
    ```

11.  Create the client's certificate.

      ```bash
      CLIENT_CERT_CRT_FILE="client_cert_cn.crt"
      CLIENT_CERT_CSR_FILE="client_cert_cn.csr"
      CLIENT_CERT_KEY_FILE="client_cert_cn.key"
      openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=Example Client Cert CN/O=Example Client Cert Org"
      ``` 

12.  Sign the client's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"
    ```

13.  Create a Secret for the mTLS Gateway containing the client's CA certificate. 
    The Secret must follow Istio convention. See [Key Formats](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#key-formats).
    
    ```bash
    kubectl create secret generic -n istio-system "kyma-mtls-cacert" --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"
    ```

14.  Create the mTLS Gateway.
    
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

15.  Create a sample HTTPBin Deployment.

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

16. To expose the sample HTTPBin Deployment, create an APIRule custom resource.
    The APIRule appends the headers `X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'`, `X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'`, and `X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'` to the request. 
    These headers provide the upstream (your workload) with the downstream (authenticated client's) identity.
    This is optional configuration is commonly used in mTLS use cases.
    For more information about these values, see [Envoy Access logging](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#access-logging)

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

17.  To test the mTLS connection, run the following curl command:

     1. Run the following curl command:
     
     ```bash
     curl --fail --verbose \
       --key "${CLIENT_CERT_KEY_FILE}" \
       --cert "${CLIENT_CERT_CRT_FILE}" \
       --cacert "${SERVER_ROOT_CA_CRT_FILE}" \
       "https://${WORKLOAD_DOMAIN}/headers?show_env==true"
     ```
     If successful, you get code `200` in response. The configured headers are also populated. See the following example:
        ```bash
        {
          "headers": {
            ...
            "X-Client-Ssl-Cn": "O=Example Client Cert Org,CN=Example Client Cert CN",
            "X-Client-Ssl-Issuer": "CN=Example Client Root CA CN,O=Example Client Root CA ORG",
            ...
          }
        }
        ```
