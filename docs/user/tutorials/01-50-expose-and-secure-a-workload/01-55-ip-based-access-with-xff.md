# Use the XFF Header to Configure IP-Based Access to a Workload

Expose your workload and configure IP-based access using the X-Forwarded-For (XFF) header. This helps to enhance security by ensuring that only trusted IPs can interact with your application.

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have a deployed workload.
* To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/). Alternatively, you can use Kyma dashboard.
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}`.

## Context

The XFF header is a standard HTTP header that conveys the client IP address and the chain of intermediary proxies that the request traverses to reach the Istio service mesh. This is particularly useful when an application must be provided with the client IP address of an originating request, for example, for access control.

However, you must be aware of some technical limitations when using the XFF header. The header might not include all IP addresses if an intermediary proxy does not support modifying the header. Due to [technical limitations of AWS Classic ELBs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-proxy-protocol.html#proxy-protocol), when using an IPv4 connection, the header does not include the public IP of the load balancer in front of Istio Ingress Gateway. Moreover, Istio Ingress Gateway's Envoy does not append the private IP address of the load balancer to the XFF header, effectively removing this information from the request. For more information on XFF, see the [IETF’s RFC documentation](https://datatracker.ietf.org/doc/html/rfc7239) and [Envoy documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for).

To use the XFF header, you must configure the corresponding settings in the Istio custom resource (CR). Then, expose your workload using an APIRule CR and create an AuthorizationPolicy resource with allowed IP addresses specified in the **remoteIpBlocks** field. To learn how to do this, follow the procedure.

## Procedure

<!-- tabs:start -->
#### **Kyma Dashboard**
1. To use the XFF header, configure the number of trusted proxies in the Istio custom resource.
   1. Go to **Cluster Details** and choose **Modify Modules**.
   2. Select the Istio module and choose **Edit**.
   3. In the `General` section, set the number of trusted proxies to `1`.
     Due to the variety of network topologies, the Istio CR must specify the number of trusted proxies deployed in front of the Istio Ingress Gateway proxy in the Istio CR so that the client address can be extracted correctly.
   4. If you use a Google Cloud or Azure cluster, navigate to the Gateway section and set the Gateway traffic policy to Local. If you use a different cloud service provider, skip this step.
      >[!WARNING]
      > For production Deployments, deploying Istio Ingress Gateway Pod to multiple nodes is strongly recommended if you enable `externalTrafficPolicy : Local`. For more information, see [Network Load Balancer](https://istio.io/latest/docs/tasks/security/authorization/authz-ingress/#network).
      >
      >Default Istio installation profile configures PodAntiAffinity to ensure that Ingress Gateway Pods are evenly spread across all nodes and, if possible, across different zones. This guarantees that the above requirement is satisfied if your IngressGateway autoscaling configuration minReplicas is equal to or greater than the number of nodes. You can configure autoscaling options in the Istio CR using the field **spec.config.components.ingressGateway.k8s.hpaSpec.minReplicas**.<br>
      
      >[!TIP]
      > If you use a Google Cloud or Azure cluster, you can find your load balancer's IP address in the field **status.loadBalancer.ingress** of the `ingress-gateway` Service.
   5. Choose **Save**.
2. To expose your workload, create an APIRule custom resource. 
     1. Go to **Discovery and Network > API Rules** and choose **Create**.
     2. Provide all the required configuration details.
     3. Choose **Create**.
   
    **Step result:** When you go to `https:/{SUBDOMAIN}.{DOMAIN}/{PATH}`, the response contains the **X-Forwarded-For** and **X-Envoy-External-Address** headers with your public IP address. See an example response for the Client IP `165.1.187.197`:
      ```json
      {
        "args": {
          "show_env": "true"
        },
        "headers": {
          "Accept": "...",
          "Host": "...",
          "User-Agent": "...",
          "X-Envoy-Attempt-Count": "...",
          "X-Envoy-External-Address": "165.1.187.197",
          "X-Forwarded-Client-Cert": "...",
          "X-Forwarded-For": "165.1.187.197",
          "X-Forwarded-Proto": "...",
          "X-Request-Id": "..."
        },
        "origin": "165.1.187.197",
        "url": "..."
      }
      ``` 

      >[!TIP]
      > You can check your public IP address at https://api.ipify.org.

3. To configure IP-based access to the exposed workload, create an AuthorizationPolicy resource.
    1. Go to **Istio > Authorization Policies** and choose **Create**.
    2. Add a selector to specify the workload for which access should be configured.
    3. Add a rule with a **From** field.
    4. In the **RemoteIpBlocks** field, specify the IP addresses that should be allowed access to the workload.
    5. Choose **Create**.


#### **kubectl**
1. To use the XFF header, configure the number of trusted proxies in the Istio custom resource.
   1. Run the following command:
       ```bash
       kubectl patch istios/default -n kyma-system --type merge -p '{"spec":{"config":{"numTrustedProxies": 1}}}'
       ```
       Due to the variety of network topologies, the Istio CR must specify the configuration property **numTrustedProxies**, so that the client IP address can be extracted correctly.

   2. If you use a Google Cloud or Azure cluster, run the following command to set the traffic policy to `Local`. If you use a different cloud service provider, skip this step.

       ```bash
       kubectl patch istios/default -n kyma-system --type merge -p '{"spec":{"config":{"gatewayExternalTrafficPolicy": "Local"}}}'
       ```
       >[!WARNING]
       > For production Deployments, it is strongly recommended that you deploy Istio Ingress Gateway Pod to multiple nodes if you enable `externalTrafficPolicy : Local`. For more information, see [Network Load Balancer](https://istio.io/latest/docs/tasks/security/authorization/authz-ingress/#network).
       >
       >Default Istio installation profile configures **PodAntiAffinity** to ensure that Ingress Gateway Pods are evenly spread across all nodes and, if possible, across different zones. This guarantees that the above requirement is satisfied if your IngressGateway autoscaling configuration **minReplicas** is equal to or greater than the number of nodes. You can configure autoscaling options in the Istio CR using the field **spec.config.components.ingressGateway.k8s.hpaSpec.minReplicas**.
       
       >[!TIP]
       > If you use a Google Cloud or Azure cluster, you can find your load balancer's IP address in the field **status.loadBalancer.ingress** of the `ingress-gateway` Service.

2. To expose your workload, create an APIRule custom resource. 
  
    You can adjust this sample configuration and use another access strategy, according to your needs.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: {APIRULE_NAME}
      namespace: {APIRULE_NAMESPACE}
    spec:
      host: {SUBDOMAIN}.{DOMAIN}
      service:
        name: {SERVICE_NAME}
        namespace: {SERVICE_NAMESPACE}
        port: {SERVICE_PORT}
      gateway: {GATEWAY_NAME}/{GATEWAY_NAMESPACE}
      rules:
        - path: /headers
          methods: ["GET"]
          accessStrategies:
            - handler: no_auth
    EOF
    ```
   
    **Step result:** When you run `curl -ik -X GET https:/{SUBDOMAIN}.{DOMAIN}/headers`, the response contains the **X-Forwarded-For** and **X-Envoy-External-Address** headers with your public IP address. See an example response for the Client IP `165.1.187.197`:
    ```json
    {
      "args": {
        "show_env": "true"
      },
      "headers": {
        "Accept": "...",
        "Host": "...",
        "User-Agent": "...",
        "X-Envoy-Attempt-Count": "...",
        "X-Envoy-External-Address": "165.1.187.197",
        "X-Forwarded-Client-Cert": "...",
        "X-Forwarded-For": "165.1.187.197",
        "X-Forwarded-Proto": "...",
        "X-Request-Id": "..."
      },
      "origin": "165.1.187.197",
      "url": "..."
    }
    ``` 
    >[!TIP]
    > You can check your public IP address at https://api.ipify.org.    

3. To configure IP-based access to the exposed workload, create an AuthorizationPolicy resource.

  The selector specifies the workload for which access should be configured, and the **RemoteIpBlocks** field specifies the IP addresses for which access should be allowed.
    
    ```bash
    apiVersion: security.istio.io/v1beta1
    kind: AuthorizationPolicy
    metadata:
      name: {AUTHORIZATIONPOLICY_NAME}
      namespace: {AUTHORIZATIONPOLICY_NAMESPACE}
    spec:
      action: ALLOW
      rules:
        - from:
            - source:
                ipBlocks: []
                remoteIpBlocks:
                  - {ALLOWED_IP}
      selector:
        matchLabels:
          {KEY}: {VALUE}
    ```
<!-- tabs:end -->

### Results
You have configured the XFF header in the Istio CR and exposed your workload to the internet. Access to the workload is limited to the IP addresses that you have specified.