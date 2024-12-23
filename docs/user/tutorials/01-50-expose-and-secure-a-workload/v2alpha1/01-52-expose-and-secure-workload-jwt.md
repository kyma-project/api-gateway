# Expose and Secure a Workload with JWT

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/), [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* You have a deployed workload.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.

  >**NOTE**: As Kyma domain is a widlcard domain, which uses a simple TLS gateway, it recommended that you set up your custom domain istead for use in a production environment.
  >**TIP**: To learn what is the default domain of your Kyma cluster, run  `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

* [Obtain a JSON Web Token (JWT)](../01-51-get-jwt.md).


## Steps

### Expose and Secure Your Workload

#### **kubectl**

To expose and secure the Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      hosts:
        - httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
      service:
        name: httpbin
        port: 8000
      gateway: $GATEWAY
      rules:
        - jwt:
            authentications:
              -  issuer: $ISSUER
                 jwksUri: $JWKS_URI
          methods:
            - GET
          path: /.*
    EOF
    ```

### Access the Secured Resources

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
    ```
    You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
    You get the `200 OK` response code.