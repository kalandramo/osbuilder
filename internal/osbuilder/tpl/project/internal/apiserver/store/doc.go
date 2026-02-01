{{- $D := or .Web .MQ .Job -}}
{{- if .Web }}
// Package store defines the persistence layer interfaces and implementations for the apiserver.
{{- else if .Job }}
// Package store defines the persistence layer interfaces and implementations for the jobserver.
{{- end }}
package store
