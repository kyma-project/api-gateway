# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have a deployed workload.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: To learn what is the default domain of your Kyma cluster, run  `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

## Steps

### Expose a Workload

To simply expose your workload, create an APIRule CR. You can adjust the configuration, if needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: {APIRULE_NAME}
  namespace: {APIRULE_NAMESPACE}
spec:
  hosts:
    - {SUBDOMAIN}.{DOMAIN_NAME}
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
SUBDOMAIN.DOMAIN | The name of your subdomain and domain.
SERVICE_NAME | The name of the service to be exposed.
SERVICE_NAMESPACE | The namespace of the service to be exposed.
SERVICE_PORT | The port of the service to be exposed.
NAMESPACE/GATEWAY | The namespace and name of the Istio Gateway to be used.


### Access Your Workload

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.