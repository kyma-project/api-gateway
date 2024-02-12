# Set Up a TLS Gateway

This tutorial shows how to set up a TLS Gateway in simple modes.

## Prerequisites

* Deploy [a sample HTTPBin Service](./01-00-create-workload.md).
* Set up [your custom domain](./01-10-setup-custom-domain-for-workload.md) and export the following values as environment variables:

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
   

## Set Up a TLS Gateway in Simple Mode

To create a TLS Gateway in simple mode, run:

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
      istio: ingressgateway # Use Istio Ingress Gateway as default
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