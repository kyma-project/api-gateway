# Expose a Workload with noAuth in SAP BTP, Kyma Runtime

Learn how to expose an unsecured instance of the HTTPBin Service using the `noAuth` access strategy and call its endpoints.

## Context

The `noAuth` access strategy allows public access to your workload without any authentication or authorization checks. This is useful for:
- Development and testing environments
- Public APIs that don't require authentication
- Services that implement their own authentication logic

> [!WARNING]
> Exposing a workload without authentication is a potential security vulnerability. In production environments, always secure your workloads with proper authentication such as [JWT](./01-40-expose-workload-jwt.md).

To expose a workload without authentication, create an APIRule with `noAuth: true` configured for each path you want to expose publicly.

## Prerequisites

- You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud).

## Procedure

>[!NOTE]
> To expose a workload using APIRule in version `v2`, the workload must be part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/external-content/istio/docs/user/tutorials/01-40-enable-sidecar-injection.html#enable-istio-sidecar-proxy-injection).

<!-- tabs:start -->
#### **Kyma Dashboard**
1. In a namespace of your choice go to **Discovery and Network > API Rules** and choose **Create**. 
2. Provide all the required configuration details.
3. Add a rule with the following configuration.
    - **Access Strategy**: `noAuth`
    - Add allowed methods and the request path.
4. Choose **Create**.  

#### **kubectl**

Replace the placeholders and apply the following configuration. Adjust the rules sections as needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ${APIRULE_NAME}
  namespace: ${NAMESPACE}
spec:
  hosts:
    - ${SUBDOMAIN}.${PARENT_DOMAIN}
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

## Example

Follow this example to create an APIRule that exposes a sample HTTPBin Deployment.

<!-- tabs:start -->
#### Kyma Dashboard

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
    Alternatively, you can replace these values and use your custom Gateway. See [Introduction to Istio Gateways](./01-20-set-up-tls-gateway.md) and [Set Up a TLS Gateway](./01-10-setup-custom-domain-for-workload.md).
5. Add the host `httpbin.${PARENT_DOMAIN}`.
  
  To learn what your default parent domain is, go to the Kyma Environment section of your subaccount overview, and copy the part of the **APIServerURL** link after `https://api.`. For example, if your **APIServerURL** link is `https://api.c123abc.kyma.ondemand.com`, use `httpbin.c123abc.kyma.ondemand.com` as the host. If you use a custom Gateway, add the host configured in the Gateway.

6. Add a rule with the following configuration:
    - **Path**: `/post`
    - **Handler**: `No Auth`
    - **Methods**: `POST`
7. Add one more rule with the following configuration:
    - **Path**: `/{**}`
    - **Handler**: `No Auth`
    - **Methods**: `GET`
8. Choose **Create**.

#### Kubectl

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    NAMESPACE="test"
    kubectl create ns "${NAMESPACE}"
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite
    ```

2. Get the default domain of your Kyma cluster.

    ```bash
    PARENT_DOMAIN=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.domain}")
    WORKLOAD_DOMAIN="httpbin.${PARENT_DOMAIN}"
    GATEWAY="kyma-system/kyma-gateway"
    echo "Parent domain: ${PARENT_DOMAIN}"
    echo "Workload domain: ${WORKLOAD_DOMAIN}"
    echo "Gateway namespace and name: ${GATEWAY}"
    ```

    This procedure uses the default domain of your Kyma cluster and the default Gateway. Alternatively, you can replace these values and use your custom domain and Gateway instead. See [Introduction to Istio Gateways](./01-20-set-up-tls-gateway.md) and [Set Up a TLS Gateway](./01-10-setup-custom-domain-for-workload.md).

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

3. Expose the workload with an APIRule using the **noAuth** access strategy.

    ```bash
    cat <<EOF | kubectl apply -n "${NAMESPACE}" -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: apirule-noauth
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
    kubectl get apirules apirule-noauth -n "${NAMESPACE}" 
    ```


### Results

You can access your Service at `https://${WORKLOAD_DOMAIN}`.

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET "https://${WORKLOAD_DOMAIN}/ip"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST "https://${WORKLOAD_DOMAIN}/post" -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.
