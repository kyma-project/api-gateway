{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}

### {{ $type.Name }}
{{- if $type.Doc }}

{{ $type.Doc }}
{{- end }}
{{ if $type.IsAlias }}
Underlying type: {{ markdownRenderTypeLink $type.UnderlyingType }}
{{ end -}}
{{ if $type.Validation }}
Validation:
{{- range $type.Validation }}
- {{ if hasPrefix "Pattern: " . }}{{ else }}{{ . | replace ": " ": `"}}`{{ end }}
{{- end }}
{{ end -}}
{{ if $type.References }}
Appears in:
{{- range $type.SortedReferences }}
- {{ markdownRenderTypeLink . }}
{{- end }}
{{ end -}}

{{- if $type.Members }}
| Field | Description | Validation |
| --- | --- | --- |
{{ if $type.GVK -}}
| **apiVersion** <br /> string | `{{ $type.GVK.Group }}/{{ $type.GVK.Version }}` | None |
| **kind** <br /> string | `{{ $type.GVK.Kind }}` | None |
{{ end -}}

{{ range $type.Members -}}
| **{{ .Name  }}** <br /> {{ markdownRenderType .Type }} | {{ template "type_members" . }} | {{ if .Validation }}{{ range .Validation -}} {{ $processed := markdownRenderFieldDoc . }}{{ if contains ": `" $processed }}{{ $processed }}{{ else if hasPrefix "Pattern: " $processed }}{{ $processed | replace "Pattern: " "Pattern: `" | printf "%s`" }}{{ else }}{{ $processed | replace ": " ": `" | printf "%s`" }}{{ end }} <br />{{ end }}{{ else }}None{{ end }} |
{{ end -}}

{{ end -}}

{{ if $type.EnumValues }} 
| Field | Description |
| --- | --- |
{{ range $type.EnumValues -}}
| **{{ .Name }}** | {{ markdownRenderFieldDoc .Doc }} |
{{ end -}}
{{ end -}}


{{- end -}}
{{- end -}}