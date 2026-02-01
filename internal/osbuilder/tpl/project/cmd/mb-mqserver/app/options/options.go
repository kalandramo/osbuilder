// nolint: err113
package options

import (
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"{{.M.ModuleName}}/internal/{{.MQ.Name}}"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
    // ObservabilityOptions contains the observability configuration options.
    ObservabilityOptions *genericoptions.ObservabilityOptions `json:"obs" mapstructure:"obs"`
	{{- if eq .MQ.MQFramework "kafka"}}
    // KafkaOptions specifies whether to create a kafka client.
    KafkaOptions *genericoptions.KafkaOptions `json:"kafka" mapstructure:"kafka"`
	{{- end}}
	{{- if eq .MQ.StorageType "mariadb"}}
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	{{- end}}
	{{- if eq .MQ.StorageType "postgresql"}}
	// PostgreSQLOptions contains the PostgreSQL configuration options.
	PostgreSQLOptions *genericoptions.PostgreSQLOptions `json:"postgresql" mapstructure:"postgresql"`
	{{- end}}
	{{- if eq .MQ.StorageType "sqlite"}}
	// SQLiteOptions contains the SQLite configuration options.
	SQLiteOptions *genericoptions.SQLiteOptions `json:"sqlite" mapstructure:"sqlite"`
	{{- end}}
	{{- if .MQ.WithOTel}}
    // OTelOptions used to specify the otel options.
    OTelOptions *genericoptions.OTelOptions `json:"otel" mapstructure:"otel"`
	{{- else}}
    // SlogOptions used to specify the slog options.
    SlogOptions *genericoptions.SlogOptions `json:"slog" mapstructure:"slog"`
	{{- end }}
	{{- range .MQ.Clients }}
    // {{. | kind}}Options specifies whether to create a {{. | lowerkind}} client.
    {{. | kind}}Options *genericoptions.RestyOptions `json:"{{. | lowerkind}}" mapstructure:"{{. | lowerkind}}"`
	{{- end }}
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		ObservabilityOptions: genericoptions.NewObservabilityOptions(),
		{{- if eq .MQ.MQFramework "kafka"}}
		KafkaOptions:         genericoptions.NewKafkaOptions(),
		{{- end}}
		{{- if eq .MQ.StorageType "mariadb"}}
		MySQLOptions:      genericoptions.NewMySQLOptions(),
		{{- end}}
		{{- if eq .MQ.StorageType "postgresql"}}
		PostgreSQLOptions:      genericoptions.NewPostgreSQLOptions(),
		{{- end}}
		{{- if eq .MQ.StorageType "sqlite"}}
		SQLiteOptions:      genericoptions.NewSQLiteOptions(),
		{{- end}}
		{{- if .MQ.WithOTel }}
		OTelOptions: genericoptions.NewOTelOptions(),
		{{- else}}
		SlogOptions: genericoptions.NewSlogOptions(),
		{{- end}}
		{{- range .MQ.Clients }}
		{{. | kind}}Options: genericoptions.NewRestyOptions(),
		{{- end}}
	}

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// Add command-line flags for sub-options.
	o.ObservabilityOptions.AddFlags(fs, "obs")
	{{- if eq .MQ.MQFramework "kafka"}}
	o.KafkaOptions.AddFlags(fs, "kafka")
	{{- end}}
	{{- if eq .MQ.StorageType "mariadb"}}
	o.MySQLOptions.AddFlags(fs, "mysql")
	{{- end}}
	{{- if eq .MQ.StorageType "postgresql"}}
	o.PostgreSQLOptions.AddFlags(fs, "postgresql")
	{{- end}}
	{{- if eq .MQ.StorageType "sqlite"}}
	o.SQLiteOptions.AddFlags(fs, "sqlite")
	{{- end}}
    {{- if .MQ.WithOTel}}
	o.OTelOptions.AddFlags(fs, "otel")
    {{- else}}
	o.SlogOptions.AddFlags(fs, "slog")
    {{- end}}
	{{- range .MQ.Clients }}
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
	errs = append(errs, o.ObservabilityOptions.Validate()...)
	{{- if eq .MQ.MQFramework "kafka"}}
	errs = append(errs, o.KafkaOptions.Validate()...)
	{{- end}}
	{{- if eq .MQ.StorageType "mariadb"}}
	errs = append(errs, o.MySQLOptions.Validate()...)
	{{- end}}
	{{- if eq .MQ.StorageType "postgresql"}}
	errs = append(errs, o.PostgreSQLOptions.Validate()...)
	{{- end}}
	{{- if eq .MQ.StorageType "sqlite"}}
	errs = append(errs, o.SQLiteOptions.Validate()...)
	{{- end}}
	{{- if .MQ.WithOTel}}
	errs = append(errs, o.OTelOptions.Validate()...)
    {{- else}}
	errs = append(errs, o.SlogOptions.Validate()...)
    {{- end}}
	{{- range .MQ.Clients }}
	errs = append(errs, o.{{. | kind}}Options.Validate()...)
    {{- end}}

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an {{.MQ.Name}}.Config based on ServerOptions.
func (o *ServerOptions) Config() (*{{.MQ.Name}}.Config, error) {
	return &{{.MQ.Name}}.Config{
		ObservabilityOptions: o.ObservabilityOptions,
		{{- if eq .MQ.MQFramework "kafka"}}
		KafkaOptions:         o.KafkaOptions,
		{{- end}}
		{{- if eq .MQ.StorageType "mariadb"}}
		MySQLOptions:      o.MySQLOptions,
		{{- end}}
		{{- if eq .MQ.StorageType "postgresql"}}
		PostgreSQLOptions:      o.PostgreSQLOptions,
		{{- end}}
		{{- if eq .MQ.StorageType "sqlite"}}
		SQLiteOptions:      o.SQLiteOptions,
		{{- end}}
		{{- range .MQ.Clients }}
		{{. | kind}}Options: o.{{. | kind}}Options,
		{{- end}}
	}, nil
}
