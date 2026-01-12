# Expose a Workload with noAuth on k3d

Learn how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose with [JWT](./01-40-expose-workload-jwt.md).

## Prerequisites

- You have Istio and API Gateway modules in your [k3d](https://k3d.io/stable/) cluster. See [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html).
- You have installed [curl](https://curl.se).

## Context
This guide shows how to create a sample HTTPBin workload and expose it using the APIRule custom resource (CR). For this purpose, the guide uses a wildcard public domain `*.local.kyma.dev`. The domain is registered in public DNS and points to the local host `127.0.0.1`.

## Procedure

Follow this example to create an APIRule that exposes a sample HTTPBin Deployment.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Namespaces** and create a namespace with enabled Istio sidecar proxy injection.
2. Select **+ Upload YAML**, paste the following cofiguration and upload it.
    
    ```yaml
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
    ```
2. Go to **Discovery and Network > API Rules** and select **Create**.
2. Provide the name of the APIRule CR.
3. In the **Service** section, add the name `httpbin` and port `8000`.
4. Use the default Gateway `kyma-system/kyma-gateway`.
5. Add the host `httpbin.local.kyma.dev`.
6. Add a rule with the following configuration:
    - **Path**: `/post`
    - **Handler**: `No Auth`
    - **Methods**: `POST`
7. Add one more rule with the following configuration:
    - **Path**: `/{**}`
    - **Handler**: `No Auth`
    - **Methods**: `GET`
8. Choose **Create**.

#### **kubectl**

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE="api-gateway-tutorial"
    kubectl create ns "${NAMESPACE}"
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite
    ```

2. Deploy a sample instance of the HTTPBin Service.

    ```bash
    cat <<EOF | kubectl -n "${NAMESPACE}" apply -f -
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
    kubectl get pods -l app=httpbin -n "${NAMESPACE}"
    ```

    If successful, you get a result similar to this one:

    ```shell
    NAME                 READY    STATUS     RESTARTS    AGE
    httpbin-{SUFFIX}     2/2      Running    0           96s
    ```

3. To expose the HTTPBin Service, create the following APIRule CR, which uses the default Kyma Gateway `kyma-system/kyma-gateway`. Run:

    ```bash
    cat <<EOF | kubectl apply -n "${NAMESPACE}" -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin
    spec:
      hosts:
        - httpbin.local.kyma.dev
      service:
        name: httpbin
        namespace: ${NAMESPACE}
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

## Result

To access the HTTPBin Service, use curl.

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET "https://${WORKLOAD_DOMAIN}:30443/ip"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST "https://${WORKLOAD_DOMAIN}:30443/post" -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.

<!-- tabs:end -->