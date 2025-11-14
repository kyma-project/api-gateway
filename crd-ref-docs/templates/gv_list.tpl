{{- define "gvList" -}}
{{- $groupVersions := . -}}

# APIRule Custom Resource

The `istios.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that Istio Controller uses to configure,
update, and manage the Istio installation. Applying the CR triggers the installation of Istio,
and deleting it triggers the uninstallation of Istio. The default Istio CR has the name `default`.

To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```
You are only allowed to use one Istio CR, which you must create in the `kyma-system` namespace.
If the namespace contains multiple Istio CRs, the oldest one reconciles the module.
Any additional Istio CR is placed in the `Warning` state.

## APIVersions
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}