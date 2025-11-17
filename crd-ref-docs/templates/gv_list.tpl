{{- define "gvList" -}}
{{- $groupVersions := . -}}

# APIRule Custom Resource

The apirules.gateway.kyma-project.io CustomResourceDefinition (CRD) describes the kind and the format of data the APIGateway Controller listens for. To get the up-to-date CRD in the yaml format, run the following command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```

## APIVersions
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}