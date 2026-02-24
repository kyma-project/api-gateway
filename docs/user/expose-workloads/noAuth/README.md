# Configure the **noAuth** Access Strategy
Use the **noAuth** access strategy to simply expose a workload with no authorization or authentication configurations.

The **noAuth** access strategy provides a simple configuration for exposing workloads. Use it when you simply need to expose your workloads and allow access to specific HTTP methods without any authentication or authorization checks. This setup is suitable for development and testing environments where security requirements are lower and quick access to services is necessary, or when the data being accessed is not sensitive and does not require strict security measures.

See the following sample configuration:
```yaml
...
rules:
  - path: /headers
    methods: ["GET"]
    noAuth: true
```