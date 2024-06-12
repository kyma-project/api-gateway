# Expose Multiple Workloads on the Same Host

This tutorial shows how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING] 
>  Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, remember to secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Deploy two instances of a sample HTTPBin Service](../../01-00-create-workload.md) in one namespace. 
* [Set up your custom domain](../../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 

## Define Multiple Services on Different Paths

Follow the instructions to expose the instances of the HTTPBin Service on different paths at the `spec.rules` level without a root Service defined.

#### **kubectl**
1. Export the names of two deployed HTTPBin Services as environment variables:
  
  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  ```

2. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <!-- tabs:start -->
  #### **Custom Domain**
      
  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
  #### **Kyma Domain**

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=kyma-system/kyma-gateway
  ```
  <!-- tabs:end --> 

3. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      hosts: 
        - multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
      - path: /headers
        methods: ["GET"]
        noAuth: true
        service:
          name: $FIRST_SERVICE
          port: 8000
      - path: /get
        methods: ["GET"]
        noAuth: true
        service:
          name: $SECOND_SERVICE
          port: 8000
    EOF
    ```

## Define a Service at the Root Level

You can also define a Service at the root level. Such a definition is applied to all the paths specified at `spec.rules` that do not have their own Services defined.
 
> [!NOTE] 
>Services defined at the `spec.rules` level have precedence over Service definition at the `spec.service` level.

#### **kubectl**

1. Export the names of the two deployed HTTPBin Services as environment variables:
  
  ```bash
  export FIRST_SERVICE={SERVICE_NAME}
  export SECOND_SERVICE={SERVICE_NAME}
  ```

2. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <!-- tabs:start -->
  #### **Custom Domain**
      
  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
  #### **Kyma Domain**

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=kyma-system/kyma-gateway
  ```
  <!-- tabs:end --> 


3. To expose the instances of the HTTPBin Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      hosts: 
        - multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      service:
        name: $FIRST_SERVICE
        port: 8000
      rules:
        - path: /headers
          methods: ["GET"]
          noAuth: true
        - path: /get
          methods: ["GET"]
          noAuth: true
          service:
            name: $SECOND_SERVICE
            port: 8000
    EOF
    ```

## Access Your Workloads
To access your HTTPBin Services, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

- Enter the URL `https://multiple-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with your domain name. To call the endpoint, send a `GET` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

- Enter the URL `https://multiple-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/get` and replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with your domain name. To call the endpoint, send a `GET` request to the HTTPBin Service. If successful, the call returns the `200 OK` response code.

#### **curl**

To call the endpoints, send `GET` requests to the HTTPBin Services:

  ```bash 
  curl -ik -X GET https://multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

  curl -ik -X GET https://multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get 
  ```
If successful, the calls return the `200 OK` response code.

<!-- tabs:end -->
