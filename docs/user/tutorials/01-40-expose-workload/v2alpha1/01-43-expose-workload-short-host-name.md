# Expose a Workload with Short Host Name

This tutorial demonstrates how to expose an unsecured Service instance using a short host name instead of the full domain name. Using a short host makes it simpler to apply APIRules because the domain name is automatically retrieved from the referenced Gateway, and you don’t have to manually set it in each APIRule. This might be particularly useful when reconfiguring resources in a new cluster, as it reduces the chance of errors and streamlines the process.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have a deployed workload.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: 

* To test access to the exposed service, you must install [curl](https://curl.se).

## Steps

### Expose Your Workload
To expose your workload using a short host, replace placeholders and create the following APIRule CR. You can adjust the configuration, if needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: {APIRULE_NAME}
  namespace: {APIRULE_NAMESPACE}
spec:
  hosts:
    - {SUBDOMAIN}
  service:
    name: {SERVICE_NAME}
    namespace: {SERVICE_NAMESPACE}
    port: {SERVICE_PORT}
  gateway: {NAMESPACE/GATEWAY}
  rules:
    - path: /*
      methods: ["GET"]
      noAuth: true
    - path: /post
      methods: ["POST"]
      noAuth: true
EOF
```

Option | Description
---------|----------
 APIRULE_NAME | Choose a name for the APIRule CR.
 APIRULE_NAMESPACE | Choose the namespace for the APIRule CR.
 SUBDOMAIN | Choose the subdomain name. When using a short host you are not required to specify the full domain.
 SERVICE_NAME | The name of the service to be exposed.
 SERVICE_NAMESPACE | The namespace of the service to be exposed.
 SERVICE_PORT | The port of the service to be exposed.
 NAMESPACE/GATEWAY | The namespace and name of the Istio Gateway to be used.
    
The host domain name is obtained from the referenced Gateway.

### Access Your Workload

- Replace the placeholder and send a `GET` request to the service.

  ```bash
  curl -ik -X GET https://{FULL_DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Replace the placeholder and send a `POST` request to the service.

  ```bash
  curl -ik -X POST https://{FULL_DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.