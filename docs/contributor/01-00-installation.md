# Install API Gateway

## Prerequisites

- For API Gateway to work, the [Istio module](https://github.com/kyma-project/istio) must be installed in the cluster.
- Access to a Kubernetes cluster (you can use [k3d](https://k3d.io/v5.5.1/))
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kubebuilder](https://book.kubebuilder.io/)
- [Docker](https://www.docker.com)
- [Kyma CLI](https://github.com/kyma-project/cli/blob/main/README.md#installation)

## Install Kyma API Gateway Operator Manually

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/api-gateway.git && cd api-gateway
    ```

2. Set the API Gateway Operator image name.

    ```bash
    export IMG=api-gateway-operator:0.0.1
    export K3D_CLUSTER_NAME=kyma
    ```

3. Provision the k3d cluster.

    ```bash
    kyma provision k3d
    ```
    >**TIP:** To verify the correctness of the project, build it using the `make build` command.

4. Build the image.

    ```bash
    make docker-build
    ```

5. Push the image to the registry.

    <div tabs name="Push image" group="api-gateway-operator-installation">
      <details>
      <summary label="k3d">
      k3d
      </summary>

      ```bash
      k3d image import $IMG -c $K3D_CLUSTER_NAME
      ```

      </details>
      <details>
      <summary label="Docker registry">
      Globally available Docker registry
      </summary>

      ```bash
      make docker-push
      ```

      </details>
    </div>

6. Create the `kyma-system` namespace and deploy API Gateway Operator in it.

    ```bash
    make deploy
    ```