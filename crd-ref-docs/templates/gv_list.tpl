{{- define "gvList" -}}
{{- $groupVersions := . -}}

# APIRule Custom Resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data used to configure an APIRule custom resource (CR). To get the up-to-date CRD in the yaml format, run the following command:

```bash
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## APIVersions
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}