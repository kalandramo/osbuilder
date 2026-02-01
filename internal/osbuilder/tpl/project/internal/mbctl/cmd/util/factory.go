package util

import (

	{{- if .CLI.Clients }}
	"{{.M.ModuleName}}/internal/{{.CLI.Name}}/pkg/clientset"
    {{- end}}
    {{- range .CLI.Clients }}
    "{{$.M.ModuleName}}/internal/{{$.CLI.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
    {{- end}}
	clioptions "{{.M.ModuleName}}/internal/{{.CLI.Name}}/util/options"
)

// Factory defines the contract for accessing global CLI options and the unified clientset.
// It serves as the central access point for infrastructure components.
type Factory interface {
	Options() *clioptions.CLIToolOptions
	{{- if .CLI.Clients }}
	Clientset() clientset.Interface
    {{- end}}
}

type factoryImpl struct {
	opts *clioptions.CLIToolOptions
	{{- if .CLI.Clients }}
	cs   clientset.Interface
    {{- end}}
}

// Ensure factoryImpl implements the Factory interface at compile time.
var _ Factory = (*factoryImpl)(nil)

func (f *factoryImpl) Options() *clioptions.CLIToolOptions {
	return f.opts
}

{{- if .CLI.Clients }}
func (f *factoryImpl) Clientset() clientset.Interface {
	return f.cs
}
{{- end}}

{{- range .CLI.Clients }}                 
// Provide{{. | kind}}Client creates and returns a {{. | lowerkind}} client instance using the provided configuration.
func Provide{{. | kind}}Client(opts *clioptions.CLIToolOptions) {{. | lowerkind}}.Interface {
    return {{. | lowerkind}}.NewForConfig(opts.{{. | kind}}Options)
}                                              
{{- end}}
