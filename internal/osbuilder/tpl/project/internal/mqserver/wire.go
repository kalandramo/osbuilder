//go:build wireinject
// +build wireinject

package {{.MQ.Name}}

import (
    "context"

	"github.com/google/wire"

	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/biz"
	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/handler"
	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/pkg/validation"
	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/store"
    {{- if .MQ.Clients }}
    "{{.M.ModuleName}}/internal/{{.MQ.Name}}/pkg/clientset"
    {{- end}}
)

// infrastructureSet groups all infrastructure-related providers.
// This keeps the main wire.Build call clean.
var infrastructureSet = wire.NewSet(
    ProvideDB,
    {{- if .MQ.Clients }}
    {{- range .MQ.Clients }}
    Provide{{. | kind}}Client,
    {{- end}}
    clientset.New,
    wire.Bind(new(clientset.Interface), new(*clientset.Clientset)),
    {{- end}}
    {{- if .MQ.WithPreloader }}
    ProvideAStore,
    {{- end}}
)

// NewServer initializes and creates the web server with all necessary dependencies using Wire.
func NewServer(context.Context, *Config) (*Server, error) {
    wire.Build(
        // Server infrastructure
        NewMQServer,
        NewDependencies,
        wire.Struct(new(ServerConfig), "*"), // Inject all fields
        wire.Struct(new(Server), "*"),

        // Domain layers
        store.ProviderSet,
        biz.ProviderSet,
        validation.ProviderSet,
        handler.NewHandler,

        // Infrastructure dependencies
        infrastructureSet,
    )
    return nil, nil
}
