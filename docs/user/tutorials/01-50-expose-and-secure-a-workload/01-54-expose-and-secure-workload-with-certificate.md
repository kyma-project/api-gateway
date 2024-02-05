# Expose and Secure a Workload with a Certificate

This tutorial shows how to expose and secure a workload with mutual authentication using TLS Gateway.

## Prerequisites

* Deploy [a sample HTTPBin Service](../01-00-create-workload.md).
* Set up [your custom domain](../01-10-setup-custom-domain-for-workload.md).
* Set up [a mutual TLS Gateway](../01-20-set-up-tls-gateway.md) and export the bundle certificates.
* To learn how to create your own self-signed Client Root CA and certificate, see [this tutorial](../01-60-security/01-61-mtls-selfsign-client-certicate.md). This step is optional.
* Export the following values as environment variables:

  ```bash
  export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
  export CLIENT_CERT_CN={COMMON_NAME}
  export CLIENT_CERT_ORG={ORGANIZATION}
  export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
  export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
  ```

## Authorize a Client with a Certificate

The following instructions describe how to secure an mTLS Service. 
> [!NOTE]
>  Create AuthorizationPolicy to check if the client's common name in the certificate matches.

1. Create VirtualService that adds the X-CLIENT-SSL headers to incoming requests:

    ```bash
    cat <<EOF | kubectl apply -f - 
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: httpbin-vs
      namespace: ${NAMESPACE}
    spec:
      hosts:
      - "httpbin-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"
      gateways:
      - ${MTLS_GATEWAY_NAME}
      http:
      - route:
        - destination:
            port:
              number: 8000
            host: httpbin
          headers:
            request:
              set:
                X-CLIENT-SSL-CN: "%DOWNSTREAM_PEER_SUBJECT%"
                X-CLIENT-SSL-SAN: "%DOWNSTREAM_PEER_URI_SAN%"
                X-CLIENT-SSL-ISSUER: "%DOWNSTREAM_PEER_ISSUER%"
    EOF
    ```

2. Create AuthorizationPolicy that verifies if the request contains a client certificate:
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: security.istio.io/v1beta1
    kind: AuthorizationPolicy
    metadata:
      name: test-authz-policy
      namespace: ${NAMESPACE}
    spec:
      action: ALLOW
      rules:
      - to:
        - operation:
            hosts: ["httpbin-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"]
        when:
        - key: request.headers[X-Client-Ssl-Cn]
          values: ["O=${CLIENT_CERT_ORG},CN=${CLIENT_CERT_CN}"]
    EOF
    ```
  
3. Call the secured endpoints of the HTTPBin Service.

    Send a `GET` request to the HTTPBin Service with the client certificates that you used to create mTLS Gateway:

    ```shell
    curl --key ${CLIENT_CERT_KEY_FILE} \
          --cert ${CLIENT_CERT_CRT_FILE} \
          --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
          -ik -X GET https://httpbin-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
    ```

    If successful, the call returns the code `200 OK` response. If you call the Service without the proper certificates or with invalid ones, you get the code `403` response.