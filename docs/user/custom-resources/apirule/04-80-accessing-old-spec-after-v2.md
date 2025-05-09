# Retrieving APIRule v1beta1 after module upgrade

In future version of API Gateway module the v1beta1 version of APIRule will be removed.
That means the old configuration of APIRule that is not compatible with v2 specification will result in broken `Error` state,
However, due to compatibility reasons, the old configuration of APIRule will be available in the resource annotation as a reference to recover old information.

> ![NOTE] It is strongly recommended to migrate the old configuration of APIRule to the new v2 specification when the 
> v1beta1 version is still available, to avoid potential data loss.

## Retrieving old configuration

In v2alpha1 and v2 version of APIRules, the incompatible spec field is saved in an annotation named `gateway.kyma-project.io/v1beta1-spec`.
It contains the old configuration of APIRule in version v1beta1 in JSON format.

To retrieve the old configuration, get the affected APIRule JSON using the `kubectl get` command. The command below will create the `old-rule.json` file in the current directory with the contents of the annotation:
```bash
kubectl get apirule api-rule -o jsonpath='{.metadata.annotations.gateway\.kyma-project\.io/v1beta1-spec}' > old-rule.json
```

Open the file in a text editor to view the old configuration. You can also use the `jq` command to format the JSON output:
```bash
jq '.' old-rule.json
```

After that you can use parsed configuration and follow [the procedure](04-60-apirule-migration.md) to migrate APIRule v1beta1 to v2.