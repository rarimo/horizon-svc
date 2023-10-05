{{- $t := .Data -}}
{{- if $t.Comment -}}
// {{ $t.Comment | eval $t.GoName }}
{{- else -}}
// {{ $t.GoName }} represents a row from '{{ schema $t.SQLName }}'.
{{- end }}
type {{ $t.GoName }} struct {
{{ range $t.Fields -}}
	{{ .GoName }} {{ .Type }} `db:"{{ .SQLName }}" json:"{{ .SQLName }}" {{ if .IsPrimary }}structs:"-"{{ else }}structs:"{{ .SQLName }}"{{ end }}` // {{ .SQLName }}
{{ end }}
}
