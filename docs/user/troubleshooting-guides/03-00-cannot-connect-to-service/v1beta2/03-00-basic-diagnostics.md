# Issues with APIRules and Service Connection

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). See [Issues When Creating an APIRule in version v1beta2](./03-40-api-rule-troubleshooting.md).

To diagnose problems, inspect the [`status` code](../../custom-resources/apirule/04-10-apirule-custom-resource.md) of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `ERROR`, edit the APIRule and fix the issues described in the **status.description** field.