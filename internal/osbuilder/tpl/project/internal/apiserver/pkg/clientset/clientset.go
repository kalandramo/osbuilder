{{- $D := or .Web .MQ .Job .CLI -}}
package clientset

import (
	"sync"

	{{- range $D.Clients }}
    "{{$.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
    {{- end}}
)

// Interface defines the operations for accessing different client types within the clientset.
type Interface interface {
	{{- range $D.Clients }}
	// {{. | kind}} returns the client interface for managing {{. | lowerkind}} resources.
	{{. | kind}}() {{. | lowerkind}}.Interface
    {{- end}}
}

// Clientset provides a unified entry point to access various typed clients.
type Clientset struct {
	{{- range $D.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface
    {{- end}}
}

var (
    // CS is the global singleton instance of the clientset interface.
    // It allows global access to the initialized clientset from legacy packages.
    CS *Clientset
    // once ensures that the clientset initialization logic executes exactly once.
    once sync.Once
)

// New creates a new Clientset with the provided client interfaces.
// It acts as a constructor for the Clientset type.
func New(
	{{- range $D.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface,
    {{- end}}
) *Clientset {
	 once.Do(func() {
        CS = &Clientset{
			{{- range $D.Clients }}
			{{. | lowerkind}}: {{. | lowerkind}},
			{{- end}}
        }
    })

    return CS
}

{{- range $D.Clients }}
// {{. | kind}} returns the client interface for {{. | lowerkind}} resources.
func (c *Clientset) {{. | kind}}() {{. | lowerkind}}.Interface {
	return c.{{. | lowerkind}}
}
{{- end}}
