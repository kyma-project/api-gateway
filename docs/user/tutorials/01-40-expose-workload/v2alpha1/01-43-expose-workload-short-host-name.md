# Expose a Workload with Short Host Name

Learn how to expose an unsecured Service instance using a short host name instead of the full domain name. 

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have a deployed workload.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: To learn what is the default domain of your Kyma cluster, run  `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

* To test access to the exposed service, you must install [curl](https://curl.se).

## Context
Using a short host makes it simpler to apply APIRules because the domain name is automatically retrieved from the referenced Gateway, and you don’t have to manually set it in each APIRule. This might be particularly useful when reconfiguring resources in a new cluster, as it reduces the chance of errors and streamlines the process.

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

### Access Your Workload

- Replace the placeholder and send a `GET` request to the service.

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Replace the placeholder and send a `POST` request to the service.

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.