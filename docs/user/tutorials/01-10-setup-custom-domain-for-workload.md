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
    <ol>
      <li> Go to <b>Configuration > Secrets</b>. </li>
      <li> Select <b>Create Secret</b> and provide your configuration details. </li>
      <li> Select <b>Create</b>. </li>
    </ol>

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
        - Add the following annotation:
          - **dns.gardener.cloud/class**: `garden`
        - In the `Secret Reference` section, add these fields:
          - **Namespace**: is the name of the namespace in which you created the Secret containing the credentials. 
          - **Name**: is the name of the Secret.
        - In the `Include Domains` section, add the name of your custom domain.
    3. Select **Create**.

    #### **kubectl**
    <ol>
      <li> Export the following values as environment variables. Replace <code>PROVIDER_TYPE</code> with the type of your DNS cloud service provider. <code>DOMAIN_NAME</code> value specifies the name of your custom domain, for example, <code>mydomain.com</code>.

      ```bash
      export PROVIDER_TYPE={YOUR_PROVIDER_TYPE}
      export DOMAIN_TO_EXPOSE_WORKLOADS{YOUR_DOMAIN_NAME} 
      ```
      </li>
      <li>To create a DNSProvider CR, run: 

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
      </li>
    </ol>
  <!-- tabs:end -->

3. Create a DNSEntry CR.

    <!-- tabs:start -->
    #### **Kyma Dashboard**
    <ol>
    
    <li> 
    Go to <b>Discovery and Network > Services</b> in the <code>istio-system</code> namespace. Select the <code>istio-ingressgateway</code> Service and copy its external IP address.
    </li>
    <li>
    In the namespace of your HTTPBin Deployment, go to <b>Configuration > DNS Entries</b>.
    </li>
    <li>
    Select <b>Create DNS Provider</b>, switch to the <code>Advanced</code> tab, and provide the details:
      <ul>    
        <li>
        <b>Name</b>:<code>dns-entry</code>
        </li>
        <li>
        Add the annotation:
          <ul><li>
          <b>dns.gardener.cloud/class</b>: <code>garden</code>
          </ul></li>
        </li> 
        <li>
        For <b>DNSName</b>, use <code>*.{DOMAIN_TO_EXPOSE_WORKLOADS}</code>. Replace <code>{DOMAIN_TO_EXPOSE_WORKLOADS}</code> with the name of your custom domain.
        </li>
        <li>
        - <b>TTL</b>:<code>600</code>
        </li>
        <li>
        Paste the external IP address of the <codeistio-ingressgateway</code> Service in the <b>Target</b> field.
        </li>
      </ul>
    <li> 
    Select <b>Create</b>.
    </li>
    </ol>

    #### **kubectl**
    <ol>
      <li> Export the following values as environment variables:

      ```bash
      export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # Assuming only one LoadBalancer with external IP
      ```
      > [!NOTE]
      > For some cluster providers you need to replace the `ip` with the `hostname`, for example, in AWS, set `jsonpath='{.status.loadBalancer.ingress[0].hostname}'`.
      </li>
      <li> To create a DNSEntry CR, run:
      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: dns.gardener.cloud/v1alpha1
      kind: DNSEntry
      metadata:
        name: dns-entry
        namespace: $NAMESPACE
        annotations:
          dns.gardener.cloud/class: garden
      spec:
        dnsName: "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
        ttl: 600
        targets:
          - $IP
      EOF
      ```
      </li>
    </ol>
    <!-- tabs:end -->

4. Create a Certificate CR.
    
    <!-- tabs:start -->
    #### **Kyma Dashboard**
    <ol>
    <li> Go to the <code>istio-system</code> namespace.</li>
    <li> Go to <b>Configuration > Certificates</b>. </li>
    <li> Select <b>Create Certificate</b> and provide the details:
    <ul>
      <li> <b>Name</b><code>httpbin-cert</code></li>
      <li> <b>Secret Name</b> is the name of your TLS Secret. </li>
      <li> <b>Custom Name</b> is the name of your custom domain </li>
    </ul>
    </li>
    <li> Select <b>Create</b>.</li>

    #### **kubectl**
    <ol>
    <li> Export the name of your TLS Secret as an environment variable:

    ```bash
    export TLS_SECRET={TLS_SECRET_NAME}
    ```
    </li>
    <li> To create a Certificate CR, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: httpbin-cert
      namespace: istio-system
    spec:  
      secretName: $TLS_SECRET
      commonName: $DOMAIN_TO_EXPOSE_WORKLOADS
    EOF
    ```
    </li>
    <li> To check the certificate status, run: 
     
    ```bash
    kubectl get certificate httpbin-cert -n istio-system
    ```
    </li>
    </ol>
    <!-- tabs:end -->

    > [!NOTE]
    > While using the default configuration, certificates with the Let's Encrypt Issuer are valid for 90 days and automatically renewed 30 days before their validity expires. For more information, read the documentation on [Gardener Certificate Management](https://github.com/gardener/cert-management#requesting-a-certificate) and [Gardener extensions for certificate Services](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/).

5. [Set Up a TLS Gateway](./01-20-set-up-tls-gateway.md) or [Set up an mTLS Gateway](./01-30-set-up-mtls-gateway.md).

Visit the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples) to see more examples of CRs for Services and Ingresses.