# Create a Workload

This tutorial explains how to create a sample HTTPBin Service deployment.

## Steps

<!-- tabs:start -->
#### **Kyma dashboard**

1. Create a namespace with enabled Istio sidecar injection and navigate to it.
2. In the **Configuration** section, select **Service Accounts**, and then **Create**. Enter `httpbin` as your Service Account's name and confirm creation.
4. In the **Discovery and Network** section, select **Services**, and then **Create**. Switch to the `Advanced` tab and provide the following configuration details:
    - Name: `httpbin`
    - Labels: `service=httpbin`, `app=httpbin`
    - Selectors: `app=httpbin`
    - In the Ports section, select **Add**. Then, use these values:
      - Name: `http`
      - Protocol: `TCP`
      - Port: `8000`
      - Target Port: `8000`

    To confirm the Service creation, select **Create**.
5. In the **Workloads** section, select **Deployments**, and then **Create**. Toggle the bar and select the HTTPBin preset. To confirm the Deployment's creation, select **Create**.

#### **kubectl**

<!-- tabs:end -->

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Choose a name for your HTTPBin Service instance and export it as an environment variable.

    ```bash
    export SERVICE_NAME={SERVICE_NAME}
    ```

3. Deploy a sample instance of the HTTPBin Service.

    ```shell
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: $SERVICE_NAME
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: $SERVICE_NAME
      labels:
        app: httpbin
        service: httpbin
    spec:
      ports:
      - name: http
        port: 8000
        targetPort: 8000
      selector:
        app: httpbin
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: $SERVICE_NAME
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
          serviceAccountName: $SERVICE_NAME
          containers:
          - image: docker.io/kennethreitz/httpbin
            imagePullPolicy: IfNotPresent
            name: httpbin
            ports:
            - containerPort: 80
    EOF
    ```

4. Verify if an instance of the HTTPBin Service is successfully created.
   
    ```shell
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```
    
    You should get a result similar to this one:
    
    ```shell
    NAME                        READY    STATUS     RESTARTS    AGE
    {SERVICE_NAME}-{SUFFIX}     2/2      Running    0           96s
    ```