# APIRule `v1beta1` not working/not diaplayed


TODO:
- ogolnie nannowych klastrach nie bedzie mozna dawac v1beta czyli scenario miales v1beta1 na starym zlaozyles nowy i nie dziala to przechodzisz na v2
- ze da sie tylko geta zrobic i zrobic migracje tej apiruli zgodnie z tutorialami migracyjnymi  i tutaj daz znac ze ta v1beta1 nadal dziala podspodem doi momentu kiedy sie ludzie zmigruja lub my skonczymy wsparcie
- musze o to podpytac czy robimy to z fetaure brancha to raczej krotkie zadanie bo Natali juz tam porobila duzo 
- zrobic klaster z ta configmpaa zeby zobaczyc jak to dziala
- podobne opisy będą pewnie w FAQ i jakimś troubleshooitng guidzie albo w tutorialach migracyjnych tak mi się zdaje ''
- czy pojawia sie warning w dashboardzie jesli jest v1beta1 ?
- czy to jest juz kolejny etap
Conversion Webhook Not Working Because of Certificate Issue

## Symptoms
- Kyma Dashoard does not display APIRules created in version `v1beta1`.
- When you try to create, modify or delete an `APIRule` created in `v1beta1` resource using `kubectl`, you encounter an error related to the conversion webhook certificate, for example:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
  ```bash
  Error from server: conversion webhook for gateway.kyma-project.io/v1beta1, Kind=APIRule failed: Post "https://api-gateway-webhook-service.kyma-system.svc:9443/convert?timeout=30s": x509: certificate has expired or is not yet valid
  ```
  

- There is an error related to the conversion webhook certificate when fetching or creating an `APIRule` resource, for example:
  ```bash
  Error from server: conversion webhook for gateway.kyma-project.io/v1beta1, Kind=APIRule failed: Post "https://api-gateway-webhook-service.kyma-system.svc:9443/convert?timeout=30s": x509: certificate has expired or is not yet valid


- There is an error in the `api-gateway-controller-manager` logs with the message `"Secret \"api-gateway-webhook-certificate\" not found"`, for example:    
  ```bash
  ERROR	Reconciler error	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"api-gateway-webhook-certificate","namespace":"kyma-system"}, "namespace": "kyma-system", "name": "api-gateway-webhook-certificate", "reconcileID": "a808b99f-6db6-47f5-a82e-8176811238ac", "error": "Secret \"api-gateway-webhook-certificate\" not found"}

## Cause

The APIRule was originally created using version `v1beta1`, and you haven’t yet migrated it to version `v2`. Since the latest stable version of the APIRule in the Kubernetes API is now `v2` and version `v1beta1` has been unserved and deleted and deprecated hence you cannot create it,modify it or delete it
APIRules in `v1beta1` won't display in Kyma Dashboard anymore. However, the APIRule in version `v1beta1` still exists in the cluster, and under the hood everything is wokring but you won't be able to manage it or change the configuration via Kyma Dashboard or via `kubectl` commands. To manage the APIRule, you must migrate it to version `v2`. For more information, see [APIRule Migration](../../apirule-migration/README.md).

When both versions of the APIRule CRD are present in the cluster


You have new cluster and cannot create an APIRule in v1beta1 anymore, only v2 is possible
even if on old clsuter you still have v1beta1 you should migrate to v2 as soon as possible

## Solution

Restart `api-gateway-controller-manager` to recreate the Secret:

```bash
kubectl rollout restart deployment -n kyma-system api-gateway-controller-manager
```
