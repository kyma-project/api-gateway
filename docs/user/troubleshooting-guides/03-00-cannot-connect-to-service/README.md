# Cannot connect to a Service exposed by an APIRule - basic diagnostics

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). To diagnose problems, inspect the [`status` code](../../custom-resources/apirule/04-10-apirule-custom-resource.md) of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `Error`, edit the APIRule and fix the issues described in the **.Status.APIRuleStatus.desc** field. If you still encounter issues, make sure that API Gateway, and Oathkeeper are running, or take a look at one of the more specific troubleshooting guides:

- [Cannot connect to a Service exposed by an APIRule - `404 Not Found`](./03-02-404-not-found.md)
- [Cannot connect to a Service exposed by an APIRule - `401 Unathorized or 403 Forbidden`](./03-01-401-unauthorized-403-forbidden.md)
- [Cannot connect to a Service exposed by an APIRule - `500 Internal Server Error`](./03-03-500-server-error.md)