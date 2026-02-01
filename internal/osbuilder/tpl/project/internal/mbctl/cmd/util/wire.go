//go:build wireinject
// +build wireinject

package util

import (
	"github.com/google/wire"

	{{- if .CLI.Clients }}
	"{{.M.ModuleName}}/internal/{{.CLI.Name}}/pkg/clientset"
	{{- end}}
	clioptions "{{.M.ModuleName}}/internal/{{.CLI.Name}}/util/options"
)

func NewFactory(*clioptions.CLIToolOptions) (*factoryImpl, error) {
	wire.Build(
		{{- if .CLI.Clients }}
		{{- range .CLI.Clients }}
    	Provide{{. | kind}}Client,
    	{{- end}}
    	clientset.New,
    	wire.Bind(new(clientset.Interface), new(*clientset.Clientset)),
		{{- end}}
		wire.Struct(new(factoryImpl), "*"),
	)
	return nil, nil
}
