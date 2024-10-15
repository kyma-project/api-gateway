# Set Up a TLS Gateway in Simple Mode

This tutorial shows how to set up a TLS Gateway in simple mode.

## Prerequisites

* [Deploy a sample HTTPBin Service](./01-00-create-workload.md).
* [Set up your custom domain](./01-10-setup-custom-domain-for-workload.md).

## Steps


<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Istio > Gateways** and select **Create**.
2. Provide the following configuration details:
    - **Name**: `httpbin-gateway`
    - In the `Servers` section, select **Add**. Then, use these values:
      - **Port Number**: `443`
      - **Name**: `https`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `SIMPLE`
      - **Credential Name** is the name of the Secret that contains the credentials.
    - Use `httpbin.{CUSTOM_DOMAIN}` as **Host**. Replace `{CUSTOM_DOMAIN}` with the name of your custom domain.

3. Select **Create**.

#### **kubectl**

1. Export the following values as environment variables:

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```

2. To create a TLS Gateway in simple mode, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: httpbin-gateway
      namespace: $NAMESPACE
    spec:
      selector:
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: https
            protocol: HTTPS
          tls:
            mode: SIMPLE
            credentialName: $TLS_SECRET
          hosts:
            - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
    EOF
    ```

<!-- tabs:end -->