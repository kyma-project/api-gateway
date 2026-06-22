# FAQ <!-- omit in toc -->

APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed with version 3.4 of the API Gateway module. See the frequently asked questions related to the migration process.

- [Displaying an APIRule's Spec](#displaying-an-apirules-spec)
  - [Why doesn't my APIRule contain any rules?](#why-doesnt-my-apirule-contain-any-rules)
  - [Why doesn't my APIRule contain a Gateway?](#why-doesnt-my-apirule-contain-a-gateway)
  - [How do I request an APIRule in a particular version?](#how-do-i-request-an-apirule-in-a-particular-version)
  - [Why doesn't Kyma dashboard display all my APIRules?](#why-doesnt-kyma-dashboard-display-all-my-apirules)
- [Checking an APIRule's Version](#checking-an-apirules-version)
  - [Why does the `kubectl get` command return my APIRule in version `v2`?](#why-does-the-kubectl-get-command-return-my-apirule-in-version-v2)
  - [How do I check which version of APIRule I'm using?](#how-do-i-check-which-version-of-apirule-im-using)
- [Migrating an APIRule `v1beta1` to Version `v2`](#migrating-an-apirule-v1beta1-to-version-v2)
  - [By when must I migrate my APIRules `v1beta1`?](#by-when-must-i-migrate-my-apirules-v1beta1)
  - [How do I know which APIRules must be migrated?](#how-do-i-know-which-apirules-must-be-migrated)
  - [If `kubectl get` returns an APIRule in version `v2`, does it mean that my APIRule is migrated to `v2`?](#if-kubectl-get-returns-an-apirule-in-version-v2-does-it-mean-that-my-apirule-is-migrated-to-v2)
  - [Why do I get CORS policy errors after applying APIRule `v2`?](#why-do-i-get-cors-policy-errors-after-applying-apirule-v2)
  - [I used **oauth2-introspection** in APIRule `v1beta1`. How do I migrate it to `v2`?](#i-used-oauth2-introspection-in-apirule-v1beta1-how-do-i-migrate-it-to-v2)
  - [I used regexp in the paths of APIRule `v1beta1`. How do I migrate it to `v2`?](#i-used-regexp-in-the-paths-of-apirule-v1beta1-how-do-i-migrate-it-to-v2)
  - [Why do I get a validation error for the legacy gateway format while trying to migrate to `v2`?](#why-do-i-get-a-validation-error-for-the-legacy-gateway-format-while-trying-to-migrate-to-v2)
  - [How to migrate multiple APIRules `v1beta1` targeting same workload to version `v2`?](#how-to-migrate-multiple-apirules-v1beta1-targeting-same-workload-to-version-v2)
- [Using APIRules `v1beta1`](#using-apirules-v1beta1)
  - [Why are my APIRules `v1beta1` in the `Warning` state?](#why-are-my-apirules-v1beta1-in-the-warning-state)
  - [When will APIRules `v1beta1` stop being reconciled?](#when-will-apirules-v1beta1-stop-being-reconciled)
  - [Why can't I create APIRules `v1beta1`?](#why-cant-i-create-apirules-v1beta1)


## Displaying an APIRule's Spec

### Why doesn't my APIRule contain any rules?

This APIRule is not migrated to version `v2`. Since version `v2` is now the default, when you request an APIRule, kubectl converts it to version `v2`. This conversion only affects the displayed resource’s textual format and does not modify the resource in the cluster. If the full conversion is possible, the **rules** field is presented in the output. However, if the conversion cannot be completed, the rules are missing, and the original rules are stored in the resource’s annotation `gateway.kyma-project.io/v1beta1-spec`. For more information, see [Retrieving the Complete **spec** of an APIRule in Version `v1beta1`
](./01-81-retrieve-v1beta1-spec.md).

### Why doesn't my APIRule contain a Gateway?

If your APIRule doesn't contain the Gateway when displayed using kubectl, this means that your APIRule is in version `v1beta1` and uses an unsupported Gateway format. The APIRule `v2` supports only the Gateway format `namespace/gateway-name`. When you try to display the APIRule `v1beta1` using kubectl, its textual format is converted to version `v2`. Since the Gateway format you're using is neither available in version `v2` nor `v1beta1`, it is not included in the output.

### How do I request an APIRule in a particular version?

Specify the version you want to request in the kubectl command. 

To get version `v1beta1`, run: 
```bash
kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```

To get version `v2alpha1`, run: 
```bash
kubectl get apirules.v2alpha1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```

To get version `v2`, run: 
```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```
Version `v2` is the stored version, so kubectl uses it by default to display your APIRules, no matter if you specify version `v2` in the command or do not specify any version.

### Why doesn't Kyma dashboard display all my APIRules?

 As part of APIRule `v1beta1` deletion, APIRule `v1beta1` support has been removed from Kyma dashboard. To display or migrate APIRules `v1beta1`, use kubectl.

## Checking an APIRule's Version
  
### Why does the `kubectl get` command return my APIRule in version `v2`?

APIRule `v2` is now the default version displayed by kubectl. This means that no matter in which version the APIRule was actually created in the cluster, kubectl converts the APIRule's displayed textual format to the latest stable version `v2`. It does not modify the resource in the cluster.

### How do I check which version of APIRule I'm using?

To check the version of your APIRule, run the following command: 

```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```
The annotation `gateway.kyma-project.io/original-version` specifies the version of your APIRule.

## Migrating an APIRule `v1beta1` to Version `v2`

### By when must I migrate my APIRules `v1beta1`?

Migrate your `v1beta1` resources to version `v2` before release 3.9, which disables migration and reconciliation of these resources. It is scheduled for 19 August 2026 (fast channel) and 2 September 2026 (regular channel).

Once reconciliation is disabled, APIRules `v1beta1` will continue to function as currently configured, but the API Gateway module will no longer own or manage them. Changes you make will not be reverted, and unmanaged resources may cause disruptions in workload availability and access.

### How do I know which APIRules must be migrated?
You must migrate all APIRules `v1beta1` to version `v2`. To list all your APIRules `v1beta1`, run the following command:
```bash
kubectl get apirules.gateway.kyma-project.io -A -o json | jq '.items[] | select(.metadata.annotations["gateway.kyma-project.io/original-version"] == "v1beta1") | {namespace: .metadata.namespace, name: .metadata.name}'
```

### If `kubectl get` returns an APIRule in version `v2`, does it mean that my APIRule is migrated to `v2`?

No. APIRule `v2` is now the default version displayed by kubectl. Kubectl converts the textual format of each APIRule, no matter what its original version is. So, if your APIRule is in version `v1beta1`, kubectl converts it to version `v2` and displays it in the command's output. This conversion does not affect the resource itself.

To verify if your APIRule is migrated, check the annotation `gateway.kyma-project.io/original-version`. If it specifies version `v2`, your APIRule is migrated. If the annotation is `gateway.kyma-project.io/original-version: v1beta1`, this means that the resource is in version `v1beta1` even though in the command line it is converted to match the `v2` specification. 

>**NOTE:** Do not manually change the `gateway.kyma-project.io/original-version` annotation. This annotation is automatically updated when you migrate your APIRule to version `v2`.

### Why do I get CORS policy errors after applying APIRule `v2`?

APIRule `v1beta1` applied the default CORS configuration. APIRule `v2` does not apply any default values, which means that by default, it is only allowed to request resources from the same origin from which the application is loaded. If you want to use a less restrictive CORS policy in APIRule `v2`, you must define it in the **spec.corsPolicy** field. For more information, see [Changes Introduced in APIRule `v2`](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/changes-introduced-in-apirule-v2?locale=en-US&state=DRAFT&version=Internal#cors-policy-is-not-applied-by-default).

### I used **oauth2-introspection** in APIRule `v1beta1`. How do I migrate it to `v2`?

The **oauth2-introspection** handler is removed from APIRule `v2`. To migrate your APIRule that uses this handler, you must first deploy a service that acts as an external authorizer for Istio and then define the **extAuth** access strategy in your APIRule CR. See [Migrating APIRule `v1beta1` of type **oauth2_introspection** to version `v2`](./01-84-migrate-oauth2-v1beta1-to-v2.md).

### I used regexp in the paths of APIRule `v1beta1`. How do I migrate it to `v2`?

APIRule `v2` does not support regexp in the **spec.rules.path** field of APIRule CR. Instead, it supports using the `{*}` and `{**}` operators and the `/*` wildcard. For more information, see [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md).

### Why do I get a validation error for the legacy gateway format while trying to migrate to `v2`?

In APIRule `v2`, you must provide the Gateway using the format `namespace/gateway-name`. The legacy formats are not supported.

### How to migrate multiple APIRules `v1beta1` targeting same workload to version `v2`?
When multiple APIRules `v1beta1` target the same workload using different host names, you must apply an additional, temporary AuthorizationPolicy in order to have continuous access to endpoints exposed by APIRules `v1beta1` during the migration. For more information, see [Migrate Multiple APIRules Targeting the Same Workload from `v1beta1` to `v2`](./01-90-migrate-multiple-apirules-targeting-same-workload.md) and [Migration guidelines](./README.md).

## Using APIRules `v1beta1`

### Why are my APIRules `v1beta1` in the `Warning` state?
When a resource is in the `Warning` state, it signifies that user action is required. All APIRules `v1beta1` are set to this state to indicate that you must migrate these resources to version `v2`.

### When will APIRules `v1beta1` stop being reconciled?

Release 3.9 disables migration and reconciliation of APIRules `v1beta1`. For the complete timeline for SAP BTP, Kyma runtime, follow [API Gateway what's new notes](https://help.sap.com/whats-new/cf0cb2cb149647329b5d02aa96303f56?locale=en-US&version=Cloud&q=API+Gateway+module:).

Once reconciliation is disabled, APIRules `v1beta1` will continue to function as currently configured, but the API Gateway module will no longer own or manage them. Changes you make will not be reverted, and unmanaged resources may cause disruptions in workload availability and access.

### Why can't I create APIRules `v1beta1`?

As announced in [what's new notes](https://help.sap.com/whats-new/cf0cb2cb149647329b5d02aa96303f56?locale=en-US&version=Cloud&q=API+Gateway+module:+APIRule+v1beta1+deletion), it's no longer possible to create APIRule CRs `v1beta1` in new and existing clusters, modify existing APIRule CRs `v1beta1`, or delete them. All APIRule `v1beta1` configurations set up prior to this restriction remain active and continue to function as expected. Reconciliation of these resources will be disabled with release 3.9. For the complete timeline for SAP BTP, Kyma runtime, follow [API Gateway what's new notes](https://help.sap.com/whats-new/cf0cb2cb149647329b5d02aa96303f56?locale=en-US&version=Cloud&q=API+Gateway+module:).