{{- define "type_members" -}}
{{- $field := . -}}
{{- if eq $field.Name "metadata" -}}
For more information on the metadata fields, see Kubernetes API documentation.
{{- else -}}
{{ markdownRenderFieldDoc $field.Doc }}
{{- end -}}
{{- end -}}