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
2. Go to **Configuration > Secrets**.
3. Select **Create Secret** and provide your configuration details.
4. Select **Create**.

#### **kubectl**
Use `kubectl apply` to create a Secret containing the credentials and export its name as an environment variable:

```bash
export SECRET={SECRET_NAME}
```
<!-- tabs:end -->

2. Create a DNSProvider custom resource (CR).
    
<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Configuration > DNS Providers**.
2. Select **Create DNS Provider**, switch to the `Advanced` tab, and provide the details:
    - **Name**: `dns-provider`
    - **Type**: is the type of your DNS cloud service provider.
    - Add the annotation **dns.gardener.cloud/class**: `garden`
    - In the `Secret Reference` section, add these fields:
        - **Namespace**: is the name of the namespace in which you created the Secret containing the credentials. 
        - **Name**: is the name of the Secret.
    - In the `Include Domains` section, add the field:
        - **Include Domains**: is the name of your custom domain.
3. Select **Create**.

#### **kubectl**

1. Export the following values as environment variables. Replace `PROVIDER_TYPE` with the type of your DNS cloud service provider. `DOMAIN_NAME` value specifies the name of your custom domain, for example, `mydomain.com`.

    ```bash
    export PROVIDER_TYPE={YOUR_PROVIDER_TYPE}
    export DOMAIN_TO_EXPOSE_WORKLOADS={YOUR_DOMAIN_NAME} 
    ````
2. To create a DNSProvider CR, run: 

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: dns-provider
      namespace: $NAMESPACE
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      type: $SPEC_TYPE
      secretRef:
        name: $SECRET
      domains:
        include:
          - $DOMAIN_TO_EXPOSE_WORKLOADS
    EOF
    ```
<!-- tabs:end -->

3. Create a DNSEntry CR.

    <!-- tabs:start -->
    #### **Kyma Dashboard**
    g
    #### **kubectl**
   
    g
    <!-- tabs:end -->

4. Create a Certificate CR.
    
    > [!NOTE]
    > While using the default configuration, certificates with the Let's Encrypt Issuer are valid for 90 days and automatically renewed 30 days before their validity expires. For more information, read the documentation on [Gardener Certificate Management](https://github.com/gardener/cert-management#requesting-a-certificate) and [Gardener extensions for certificate Services](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/).

    <!-- tabs:start -->
    #### **Kyma Dashboard**

    h

    #### **kubectl**

    h

    <!-- tabs:end -->

5. [Set Up a TLS Gateway](./01-20-set-up-tls-gateway.md) or [Set up an mTLS Gateway](./01-30-set-up-mtls-gateway.md).

Visit the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples) to see more examples of CRs for Services and Ingresses.