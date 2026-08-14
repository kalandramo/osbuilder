package project

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
)

// Options 对应 create_project.go 的 ProjectOptions（内嵌 IOStreams）
type Options struct {
	Config string
	genericiooptions.IOStreams
}

func NewCmdProject(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := &Options{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:              "project [DIR]",
		Short:            "Create a new project from a config file",
		Args:             cobra.MaximumNArgs(1),
		TraverseChildren: true,
		SilenceUsage:     true,
		SilenceErrors:    true,
		Run: func(cmd *cobra.Command, _ []string) {
			cmdutil.CheckErr(o.Complete())
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.Config, "config", "c", "", "path to project config file")
	return cmd
}

func (o *Options) Complete() error { return nil }
func (o *Options) Validate() error { return nil }
func (o *Options) Run() error {
	fmt.Fprintf(o.Out, "creating project with config=%q\n", o.Config)
	return nil
}
