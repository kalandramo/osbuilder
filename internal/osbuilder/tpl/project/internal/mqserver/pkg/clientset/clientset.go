{{- $D := or .Web .MQ .Job .CLI -}}
package clientset

import (
	{{- range $D.Clients }}
    "{{$.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
    {{- end}}
	{{- if and .Job (not .Job.DisableCronJob) }}
    "{{$.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset/typed/train"
    "{{$.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset/typed/minio"
    {{- end}}
)

// Interface defines the operations for accessing different client types within the clientset.
type Interface interface {
	{{- if and .Job (not .Job.DisableCronJob) }}
	Minio() minio.Interface
	Train() train.Interface
    {{- end}}
	{{- range $D.Clients }}
	// {{. | kind}} returns the client interface for managing {{. | lowerkind}} resources.
	{{. | kind}}() {{. | lowerkind}}.Interface
    {{- end}}
}

// Clientset provides a unified entry point to access various typed clients.
type Clientset struct {
	{{- if and .Job (not .Job.DisableCronJob) }}
    minio minio.Interface
    train train.Interface
    {{- end}}
	{{- range $D.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface
    {{- end}}
}

// New creates a new Clientset with the provided client interfaces.
// It acts as a constructor for the Clientset type.
func New(
	{{- if and .Job (not .Job.DisableCronJob) }}
    minio minio.Interface,
    train train.Interface,
    {{- end}}
	{{- range $D.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface,
    {{- end}}
) *Clientset {
	return &Clientset{
		{{- if and .Job (not .Job.DisableCronJob) }}
        minio: minio,
        train: train,
    	{{- end}}
		{{- range $D.Clients }}
		{{. | lowerkind}}: {{. | lowerkind}},
    	{{- end}}
	}
}

{{- if and .Job (not .Job.DisableCronJob) }}
func (c *Clientset) Minio() minio.Interface {
    return c.minio
}

func (c *Clientset) Train() train.Interface {
    return c.train
}
{{- end}}

{{- range $D.Clients }}
// {{. | kind}} returns the client interface for {{. | lowerkind}} resources.
func (c *Clientset) {{. | kind}}() {{. | lowerkind}}.Interface {
	return c.{{. | lowerkind}}
}
{{- end}}
