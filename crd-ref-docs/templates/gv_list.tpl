{{- define "gvList" -}}
{{- $groupVersions := . -}}

{{- range $groupVersions }}
{{- if has "APIGateway" .Kinds }}

# APIGateway Custom Resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes 
the kind and the format of data that APIGateway Controller uses to configure the 
API Gateway resources. Applying the custom resource (CR) triggers the installation 
of API Gateway resources, and deleting it triggers the uninstallation of those resources. 
The default CR has the name `default`.

```bash
kubectl get crd apigateways.operator.kyma-project.io -o yaml
```

You are only allowed to have one APIGateway CR. If there are multiple APIGateway CRs 
in the cluster, the oldest one reconciles the module. Any additional APIGateway CR 
is placed in the `Warning` state.

## Sample Custom Resource
This is a sample APIGateway CR:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: APIGateway
metadata:
  labels:
    operator.kyma-project.io/managed-by: kyma
  name: default
spec:
  enableKymaGateway: true
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
{{- range $groupVersions }}
- `{{ .GroupVersionString }}`
{{- end -}}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- else if has "RateLimit" .Kinds }}

# RateLimit Custom Resource
The `ratelimits.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind 
and the format of data that RateLimit Controller uses to configure the request rate 
limit for applications.

To get the up-to-date CRD in the yaml format, run the following command:
```bash
kubectl get crd ratelimits.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource
This is a sample RateLimit custom resource (CR):

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-path-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 5
      tokensPerFill: 5
      fillInterval: 60s
    buckets:
      - path: /ip
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 60m
```

## Custom Resource Parameters
The following tables list all the possible parameters of a given resource together with their descriptions.

### APIVersions
{{- range $groupVersions }}
- `{{ .GroupVersionString }}`
{{- end -}}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- else if has "APIRule" .Kinds }}

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
- `{{ .GroupVersionString }}`
{{- end -}}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
{{- end -}}
{{- end -}}