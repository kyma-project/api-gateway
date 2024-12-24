# Expose and Secure a Workload with a Certificate

This tutorial shows how to expose and secure a workload with mutual authentication using TLS Gateway.

## Prerequisites

* [Deploy a sample HTTPBin Service](../01-00-create-workload.md).
* [Set Up Your Custom Domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  > [!NOTE]
  > Bacuse the default Kyma domain is a widlcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what is the default domain of your Kyma cluster, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.
* [Set up a mutual TLS Gateway](../01-30-set-up-mtls-gateway.md) and export the bundle certificates.
* Optionally, you can [create your own self-signed Client Root CA and certificate](../01-60-security/01-61-mtls-selfsign-client-certicate.md).

## Authorize a Client with a Certificate

> [!NOTE]
>  Create an AuthorizationPolicy to verify that the name specified in it matches the client's common name in the certificate.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Istio > Virtual Services** and select **Create**. Provide the following configuration details:
    - **Name**: `httpbin-vs`
    - In the `HTTP` section, select **Add**. Add a route with the destination port `8000` and the host HTTPBin. Then, go to **HTTP > Headers > Request > Set** and add these headers:
      - **X-CLIENT-SSL-CN**: `%DOWNSTREAM_PEER_SUBJECT%`
      - **X-CLIENT-SSL-SAN**: `%DOWNSTREAM_PEER_URI_SAN%`
      - **X-CLIENT-SSL-ISSUER**: `%DOWNSTREAM_PEER_ISSUER%`
    - In the `Host` section, add `httpbin.{YOUR_DOMAIN}`. Replace `{YOUR_DOMAIN}` with the name of your domain.
    - In the `Gateways` section, add the name of your mTLS Gateway.
3. Go to **Istio > Authorization Policies** and select **Create**. Provide the following configuration details:
    - **Name**: `test-authz-policy`
    - **Action**: `ALLOW`
    - Add a Rule. Go to **Rule > To > Operation > Hosts** and add the host `httpbin-vs.{DOMAIN_TO_EXPOSE_WORKLOADS}`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your domain. Then, go to **Rule > When** and add:
      - **key**: `request.headers[X-Client-Ssl-Cn]`
      - **values**: `["O={CLIENT_CERT_ORG},CN={CLIENT_CERT_CN}"]`
    Replace `{CLIENT_CERT_ORG}` with the name of your organization and `{CLIENT_CERT_CN}` with the common name.

#### **kubectl**

1. Export the following values as environment variables:

  ```bash
  export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
  export CLIENT_CERT_CN={COMMON_NAME}
  export CLIENT_CERT_ORG={ORGANIZATION}
  export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
  export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
  ```

2. Create VirtualService that adds the **X-CLIENT-SSL** headers to incoming requests:

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

3. Create AuthorizationPolicy that verifies if the request contains a client certificate:

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