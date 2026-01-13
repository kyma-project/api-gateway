# Expose a Workload with noAuth on k3d

This tutorial is a continuation of the [Kyma Quick Install guide](https://kyma-project.io/02-get-started/01-quick-install.html). It shows how to expose an unsecured instance of the HTTPBin Service on your k3d cluster and call its endpoints.

## Prerequisites

- You have Istio and API Gateway modules in your [k3d](https://k3d.io/stable/) cluster. See [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html).
- You have installed [curl](https://curl.se).

## Context

After completing the Quick Install guide, you have a k3d cluster with the default Kyma Gateway configured under the `*.local.kyma.dev` wildcard domain. The domain is registered in public DNS and points to the local host `127.0.0.1`. This tutorial shows how to create a sample HTTPBin workload and expose it using an APIRule custom resource (CR) with the `noAuth` access strategy.

The `noAuth` access strategy allows public access to your workload without any authentication or authorization checks. This is useful for:
- Development and testing environments
- Public APIs that don't require authentication
- Services that implement their own authentication logic

> [!WARNING]
> Exposing a workload without authentication is a potential security vulnerability. In production environments, always secure your workloads with proper authentication such as [JWT](./01-40-expose-workload-jwt.md).

To expose a workload without authentication, create an APIRule with `noAuth: true` configured for each path you want to expose publicly.

## Procedure

>[!NOTE]
> To expose a workload using APIRule in version `v2`, the workload must be part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/external-content/istio/docs/user/tutorials/01-40-enable-sidecar-injection.html#enable-istio-sidecar-proxy-injection).

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE="test"
    kubectl create ns "${NAMESPACE}"
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite
    ```

2. Get the default domain of your Kyma cluster.

    ```bash
    PARENT_DOMAIN=local.kyma.dev
    WORKLOAD_DOMAIN="httpbin.${PARENT_DOMAIN}"
    GATEWAY="kyma-system/kyma-gateway"
    echo "Parent domain: ${PARENT_DOMAIN}"
    echo "Workload domain: ${WORKLOAD_DOMAIN}"
    echo "Gateway namespace and name: ${GATEWAY}"
    ```

3. Deploy a sample instance of the HTTPBin Service.

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
        - ${WORKLOAD_DOMAIN}
      service:
        name: httpbin
        namespace: ${NAMESPACE}
        port: 8000
      gateway: ${GATEWAY}
      rules:
        - path: /post
          methods: ["POST"]
          noAuth: true
        - path: /{**}
          methods: ["GET"]
          noAuth: true
    EOF
    ```

    Check if the APIRule's status is ready:

    ```bash
    kubectl get apirules httpbin -n "${NAMESPACE}" 
    ```

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