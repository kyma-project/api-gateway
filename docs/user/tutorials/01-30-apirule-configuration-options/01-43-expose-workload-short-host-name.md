# Use a Short Host Name in APIRule

Simplify your APIRule configuration by using a short host name instead of a fully qualified domain name (FQDN).

A short host name is a single lowercase [RFC 1123](https://datatracker.ietf.org/doc/html/rfc1123) subdomain label without the domain suffix. For example, instead of using the FQDN `myapp.example.com`, you use just `myapp`. The domain is automatically appended from the referenced Gateway resource. This might be particularly useful when reconfiguring resources in a new cluster, as it reduces the chance of errors and streamlines the process.

The referenced Gateway must:
- Define a single wildcard host across all [Server](https://istio.io/latest/docs/reference/config/networking/gateway/#Server) definitions
- Use the wildcard prefix `*.` (for example, `*.example.com`)

See an example Gateway resource:

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: kyma-gateway
  namespace: kyma-system
spec:
  selector:
    app: istio-ingressgateway
  servers:
  - hosts:
    - "*.example.com"
    port:
      name: https
      number: 443
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: example-com-tls
```

The following APIRule uses a short host name and references `kyma-gateway`:

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin-short
  namespace: default
spec:
  hosts:
    - myapp
  service:
    name: httpbin
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
```

When you use the short host name `myapp`, the API Gateway automatically expands it to `myapp.example.com` by retrieving the domain from the referenced Gateway.

## Example

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
    Alternatively, you can replace these values and use your custom Gateway. See [Introduction to Istio Gateways](../00-05-domains-and-gateways.md) and [Set Up a TLS Gateway](./01-20-set-up-tls-gateway.md).
5. Add the host `example`.
7. Add a rule with the following configuration:
    - `Path: /{**}`
    - `Handler: No Auth`
    - `Methods: GET`
8. Choose **Create**.
9. Copy the Hosts link from the Virtual Service section and paste it in your browser.

    If successful, the httpbin.org page is displayed. The full domain name is appended to the URL, even though it is not explicitly specified in the APIRule itself.

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

3. Expose the workload with an APIRule. You do not specify the full domain name in the APIRule's configuration.

    ```bash
    cat <<EOF | kubectl apply -n "${NAMESPACE}" -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: apirule-noauth
    spec:
      hosts:
        - example
      service:
        name: httpbin
        namespace: ${NAMESPACE}
        port: 8000
      gateway: kyma-system/kyma-gateway
      rules:
        - path: /{**}
          methods: ["GET"]
          noAuth: true
    EOF
    ```

    Check if the APIRule's status is ready:

    ```bash
    kubectl get apirules apirule-noauth -n "${NAMESPACE}" 
    ```

4. You can access your Service at `https://${WORKLOAD_DOMAIN}`. To test the connection, send the `GET` request to the exposed workload:

    ```bash
    curl -ik -X GET "https://${WORKLOAD_DOMAIN}/ip"
    ```
  
    If successful, the calls return the `200 OK` response code. The full domain name is appended to the workload's hostname, even though it is not explicitly specified in the APIRule itself.

<!-- tabs:end -->