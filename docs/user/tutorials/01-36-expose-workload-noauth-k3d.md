# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose with [JWT](./01-40-expose-workload-jwt.md).

## Prerequisites

- You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud) for SAP BTP, Kyma runtime or [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html) for open-source Kyma.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and select **Create**.
2. Provide the name of the APIRule CR.
3. Add the name and port of the service you want to expose.
4. Add a Gateway.
5. Add a rule with the following configuration:
    - **Path**: `/post`
    - **Handler**: `No Auth`
    - **Methods**: `POST`
6. Add one more rule with the following configuration:
    - **Path**: `/{**}`
    - **Handler**: `No Auth`
    - **Methods**: `GET`
7. Choose **Create**.

#### **kubectl**

To expose your workload, create an APIRule CR. You can adjust the configuration, if needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ${APIRULE_NAME}
  namespace: ${APIRULE_NAMESPACE}
spec:
  hosts:
    - ${SUBDOMAIN}.${DOMAIN_NAME}
  service:
    name: ${SERVICE_NAME}
    namespace: ${SERVICE_NAMESPACE}
    port: ${SERVICE_PORT}
  gateway: ${GATEWAY_NAMESPACE}/${GATEWAY_NAME}
  rules:
    - path: /post
      methods: ["POST"]
      noAuth: true
    - path: /{**}
      methods: ["GET"]
      noAuth: true
EOF
```

<!-- tabs:end -->


### Access Your Workload

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET https://${SUBDOMAIN}.${DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST https://${SUBDOMAIN}.${DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.



# Quick Start Guide for Open-Source Kyma

Follow the steps to get started with the open-source API Gateway module.

## Prerequisites

You have created a k3d cluster and added the Istio and API Gateway modules. See [Quick Install](https://kyma-project.io/#/02-get-started/01-quick-install).

## Context
This quick start guide shows how to create a sample HTTPBin workload and expose it to the internet using the APIRule custom resource (CR). For this purpose, the guide uses a wildcard public domain `*.local.kyma.dev`. The domain is registered in public DNS and points to the local host `127.0.0.1`.

## Procedure

## Create a Workload

<!-- tabs:start -->
#### **Kyma Dashboard**

1. In Kyma dashboard, go to **Namespaces** and choose **Create**.
1. Provide the name `api-gateway-tutorial` and switch the toggle to enable Istio sidecar proxy injection.
2. Choose **Create**.
3. In the created namespace, go to **Workloads > Deployments** and choose **Create**.
1. Select the HTTPBin template and choose **Create**.
3. Go to **Configuration > Service Accounts** and choose **Create**. 
4. Enter `httpbin` as your Service Account's name and choose **Create**.
6. Go to **Discovery and Network > Services** and choose **Create**. 
7. Provide the following configuration details:
    - **Name**: `httpbin`
    - Add the following labels:
      - **service**: `httpbin`
      - **app**:`httpbin`
    - Add the following selector:
      - **app**: `httpbin`
    - Add a port with the following values:
      - **Name**: `http`
      - **Protocol**: `TCP`
      - **Port**: `8000`
      - **Target Port**: `80`
8. Choose **Create**.

#### **kubectl**

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE=api-gateway-tutorial
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Deploy a sample instance of the HTTPBin Service.

    ```bash
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: httpbin
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: httpbin
      labels:
        app: httpbin
        service: httpbin
    spec:
      ports:
      - name: http
        port: 8000
        targetPort: 80
      selector:
        app: httpbin
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: httpbin
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: httpbin
          version: v1
      template:
        metadata:
          labels:
            app: httpbin
            version: v1
        spec:
          serviceAccountName: httpbin
          containers:
          - image: docker.io/kennethreitz/httpbin
            imagePullPolicy: IfNotPresent
            name: httpbin
            ports:
            - containerPort: 80
    EOF
    ```

    To verify if an instance of the HTTPBin Service is successfully created, run:

    ```bash
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```

    If successful, you get a result similar to this one:

    ```shell
    NAME                 READY    STATUS     RESTARTS    AGE
    httpbin-{SUFFIX}     2/2      Running    0           96s
    ```

<!-- tabs:end -->

## Expose a Workload

<!-- tabs:start -->
#### **Kyma Dashboard**

1. In the `api-gateway-tutorial` namespace, go to **Discovery and Network > API Rules**.
2. Choose **Create**.
3. Provide the following details:
     - **Name**: `httpbin`
     - Add a service with the following values:
       - **Service Name**: `httpbin`
       - **Port**: `8000`
     - Add a gateway with the following values:
       - **Namespace**: `kyma-system`
       - **Name**: `kyma-gateway`
       - **Host**: `httpbin.local.kyma.dev`
     - Add one Rule with the following configuration:
       - **Path**: `/post`
       - **Handler**: `No Auth`
       - **Methods**: `POST`
     - Create one more Rule with the following configuration:
       - **Path**: `/{**}`
       - **Handler**: `No Auth`
       - **Methods**: `GET`
4.  Choose **Create**.

#### **kubectl**

To expose the HTTPBin Service, create the following APIRule CR. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin
  namespace: api-gateway-tutorial
spec:
  hosts:
    - httpbin.local.kyma.dev
  service:
    name: httpbin
    namespace: api-gateway-tutorial
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /post
      methods: ["POST"]
      noAuth: true
    - path: /{**}
      methods: ["GET"]
      noAuth: true
EOF
```

<!-- tabs:end -->

## Access a Workload

To access the HTTPBin Service, use [curl](https://curl.se).

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET https://httpbin.local.kyma.dev:30443/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST https://httpbin.local.kyma.dev:30443/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.

<!-- tabs:end -->