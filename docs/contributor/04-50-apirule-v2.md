# APIRule v2

## Current state of APIRules
APIRules CRD the newest version is `v2`, APIRule stored and served version is `v1beta1`.
APIRule `v2` is exact copy of `v2alpha1` only one difference between those two are version of CRD.

## Flow of APIRule reconciliation
Reconciliation are made by `APIRuleController` which is getting APIRule as `v1beta1`, during reconciliation controller checking original version base on `gateway.kyma-project.io/original-version` of APIRule and differentiate logic of reconciliation for `v2alpha1` and `v1beta1`.

CR `v2` version is supported by `v2alpha1` controller logic, and it is translated from `v1beta1` and annotations to `v2alpha1` and is a base for creation or update of resources.
