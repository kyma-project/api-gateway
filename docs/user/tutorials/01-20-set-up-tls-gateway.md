# Set Up a TLS Gateway

This tutorial shows how to set up a TLS Gateway in simple mode.

## Prerequisites

* Deploy [a sample HTTPBin Service](./01-00-create-workload.md).
* Set up [your custom domain](./01-10-setup-custom-domain-for-workload.md).
   

## Set Up a TLS Gateway in Simple Mode

<!-- tabs:start -->
#### **Kyma dashboard**

1. In the **Istio** section, select **Gateways**, and then **Create**. 
2. Switch to the `Advanced` tab and provide the following configuration details:
  - **Name**: `httpbin-gateway`
  - In the `Selectors` section, add the following selector: 
    - **istio**: `ingress-gateway`
  - In the `Servers` section, select **Add**. Then, use these values:
    - **Port Number**: `443`
    - **Name**: `https`
    - **Protocol**: `HTTPS`
    - **TLS Mode**: `SIMPLE`
    - In the **Credential Name** fied, select the name of the Secret thet contains the credentials you want to use.
  - Use `httpbin.{CUSTOM_DOMAIN}` as a host. Replace `{CUSTOM_DOMAIN}` with the name of your custom domain. 

3. To confirm the Gateway creation, select **Create**.


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