# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose with [JWT](./01-40-expose-workload-jwt.md).

## Prerequisites

- You have Istio and API Gateway modules in your cluster. See [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html) for open-source Kyma.

## Steps

To expose your workload, create an APIRule. For each path you want to expose unsecured, configure a rule with `noAuth: true`.

>[!NOTE]
> To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/external-content/istio/docs/user/tutorials/01-40-enable-sidecar-injection.html#enable-istio-sidecar-proxy-injection).

<!-- tabs:start -->
#### **Kyma Dashboard**
1. Go to **Discovery and Network > API Rules** and choose **Create**. 
2. Provide all the required configuration details.
3. Add a rule with the following configuration.
    - **Access Strategy**: `noAuth`
    - Add allowed methods and the request path.
4. Choose **Create**.  

#### **kubectl**

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

See the following example APIRule that exposes a sample HTTPBin Deployment:

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
  To learn what your default domain is, go to the Kyma Environment section of your subaccount overview, and copy the part of the **APIServerURL** link after `https://api.`. For example, if your **APIServerURL** link is `https://api.c123abc.kyma.ondemand.com`, use `httpbin.c123abc.kyma.ondemand.com` as the host. If you use a custom Gateway, add the host configured in the Gateway.
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
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```

2. Get the default domain of your Kyma cluster.

    ```bash
    GATEWAY_DOMAIN=$(kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts[0]}')
    WORKLOAD_DOMAIN=httpbin.${GATEWAY_DOMAIN#*.}
    GATEWAY=kyma-system/kyma-gateway
    echo "Parent domain: ${PARENT_DOMAIN}"
    echo "Workload domain: ${WORKLOAD_DOMAIN}"
    echo "Gateway name and namespace: ${GATEWAY}"
    ```

    This procedure uses the default domain of your Kyma cluster and the default Gateway. Alternatively, you can replace these values and use your custom domain and Gateway instead. See [Introduction to Istio Gateways](./01-20-set-up-tls-gateway.md) and [Set Up a TLS Gateway](./01-10-setup-custom-domain-for-workload.md).

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

3. Expose the workload with an APIRule using the **noAuth** access strategy.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: apirule-noauth
      namespace: test
    spec:
      hosts:
        - ${WORKLOAD_DOMAIN}
      service:
        name: httpbin
        namespace: test
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
    kubectl --namespace=oauth2-proxy get apirules httpbin -n httpbin-system
    ```


### Results

You can access your Service at `https://${WORKLOAD_DOMAIN}`.

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET https://${WORKLOAD_DOMAIN}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST https://${WORKLOAD_DOMAIN}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.
