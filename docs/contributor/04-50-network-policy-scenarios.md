# Istio NetworkPolicy scenarios

Prerequisites for all scenarios:
- NetworkPolicies are configured for `deny-all` by default.
- `istio-system` and `kyma-system` namespaces have necessary NetworkPolicies
  to correctly install and run Istio and Kyma components.
- The Istio module is installed in your Kubernetes cluster.

1. Create namespaces `istio-system` and `kyma-system`

   ```bash
   kubectl create namespace istio-system
   kubectl create namespace kyma-system
   ```
2. Create `deny-all` NetworkPolicies

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: kyma-system
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: istio-system
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```
   
3. Install Istio module in the cluster. Here, the OS installation is used for convenience

   ```bash
   kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
   kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
   ```
   
   At this point, Istio will not work properly as the deny-all NetworkPolicies block the communication.

4. Install NetworkPolicies required for Istio to work correctly

   Allow DNS traffic in both namespaces:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-dns
     namespace: kyma-system
   spec:
     policyTypes:
     - Egress
     egress:
     - ports:
       - protocol: UDP
         port: 53
       - protocol: TCP
         port: 53
   EOF
   ```
   
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-dns
     namespace: istio-system
   spec:
     policyTypes:
     - Egress
     egress:
     - ports:
       - protocol: UDP
         port: 53
       - protocol: TCP
         port: 53
   EOF
   ```

   Allow APIServer access for istio operator, istiod and install-cni:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: istio-operator-allow-apiserver
     namespace: kyma-system
   spec:
     podSelector:
       matchLabels:
         app.kubernetes.io/name: istio-operator
     policyTypes:
     - Egress
     egress:
     - ports:
       - protocol: TCP
         port: 443
   EOF
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: istio-pilot-allow-apiserver
     namespace: istio-system
   spec:
     podSelector:
       matchLabels:
         istio: pilot
     policyTypes:
     - Egress
     egress:
     - ports:
       - protocol: TCP
         port: 443
   EOF
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: istio-cni-allow-apiserver
     namespace: istio-system
   spec:
     podSelector:
       matchLabels:
         k8s-app: istio-cni-node
     policyTypes:
     - Egress
     egress:
     - ports:
       - protocol: TCP
         port: 443
   EOF
   ```
   
   Allow communication between pilot and ingressgateway:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: ingress-gateway-allow-egress
     namespace: istio-system
   spec:
     podSelector:
       matchLabels:
         istio: ingressgateway
     policyTypes:
     - Egress
     egress:
     - to:
       - podSelector:
           matchLabels:
             istio: pilot
     - ports:
       - protocol: TCP
         port: 80
       - protocol: TCP
         port: 443
   
       - protocol: TCP
         port: 15006
       - protocol: TCP
         port: 15012
   EOF
   ```
   
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: pilot-allow-ingress
     namespace: istio-system
   spec:
     podSelector:
       matchLabels:
         istio: pilot
     policyTypes:
     - Ingress
     ingress:
     - ports:
       # Envoy outbound
       - protocol: TCP
         port: 15001
       # Listen port for failure detection
       - protocol: TCP
         port: 15002
       from:
       - podSelector:
           matchLabels:
             security.istio.io/tlsMode: istio
       - podSelector:
           matchLabels:
             istio: ingressgateway
     - ports:
       # XDS and CA services
       - protocol: TCP
         port: 15010
       - protocol: TCP
         port: 15012
   
       # Webhook
       - protocol: TCP
         port: 15017
   
       # Control plane monitoring
       - protocol: TCP
         port: 15014
       # Health
       - protocol: TCP
         port: 15021
       # Metrics
       - protocol: TCP
         port: 15090
   EOF
   ```
   
## Enabling sidecar injection for a deployment

![sidecar-netpol.svg](../assets/sidecar-netpol.svg)

This scenario demonstrates how to enable Istio sidecar injection and allow istiod to configure the sidecar proxy.

1. Create a new namespace with sidecar injection enabled

   ```bash
   kubectl create namespace my-namespace
   kubectl label namespace my-namespace istio-injection=enabled
   ```
   
2. Create `deny-all` NetworkPolicy for the new namespace

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: my-namespace
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```

3. Deploy a sample application in the new namespace

    ```bash
   kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml -n my-namespace
   ```
   
   At this point, the application will not work properly as the sidecar proxy cannot communicate with istiod.

4. Create NetworkPolicies to allow communication between the sidecar proxy and istiod

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sidecar-allow
     namespace: my-namespace
   spec:
     podSelector:
       matchLabels:
         app: httpbin
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: istio-system
         podSelector:
           matchLabels:
             istio: pilot
       ports:
       - protocol: TCP
         port: 15012
   EOF
   ```

## Sidecar-to-sidecar communication with deny-all NetworkPolicy

![sidecar-to-sidecar.svg](../assets/sidecar-to-sidecar.svg)

This scenario demonstrates how to allow communication between two sidecar proxies in different namespaces with deny-all NetworkPolicies.

1. Create two new namespaces with sidecar injection enabled

   ```bash
   kubectl create namespace namespace-a
   kubectl label namespace namespace-a istio-injection=enabled
   kubectl create namespace namespace-b
   kubectl label namespace namespace-b istio-injection=enabled
   ```
   
2. Create `deny-all` NetworkPolicies for both namespaces

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: namespace-a
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```
   
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: namespace-b
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```
   
3. Apply NetworkPolicies to allow communication between sidecar proxies and istiod in both namespaces

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sidecar-allow
     namespace: namespace-a
   spec:
     podSelector:
       matchLabels:
         app: nginx
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: istio-system
         podSelector:
           matchLabels:
             istio: pilot
       ports:
       - protocol: TCP
         port: 15012
   EOF
   ```
   
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sidecar-allow
     namespace: namespace-b
   spec:
     podSelector:
       matchLabels:
         app: nginx
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: istio-system
         podSelector:
            matchLabels:
              istio: pilot
       ports:
       - protocol: TCP
         port: 15012
   EOF
   ```
   
4. Deploy sample applications in both namespaces

   ```bash
   kubectl create deployment nginx --image=nginx -n namespace-a
   kubectl expose deployment nginx --port=80 -n namespace-a
   kubectl create deployment nginx --image=nginx -n namespace-b
   kubectl expose deployment nginx --port=80 -n namespace-b
   ```
   
   At this point, the applications will not be able to communicate with each other due to the deny-all NetworkPolicies.
   ```bash
   kubectl exec -it $(kubectl get pod -n namespace-a -l app=nginx -o jsonpath='{.items[0].metadata.name}') -n namespace-a -- curl http://nginx.namespace-b.svc.cluster.local:80
   ```

   ```
   $ upstream connect error or disconnect/reset before headers. retried and the latest reset reason: connection timeout
   ```

   ```bash
   kubectl exec -it $(kubectl get pod -n namespace-b -l app=nginx -o jsonpath='{.items[0].metadata.name}') -n namespace-b -- curl http://nginx.namespace-a.svc.cluster.local:80
   ```
   
   ```
   $ upstream connect error or disconnect/reset before headers. retried and the latest reset reason: connection timeout
   ```

5. Create NetworkPolicies to allow communication between the sidecar proxies in both namespaces

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
    name: allow-communication-to-ns-b
     namespace: namespace-a
   spec:
     podSelector:
        matchLabels:
          app: nginx
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: namespace-b
         podSelector:
           matchLabels:
             app: nginx
       ports:
       - protocol: TCP
         port: 80
   EOF
   ```
   
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-inbound-from-ns-a
     namespace: namespace-b
   spec:
     podSelector:
       matchLabels:
         app: nginx
     policyTypes:
     - Ingress
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: namespace-a
         podSelector:
           matchLabels:
             app: nginx
       ports:
       - protocol: TCP
         port: 80
   EOF
   ```
   
   This should allow communication from namespace-a to namespace-b:
   ```bash
   kubectl exec -it $(kubectl get pod -n namespace-a -l app=nginx -o jsonpath='{.items[0].metadata.name}') -n namespace-a -- curl http://nginx.namespace-b.svc.cluster.local:80
   ```
   
   ```
   $ <html>...</html>
   ```
   
   However, communication from namespace-b to namespace-a will still be blocked:
   To enable two-way communication, create the corresponding egress and ingress NetworkPolicies in namespace-b and namespace-a respectively.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-inbound-from-ns-b
     namespace: namespace-a
   spec:
     podSelector:
       matchLabels:
         app: nginx
     policyTypes:
     - Ingress
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: namespace-b
         podSelector:
           matchLabels:
             app: nginx
       ports:
       - protocol: TCP
         port: 80
   EOF
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-coomunication-to-ns-a
     namespace: namespace-b
   spec:
     podSelector:
        matchLabels:
          app: nginx
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: namespace-a
         podSelector:
           matchLabels:
             app: nginx
       ports:
       - protocol: TCP
         port: 80
   EOF
   ```
   
   The communication should now work in both directions:
   ```bash
   kubectl exec -it $(kubectl get pod -n namespace-b -l app=nginx -o jsonpath='{.items[0].metadata.name}') -n namespace-b -- curl http://nginx.namespace-a.svc.cluster.local:80
   ```
   
   ```
   $ <html>...</html>
   ```

## Expose workload via VirtualService with deny-all NetworkPolicy

![expose-netpol.svg](../assets/expose-netpol.svg)

This scenario demonstrates how to expose a workload using Istio VirtualService while having deny-all NetworkPolicies in place.
1. Create a new namespace with sidecar injection enabled

   ```bash
   kubectl create namespace my-virtualservice-namespace
   kubectl label namespace my-virtualservice-namespace istio-injection=enabled
   ```
   
2. Create `deny-all` NetworkPolicy for the new namespace

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: default-deny
     namespace: my-virtualservice-namespace
   spec:
     podSelector: {}
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```

3. Deploy NetworkPolicy allowing communication between sidecar and istiod

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sidecar-allow
     namespace: my-virtualservice-namespace
   spec:
     podSelector:
       matchLabels:
         app: httpbin
     policyTypes:
     - Egress
     egress:
     - to:
       - namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: istio-system
         podSelector:
           matchLabels:
             istio: pilot
       ports:
       - protocol: TCP
         port: 15012
   EOF
   ```
   
4. Deploy a sample application

   ```bash
   kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml -n my-virtualservice-namespace
   ```
   
5. Create a VirtualService to expose the application via Istio ingressgateway

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1beta1
   kind: VirtualService
   metadata:
     name: httpbin
     namespace: my-virtualservice-namespace
   spec:
     hosts:
     - "*"
     gateways:
     - kyma-system/kyma-gateway
     http:
     - route:
       - destination:
           host: httpbin.my-virtualservice-namespace.svc.cluster.local
           port:
             number: 8000
       timeout: 5s
   EOF
   ```
   
   Without the necessary NetworkPolicies, accessing the application through the ingressgateway will fail.
   The timeout error will be observed when trying to access the application:
   ```bash 
   curl https://my-cluster-domain.com/httpbin/status/200 -i
   ```
   
   ```
   $ HTTP/2 504
   $ content-length: 24
   $ content-type: text/plain
   $ date: Thu, 22 Jan 2026 20:10:02 GMT
   $ server: istio-envoy

   $ upstream request timeout
    ```

6. Create NetworkPolicies to allow ingress traffic to the application from the Istio ingressgateway

   Allow ingress traffic to the sidecar from the Istio ingressgateway:
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-ingress-to-sidecar
     namespace: my-virtualservice-namespace
   spec:
     podSelector:
       matchLabels:
         app: httpbin
     policyTypes:
     - Ingress
     ingress:
     - from:
       - podSelector:
           matchLabels:
              istio: ingressgateway
         namespaceSelector:
           matchLabels:
             kubernetes.io/metadata.name: istio-system
       ports:
       - protocol: TCP
         # Needs to be the same as the internal port of the httpbin service
         # This should be the same as Service targetPort
         port: 8080
   EOF
   ```

   Allow egress traffic from the Istio ingressgateway to the application:
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-egress-from-ingressgateway
     namespace: istio-system
   spec:
       podSelector:
         matchLabels:
           istio: ingressgateway
       policyTypes:
       - Egress
       egress:
       - to:
         - namespaceSelector:
             matchLabels:
               kubernetes.io/metadata.name: my-virtualservice-namespace
           podSelector:
             matchLabels:
               app: httpbin
         ports:
         - protocol: TCP
           # Needs to be the same as the internal port of the httpbin service
           # This should be the same as Service targetPort
           port: 8080
   EOF
   ```

   After applying these NetworkPolicies, accessing the application through the ingressgateway should work:
   ```bash
   curl https://my-cluster-domain.com/httpbin/status/200 -i
   ```

   ```
   $ HTTP/2 200
   $ access-control-allow-credentials: true
   $ access-control-allow-origin: *
   $ content-type: application/json; charset=utf-8
   $ date: Thu, 22 Jan 2026 20:23:59 GMT
   $ content-length: 745
   $ x-envoy-upstream-service-time: 1
   $ server: istio-envoy
   ```