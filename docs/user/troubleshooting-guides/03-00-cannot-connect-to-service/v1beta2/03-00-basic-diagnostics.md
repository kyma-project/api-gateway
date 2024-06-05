# Issues with APIRules and Service Connection

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). To diagnose problems, inspect the [`status` code](../../custom-resources/apirule/04-10-apirule-custom-resource.md) of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `Error`, edit the APIRule and fix the issues described in the **status.description** field. See the troubleshooting guides:
- [Issues When Creating an APIRule in version v1beta2](./03-40-api-rule-troubleshooting.md)
- [401 Unauthorized or 403 Forbidden](./03-01-401-unauthorized-403-forbidden.md)