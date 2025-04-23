# Title

# Context
Version 3.0.0 introduces APIRule in version v2. All APIRules v2 that you have in your cluster will be stored as version v2alpha1.

 In future versions of the API Gateway module, version v1beta1 of APIRule CRs will be removed.


This means that when you run the command 'kubectl get apirules -A -o yaml`, APIRule v2 will be displayed by default.


You have APIRule CRs in version v1beta1 in your cluster, which are in the error state. You must migrate your APIRules to version v2.

# v1beta1 -> v2 Migration Procedure

## Retrieve the v1beta1 specification that is not compatible with APIRule v2.

The incompatible spec field is saved in APIRule v2 in the `gateway.kyma-project.io/v1beta1-spec` annotation in the JSON format. To retrieve it, run the following command:

```bash
kubectl get apirule api-rule -o jsonpath='{.metadata.annotations.gateway\.kyma-project\.io/v1beta1-spec}' > old-rule.json
```

The command creates the `old-rule.json` file in your working directory with the contents of the annotation. See an example:
…

## Update the configuration of your APIRule CR.
Open the created file and analyze the v1beta1 configuration. You must manually update it to reflect the configuration in APIRule and consider all the changes introduce in the new version. Follow the checklist to make sure you considered all the aspects.
	
1. Update the handler configuration. 
   1. If you use the **jwt** handler in version v1beta1, migrate it to the jwt handler in version v2. See an example:
        <table>
        <tr>
        <th>v1beta1</th>
        <th>v2</th>
        </tr>
        <tr>
        <td>
        
        ```yaml
        v1beta1 config example
        ```

        </td>
        <td>

        ```yaml
        v2 config example
        ```
        </td>
        </tr>
        </table>

    2. If you use the no_auth, noop, or allow handlers in version v1beta1, migrate it to the noAuth handler in version v2. See an example:
        <table>
        <tr>
        <th>v1beta1</th>
        <th>v2</th>
        </tr>
        <tr>
        <td>
        
        ```yaml
        v1beta1 config example
        ```

        </td>
        <td>

        ```yaml
        v2 config example
        ```
        </td>
        </tr>
        </table>
    For more information, see the blog post.
    3. If you use the Ory Oathkeeper-based OAuth2 handlers or any other handler that is not supported in APIRule v2, migrate to the extAuth handler. See an example:
        <table>
        <tr>
        <th>v1beta1</th>
        <th>v2</th>
        </tr>
        <tr>
        <td>
        
        ```yaml
        v1beta1 config example
        ```

        </td>
        <td>

        ```yaml
        v2 config example
        ```
        </td>
        </tr>
        </table>
    For more information, see the blog post

2. In APIRule v2, the corsPolicy is not applied by default if the corsPolicy field is empty. If you rely on CORS configuration in APIRule v1beta1, you must apply in in APIRule v2. See an example:
 …
For more information, see CORS Policy Is Not Applied by Default.

3.	To retain APIRule v1beta1 internal traffic policy, apply the following AuthorizationPolicy. Remember to change the selector label to the one pointing to the target workload:
    ```bash
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
    name: allow-internal
    namespace: ${NAMESPACE}
    spec:
    selector:
        matchLabels:
        ${KEY}: ${TARGET_WORKLOAD}
    action: ALLOW
    rules:
    - from:
        - source:
            notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
    ```
4.	See the list of changes introduced in APIRule v2 to make sure that you don’t use any other configuration that is not supported anymore.
5.	Apply the updated configuration of your APIRule CR using kubectl.

## Result
As an immediate result of applying the updated APIRUle, a new Istio Authorization Policy and Istio Authentication Policy resources are created. The Istio VirtualService resource is updated to point directly to the target Service, bypassing Ory Oathkeeper.
Finally, the Ory Oathkeeper resource is deleted.


---

# v2alpha1 -> v2 Migration Procedure

You have APIRule CRs in version v2alpha1 in your cluster. You must migrate your APIRules to version v2.

1.	Edit the file with your APIRule CR.
2.	In the metadata section update the version to v2.
3.	Apply the configuration.
