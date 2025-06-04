# Expose and Secure a Workload with a Certificate

This tutorial shows how to expose and secure a workload with mutual authentication using TLS Gateway.

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Set Up Your Custom Domain](../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.
  
* [Set up a mutual TLS Gateway](../01-30-set-up-mtls-gateway.md) and export the bundle certificates.
* Optionally, you can [create your own self-signed Client Root CA and certificate](../01-60-security/01-61-mtls-selfsign-client-certicate.md).

## Authorize a Client with a Certificate

> [!NOTE]
>  Create an AuthorizationPolicy to verify that the name specified in it matches the client's common name in the certificate.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to the namespace in which you want to create an APIRule CR.
   
   > [!NOTE] The namespace that you use for creating an APIRule must have Istio sidecar injection enabled. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

2. Go to **Discovery and Network > APIRule** and select **Create**. 
3. Provide the following configuration details:
    - Add a name for your APIRule CR.
    - Add the name and namespeca of the Gateway you want to use.
    - Specify the host.
4. Add a Rule with the following configuration:
    - **Path**:`/*`
    - **Methods**:`GET`
    - **Access Strategy**: `No Auth`
    - In **Requests > Headers**, add the following key-value pairs: 
      - **X-CLIENT-SSL-CN**: `%DOWNSTREAM_PEER_SUBJECT%`
      - **X-CLIENT-SSL-SAN**: `%DOWNSTREAM_PEER_URI_SAN%`
      - **X-CLIENT-SSL-ISSUER**: `%DOWNSTREAM_PEER_ISSUER%`
    - Add the name and port of the Service you want to expose.
5. Choose **Create**.

#### **kubectl**

1. Export the following values as environment variables:

  ```bash
  export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
  export CLIENT_CERT_CN={COMMON_NAME}
  export CLIENT_CERT_ORG={ORGANIZATION}
  export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
  export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
  ```

2. Create an APIRUle CR that adds the **X-CLIENT-SSL** headers to incoming requests.
   
   > [!NOTE] The namespace that you use for creating an APIRule must have Istio sidecar injection enabled. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

    ```bash
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      labels:
      name: {APIRULE_NAME}
      namespace: {APIRULE_NAMESPACE}
    spec:
      gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
      hosts:
        - {SUBDOMAIN}.{DOMAIN}
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
              test: 'true'
      service:
        name: httpbin
        port: 8000
    ```

<!-- tabs:end -->

## Access the Secured Resources

Call the secured endpoints of the HTTPBin Service.

Send a `GET` request to the HTTPBin Service with the client certificates that you used to create mTLS Gateway:

```bash
curl --key ${CLIENT_CERT_KEY_FILE} \
      --cert ${CLIENT_CERT_CRT_FILE} \
      --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
      -ik -X GET https://httpbin-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
```

If successful, the call returns the `200 OK` response code. If you call the Service without the proper certificates or with invalid ones, you get the error `403 Forbidden`.