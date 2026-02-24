# Exposing and Securing Workloads with APIRules
Use APIRules to expose and secure your workload.

## What It Means to Expose a Workload?

When you “expose a workload” in Kyma, you make a Service inside the cluster reachable from outside the cluster through an ingress endpoint.

By default, your workloads run inside the cluster and are only reachable over the internal network. To allow external clients (such as applications, tools, or users in a browser) to call these workloads, you must choose an external host name, decide which paths and HTTP methods should be reachable, and decide who is allowed to call them and under which conditions.

## What Is an APIRule?

In Kyma, the Istio Ingress Gateway receives the incoming traffic, and the API Gateway module uses your configuration to route requests from the external host and path to the correct Service and port inside the cluster and apply access control, such as JWT-based authentication, external authorization, or unauthenticated (noAuth) access.

An APIRule is a Kyma custom resource that simplifies the API exposure. The API Gateway module watches APIRule resources and translates them into the required Istio configuration (for example, VirtualService, AuthorizationPolicy, and related objects). This way, you work with a focused, Kyma-specific API instead of managing multiple low-level Istio resources yourself.