//go:build wireinject
// +build wireinject

package {{.Job.Name}}

import (
    "context"

	"github.com/google/wire"

    {{- if .Job.Clients }}
    "{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/clientset"
    {{- end}}
)

// infrastructureSet groups all infrastructure-related providers.
// This keeps the main wire.Build call clean.
var infrastructureSet = wire.NewSet(
    ProvideDB,
    {{- if .Job.Clients }}
    {{- range .Job.Clients }}
    Provide{{. | kind}}Client,
    {{- end}}
    clientset.New,
    wire.Bind(new(clientset.Interface), new(*clientset.Clientset)),
    {{- end}}
    {{- if .Job.WithPreloader }}
    ProvideAStore,
    {{- end}}
)

// NewServer initializes and creates the web server with all necessary dependencies using Wire.
func NewServer(context.Context, *Config) (*Server, error) {
    wire.Build(
        // Server infrastructure
        NewJobServer,
        NewDependencies,
        wire.Struct(new(ServerConfig), "*"), // Inject all fields
        wire.Struct(new(Server), "*"),

        // Infrastructure dependencies
        infrastructureSet,
    )
    return nil, nil
}
