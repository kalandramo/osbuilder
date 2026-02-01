package options

import (
	{{- if .CLI.Clients }}
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	{{- end}}
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// Defines flag for {{.CLI.Name}}.
const (                                               
    FlagConfig = "config"                                                                                                          
)

// CLIToolOptions contains the configuration options for the server.
type CLIToolOptions struct {
	{{- range .CLI.Clients }}
    // {{. | kind}}Options specifies whether to create a {{. | lowerkind}} client.
    {{. | kind}}Options *genericoptions.RestyOptions `json:"{{. | lowerkind}}" mapstructure:"{{. | lowerkind}}"`
	{{- end }}
}

// NewCLIToolOptions creates a CLIToolOptions instance with default values.
func NewCLIToolOptions() *CLIToolOptions {
	opts := &CLIToolOptions{
		{{- range .CLI.Clients }}
		{{. | kind}}Options: genericoptions.NewRestyOptions(),
		{{- end}}
	}

	return opts
}

// AddFlags binds the options in CLIToolOptions to command-line flags.
func (o *CLIToolOptions) AddFlags(fs *pflag.FlagSet) {
	{{- range .CLI.Clients }}
	o.{{. | kind}}Options.AddFlags(fs, "{{. | lowerkind}}")
    {{- end}}
}

// Complete completes all the required options.
func (o *CLIToolOptions) Complete() error {
	// TODO: Add the completion logic if needed.
    return nil
}

// Validate checks whether the options in CLIToolOptions are valid.
func (o *CLIToolOptions) Validate() error {
	errs := []error{}

	{{- range .CLI.Clients }}
	errs = append(errs, o.{{. | kind}}Options.Validate()...)
    {{- end}}

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}
