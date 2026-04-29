# Using a Wildcard Host
Use a wildcard hostname in APIRule to route traffic from matching subdomains to a single service.

## What Is a Wildcard Host?

A wildcard hostname, such as `*.example.com`, matches any subdomain under a given domain and routes all matching traffic with a single APIRule.

This is especially useful for multitenant applications where each tenant is accessed through a unique URL following a predictable pattern, such as `<tenant-id>.example.com`. A wildcard hostname covers all matching subdomains without requiring a separate APIRule for each tenant.

## Requirements

- The referenced Gateway must serve a wildcard host that covers the domain you use in the APIRule.
- The `*` wildcard character is only allowed as the leftmost label of the host, for example, `*.example.com`. Patterns like `api.*.com` are
  not supported.
- Full regex host matching is not supported. If you need to match a specific pattern, use explicit hostnames in **spec.hosts** instead.

## Example

The following APIRule exposes a workload for any subdomain under `example.com`:

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin-wildcard
  namespace: default
spec:
  hosts:
    - "*.example.com”
  service:
    name: httpbin
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
```

Any request to a host matching `*.example.com` — for example `tenant1.example.com` or `tenant2.example.com` — is routed to the HTTPBin Service.

## Wildcard Host and Specific Host Conflict

> [!WARNING]
> If a wildcard APIRule (for example, `*.example.com`) uses `noAuth`, it also matches specific subdomains like `a.example.com`. In such a case, a separate APIRule for `a.example.com` with `jwt` does not enforce authentication, because the wildcard `noAuth` route already allows the request.

To avoid the issue, follow these rules:
- Do not combine a wildcard `noAuth` APIRule with specific-host APIRules that require authentication under the same domain.
- If you need different access strategies for specific subdomains, avoid using `noAuth` on the wildcard host. Instead, apply the least permissive strategy (for example, `jwt`) on the wildcard and create separate APIRules only for hosts that genuinely require `noAuth`.
