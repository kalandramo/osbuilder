package options

import (
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/watch"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"{{.M.ModuleName}}/internal/{{.Job.Name}}"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
	// WatchOptions contains the Watch configuration options.
	WatchOptions *watch.Options `json:"watch" mapstructure:"watch"`
	{{- if eq .Job.StorageType "mariadb"}}
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"coredb" mapstructure:"coredb"`
	{{- end}}
	{{- if eq .Job.StorageType "postgresql"}}
	// PostgreSQLOptions contains the PostgreSQL configuration options.
	PostgreSQLOptions *genericoptions.PostgreSQLOptions `json:"postgresql" mapstructure:"postgresql"`
	{{- end}}
	{{- if eq .Job.StorageType "sqlite"}}
	// SQLiteOptions contains the SQLite configuration options.
	SQLiteOptions *genericoptions.SQLiteOptions `json:"sqlite" mapstructure:"sqlite"`
	{{- end}}
	{{- if .Job.WithOTel}}
    // OTelOptions used to specify the otel options.
    OTelOptions *genericoptions.OTelOptions `json:"otel" mapstructure:"otel"`
	{{- else}}
    // SlogOptions used to specify the slog options.
    SlogOptions *genericoptions.SlogOptions `json:"slog" mapstructure:"slog"`
	{{- end }}
	{{- range .Job.Clients }}
    // {{. | kind}}Options specifies whether to create a {{. | lowerkind}} client.
    {{. | kind}}Options *genericoptions.RestyOptions `json:"{{. | lowerkind}}" mapstructure:"{{. | lowerkind}}"`
	{{- end }}
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		WatchOptions:  watch.NewOptions(),
		{{- if eq .Job.StorageType "mariadb"}}
		MySQLOptions:      genericoptions.NewMySQLOptions(),
		{{- end}}
		{{- if eq .Job.StorageType "postgresql"}}
		PostgreSQLOptions:      genericoptions.NewPostgreSQLOptions(),
		{{- end}}
		{{- if eq .Job.StorageType "sqlite"}}
		SQLiteOptions:      genericoptions.NewSQLiteOptions(),
		{{- end}}
		{{- if .Job.WithOTel }}
		OTelOptions: genericoptions.NewOTelOptions(),
		{{- else}}
		SlogOptions: genericoptions.NewSlogOptions(),
		{{- end}}
		{{- range .Job.Clients }}
		{{. | kind}}Options: genericoptions.NewRestyOptions(),
		{{- end}}
	}

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// Add command-line flags for sub-options.
	o.WatchOptions.AddFlags(fs, "watch")
	{{- if eq .Job.StorageType "mariadb"}}
	o.MySQLOptions.AddFlags(fs, "coredb")
	{{- end}}
	{{- if eq .Job.StorageType "postgresql"}}
	o.PostgreSQLOptions.AddFlags(fs, "postgresql")
	{{- end}}
	{{- if eq .Job.StorageType "sqlite"}}
	o.SQLiteOptions.AddFlags(fs, "sqlite")
	{{- end}}
    {{- if .Job.WithOTel}}
	o.OTelOptions.AddFlags(fs, "otel")
    {{- else}}
	o.SlogOptions.AddFlags(fs, "slog")
    {{- end}}
	{{- range .Job.Clients }}
	o.{{. | kind}}Options.AddFlags(fs, "{{. | lowerkind}}")
    {{- end}}
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// TODO: Add the completion logic if needed.
    return nil
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}

	// Validate sub-options.
	errs = append(errs, o.WatchOptions.Validate()...)
	{{- if eq .Job.StorageType "mariadb"}}
	errs = append(errs, o.MySQLOptions.Validate()...)
	{{- end}}
	{{- if eq .Job.StorageType "postgresql"}}
	errs = append(errs, o.PostgreSQLOptions.Validate()...)
	{{- end}}
	{{- if eq .Job.StorageType "sqlite"}}
	errs = append(errs, o.SQLiteOptions.Validate()...)
	{{- end}}
	{{- if .Job.WithOTel}}
	errs = append(errs, o.OTelOptions.Validate()...)
    {{- else}}
	errs = append(errs, o.SlogOptions.Validate()...)
    {{- end}}
	{{- range .Job.Clients }}
	errs = append(errs, o.{{. | kind}}Options.Validate()...)
    {{- end}}

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an {{.Job.Name}}.Config based on ServerOptions.
func (o *ServerOptions) Config() (*{{.Job.Name}}.Config, error) {
	return &{{.Job.Name}}.Config{
		WatchOptions:  o.WatchOptions,
		{{- if eq .Job.StorageType "mariadb"}}
		MySQLOptions:      o.MySQLOptions,
		{{- end}}
		{{- if eq .Job.StorageType "postgresql"}}
		PostgreSQLOptions:      o.PostgreSQLOptions,
		{{- end}}
		{{- if eq .Job.StorageType "sqlite"}}
		SQLiteOptions:      o.SQLiteOptions,
		{{- end}}
		{{- range .Job.Clients }}
		{{. | kind}}Options: o.{{. | kind}}Options,
		{{- end}}
	}, nil
}
