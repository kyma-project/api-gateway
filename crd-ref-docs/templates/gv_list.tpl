{{- define "gvList" -}}
{{- $groupVersions := . -}}

# APIRule Custom Resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data used to configure an APIRule custom resource (CR). To get the up-to-date CRD in the yaml format, run the following command:

```bash
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource
This is a sample APIRule CR that exposes the `foo-service` on the host `foo.bar`.

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: service-exposed
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360
  rules:
    - path: /*
      methods: [ "GET" ]
      noAuth: true
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
{{- range $groupVersions }}
- {{ .GroupVersionString }}
{{- end -}}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}