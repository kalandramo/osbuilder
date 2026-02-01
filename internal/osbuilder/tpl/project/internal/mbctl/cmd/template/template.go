package {{.CLI.R.SingularLower}}

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "{{.M.ModuleName}}/internal/{{.CLI.Name}}/cmd/util"
)

// {{.CLI.R.SingularName}}Options defines configuration for the "{{.CLI.R.SingularLower}}" command.
type {{.CLI.R.SingularName}}Options struct {
	RootDir string // working directory
	Output  string // output directory
	Force   bool   // overwrite if exists

	genericiooptions.IOStreams
}

var (
	{{.CLI.R.SingularLower}}LongDesc = templates.LongDesc(`{{.CLI.R.SingularLower}} command provides functionality for...
		Add your detailed description here.`)

	{{.CLI.R.SingularLower}}Example = templates.Examples(`# Basic usage
	    {{.CLI.BinaryName}} {{.CLI.R.SingularLower}} [options]

		# With custom options
		{{.CLI.BinaryName}} {{.CLI.R.SingularLower}} --output ./custom/path`)
)

// New{{.CLI.R.SingularName}}Cmd creates the "{{.CLI.R.SingularLower}}" command.
func New{{.CLI.R.SingularName}}Cmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &{{.CLI.R.SingularName}}Options{
		IOStreams: ioStreams,
		Output:    "./",
	}

	cmd := &cobra.Command{
		Use:                   "{{.CLI.R.SingularLower}}",
	    Short:                 "TODO: Modify {{.CLI.R.SingularLower}} command description here.",
		Long:                  {{.CLI.R.SingularLower}}LongDesc,
		Example:               {{.CLI.R.SingularLower}}Example,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Add flags
	cmd.Flags().StringVar(&opts.Output, "output", opts.Output, "Output directory for generated files.")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files if they exist.")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *{{.CLI.R.SingularName}}Options) Complete() error {
	o.RootDir, _ = os.Getwd()
	return nil
}

// Validate ensures provided inputs are valid.
func (o *{{.CLI.R.SingularName}}Options) Validate() error {
	// Add your validation logic here
	return nil
}

// Run performs the {{.CLI.R.SingularLower}} operation.
func (o *{{.CLI.R.SingularName}}Options) Run() error {
	// TODO: Implement your command logic here
	fmt.Printf("Successfully executed {{.CLI.R.SingularLower}} command\n")
	return nil
}

func init() {
	cmdutil.Register(cmdutil.GroupProject, New{{.CLI.R.SingularName}}Cmd)
}
