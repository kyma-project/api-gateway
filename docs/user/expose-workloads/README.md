# Expose and Secure Workloads
Use APIRules to expose and secure your workloads in Kyma.

When you expose a workload in Kyma, you make a Kubernetes Service reachable from outside the cluster. By default, workloads run only inside the cluster. To allow external clients (applications, tools, or users in a browser) to call a workload, you publish it under a host name such as `myapp.example.com`, decide which paths and HTTP methods can be called (for example, `GET /headers`), and choose whether the endpoint is open or requires authentication.

## What Is an APIRule?

An APIRule is a Kyma custom resource that describes how a workload is exposed.

You use an APIRule to define the following:

- Which host names are used.
- Which Services and ports receive the traffic.
- Which paths and HTTP methods are allowed.
- Which access strategy is used (**noAuth**, **jwt**, or **extAuth**).

The API Gateway module watches APIRules and translates them into Istio configuration, such as `VirtualService` and `AuthorizationPolicy` objects. This means you do not need to create these Istio objects yourself.

For a detailed field description, see [APIRule Custom Resource](../custom-resources/apirule/04-10-apirule-custom-resource.md).

## What Can You Do with APIRule?

With an APIRule, you can:

- **Choose an access strategy per path**:
    - **jwt** – Protect an endpoint with JSON Web Tokens. See [JWT Validation](./jwt/README.md).
    - **extAuth** – Delegate authentication and authorization to an external provider. See [External Authorization](./extAuth/README.md).
	- **noAuth** – Expose an endpoint without authentication. See [noAuth Access Strategy](./noAuth/README.md).
   
    > [!NOTE] For most common scenarios, it's recommended to use the jwt access strategy as it provides built‑in, token‑based authentication and authorization without requiring additional infrastructure.

    > The **extAuth** access strategy allows you to provide your custom authorization and authentication logic. Use this access strategy when the built-in Istio JWT authentication and authorization mechanisms don't meet your specific requirements, and you want to offload the logic to a custom external service.

    > Use **noAuth** for endpoints that must remain openly accessible. This setup is suitable for development and testing environments where security requirements are lower and quick access to services is necessary, or when the data being accessed isn't sensitive and doesn't require strict security measures.

- Use short host names instead of full domain names. The domain suffix is taken from the referenced Istio Gateway. See [Use a Short Host Name in APIRule](./expose-workload-short-host-name.md).
- Expose multiple workloads on the same host and route traffic to different Services based on paths and HTTP methods. See [Expose Multiple Workloads on the Same Host](./expose-multiple-workloads.md).
- Expose and secure workloads with mTLS by using an APIRule together with an mTLS-enabled Istio Gateway. See [Configure mTLS Authentication in SAP BTP, Kyma Runtime](../istio-gateways/configure-mtls-Gardener-certs.md).

You can combine these options in a single APIRule. For example, you can expose several Services under one host, protect some paths with **jwt** or **extAuth**, and leave public status and health check endpoints open with **noAuth**.

### Exposing Workloads with Istio

In most scenarios, using APIRules to expose and secure workloads is recommended. However, you can also configure access with Istio resources directly. For example, if you need to manage Istio configuration yourself or require a feature that APIRule does not yet support.

See the following topics:
- [Use the XFF Header to Configure IP-Based Access to a Workload](./ip-based-access-with-xff.md)
- [Expose a Workload with Istio VirtualService](./ip-based-access-with-xff.md)