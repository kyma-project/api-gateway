# FAQ: APIRule Migration <!-- omit in toc -->
APIRule v1beta1 has been deprecated and scheduled for deletion. See the frequently asked questions related to the migration process.

- [Why the `kubectl get` command returns my APIRule in verison `v2`?](#why-the-kubectl-get-command-returns-my-apirule-in-verison-v2)
- [If `kubectl get` returns an APIRule in version `v2`, does it mean that my APIRule is migrated to `v2`?](#if-kubectl-get-returns-an-apirule-in-version-v2-does-it-mean-that-my-apirule-is-migrated-to-v2)
- [How to check which version of APIRule I'm using?](#how-to-check-which-version-of-apirule-im-using)
- [How to request an APIRule in a particular version?](#how-to-request-an-apirule-in-a-particular-version)
- [Why my APIRule does not contain rules?](#why-my-apirule-does-not-contain-rules)
- [Why doesn't Kyma dashboard display all my APIRules?](#why-doesnt-kyma-dashboard-display-all-my-apirules)
- [Why do I get CORS policy errors after applying APIRule `v2`?](#why-do-i-get-cors-policy-errors-after-applying-apirule-v2)
- [I used **oauth2-introspection** in APIRule `v2`. How do I migrate to `v2`?](#i-used-oauth2-introspection-in-apirule-v2-how-do-i-migrate-to-v2)
- [I used the path `/.*` in APIRule `v1beta1`. How to migrate it to `v2`?](#i-used-the-path--in-apirule-v1beta1-how-to-migrate-it-to-v2)
  
## Why the `kubectl get` command returns my APIRule in verison `v2`?

APIRule `v2` is now the default version displayed by kubectl. This means that no metter which version is actually applied in the cluster, kubectl converts the APIRule's textual format so that it can be displayed using the `v2` specification.

## If `kubectl get` returns an APIRule in version `v2`, does it mean that my APIRule is migrated to `v2`?

No. APIRule `v2` is now the default version displayed by kubectl. Kubectl converts the textual format of each APIRule, no matter what its original version is. So, if your APIRule is in version `v1beta1`, kubectl converts it to version `v2` and displays it in the command's output. This conversion does not affect the resource itself.

To verify if your APIRule is migrated, check the annotation `gateway.kyma-project.io/original-version`. If it specifies version `v2`, your APIRule is migrated. If the annotation is `gateway.kyma-project.io/original-version: v1beta1`, this means that the resource is in version `v1beta1` even though in the command line it is converted to match the `v2` specification. 

>**NOTE:** Do not manually change the `gateway.kyma-project.io/original-version` annotation. This annotation is automatically updated when you apply your APIRule in version `v2`.

## How to check which version of APIRule I'm using?

To check the version of your APIRule, run the following command: 

```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```
The field `gateway.kyma-project.io/original-version` specifies the version of your APIRule.

## How to request an APIRule in a particular version?

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
Version `v2` is a stored version, so kubect uses it by default to display your APIRules no matter if you specify version `v2` in the command or do not specify any verison.

## Why my APIRule does not contain rules?

This APIRule is not migrated to version `v2`. Since version `v2` is now the default version, when you request an APIRule, kubectl converts it to verison `v2`. This conversion only affects the displayed resource’s textual format and does not modify the resource in the cluster. If the full conversion is possible, the rules field is presented in the output. However, if the conversion cannot be completed, the rules are missing, and the original rules are stored in the resource’s annotations. For more information, see [Retrieving the Complete **spec** of an APIRule in Version `v1beta1`
](./01-81-retrieve-v1beta1-spec.md)

## Why doesn't Kyma dashboard display all my APIRules?

APIRule deletion is divided into three steps. As part of the first step, APIRule `v1beta1` support has been removed from Kyma dashboard. This means that you can no longer view, edit, or create APIRules `v1beta1` using Kyma dashbaord. For more information on the deletion timeline and the next steps phases, see [APIRule Migration](./README.md#apirule-v1beta1-migration-timeline).

## Why do I get CORS policy errors after applying APIRule `v2`?

APIRule `v1beta1` applied default CORS configuration. APIRUle `v2` does not apply any default values, which means that by default it is only allowed to request resources from the same origin from which the application is loaded. If you want to use less restrictive CORS policy in APIRule `v2`, you must define it in the **spec.corsPolicy** field. For more information, see [Changes Introduced in APIRule v2](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/changes-introduced-in-apirule-v2?locale=en-US&state=DRAFT&version=Internal#cors-policy-is-not-applied-by-default).

## I used **oauth2-introspection** in APIRule `v2`. How do I migrate to `v2`?

The **oauth2-introspection** handler is removed from APIRule `v2`. To migrate your APIRule that uses this handler, you must first deploy a service that acts as an external authorizer for Istio and then define the **extAuth** access strategy in your APIRule CR. See [Migrating APIRule `v1beta1` of type oauth2_introspection to version `v2`](./01-84-migrate-oauth2-v1beta1-to-v2.md)


## I used the path `/.*` in APIRule `v1beta1`. How to migrate it to `v2`?

APIRule `v2` does not support regexp in the **pec.rules.path** field of APIRule CR. Instead, it supports the use of the `{*}` and `{**}` operators. So, if you want to use the wildard path in APIRule `v2`, you must replace `/.*` with `/*`. For more information, see [Changes Introduced in APIRule v2](../custom-resources/apirule/04-70-changes-in-apirule-v2.md)