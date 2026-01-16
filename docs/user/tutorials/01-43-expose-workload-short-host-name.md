# Expose a Workload with Short Host Name

Learn how to expose an unsecured Service instance using a short host name instead of the full domain name. 

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose with [JWT](./01-40-expose-workload-jwt.md).

## Prerequisites

* You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud) for SAP BTP, Kyma runtime or [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html) for open-source Kyma.
* You have a deployed workload.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* To set up a custom Gateway, see [Configure a TLS Gateway in SAP BTP, Kyma Runtime](./01-20-set-up-tls-gateway.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment. For more information, see [Getting Started with Istio Gateways](../00-05-domains-and-gateways.md).

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

## Context
Using a short host makes it simpler to apply APIRules because the domain name is automatically retrieved from the referenced Gateway, and you don’t have to manually set it in each APIRule. This might be particularly useful when reconfiguring resources in a new cluster, as it reduces the chance of errors and streamlines the process. The referenced Gateway must provide the same single host for all [Server](https://istio.io/latest/docs/reference/config/networking/gateway/#Server) definitions, and it must be prefixed with `*.`.

## Steps

### Expose Your Workload
To expose your workload using a short host, replace placeholders and create the following APIRule CR. You can adjust the configuration, if needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: ${APIRULE_NAME}
  namespace: ${APIRULE_NAMESPACE}
spec:
  hosts:
    - ${SUBDOMAIN}
  service:
    name: ${SERVICE_NAME}
    namespace: ${SERVICE_NAMESPACE}
    port: ${SERVICE_PORT}
  gateway: ${GATEWAY_NAMESPACE}/${GATEWAY_NAME}
  rules:
    - path: /post
      methods: ["POST"]
      noAuth: true
    - path: /*
      methods: ["GET"]
      noAuth: true
EOF
```

### Access Your Workload

- Replace the placeholder and send a `GET` request to the service.

  ```bash
  curl -ik -X GET https://${SUBDOMAIN}.${DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Replace the placeholder and send a `POST` request to the service.

  ```bash
  curl -ik -X POST https://${SUBDOMAIN}.${DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.