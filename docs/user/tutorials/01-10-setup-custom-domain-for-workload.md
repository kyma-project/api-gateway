# Set Up a Custom Domain for a Workload

This tutorial shows how to set up a custom domain and prepare a certificate required for exposing a workload. It uses the Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components.

> [!NOTE]
> Skip this tutorial if you use a Kyma domain instead of your custom domain.

## Prerequisites

* [Deploy a sample HTTPBin Service](./01-00-create-workload.md).
* If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management#quick-start) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated namespace. SAP BTP, Kyma runtime clusters are managed by Gardener so you are not required to install any additional components.

## Steps

1. Create a Secret containing credentials for the DNS cloud service provider account in your namespace. To learn how to do it, follow the [External DNS Management guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management).
    
    <!-- tabs:start -->
    #### **Kyma Dashboard**
    
    1. Select the namespace you want to use.
    2. Go to **Configuration > Secretes**.
    3. Select **Create Secret** and provide your configuration details.
    4. Select **Create**.

    #### **kubectl**
    
    Export the name of the created Secret as an environment variable:

    ```bash
    export SECRET={SECRET_NAME}
    ```
    <!-- tabs:end -->