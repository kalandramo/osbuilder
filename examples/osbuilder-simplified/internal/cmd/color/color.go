package color

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
)

// Options 严格对应 color.go 的 ColorOptions（内嵌 IOStreams）
type Options struct {
	Type string
	genericiooptions.IOStreams
}

func NewCmdColor(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := &Options{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:              "color",
		Short:            "Print colors supported by the current terminal",
		TraverseChildren: true,
		SilenceUsage:     true,
		SilenceErrors:    true,
		Run: func(cmd *cobra.Command, _ []string) {
			cmdutil.CheckErr(o.Complete())
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.Type, "type", "t", "", "color type: fg, bg, all")
	return cmd
}

func (o *Options) Complete() error { return nil }
func (o *Options) Validate() error {
	if o.Type != "" && o.Type != "fg" && o.Type != "bg" && o.Type != "all" {
		return fmt.Errorf("--type must be one of: fg, bg, all")
	}
	return nil
}
func (o *Options) Run() error {
	typ := o.Type
	if typ == "" {
		typ = "fg"
	}
	fmt.Fprintf(o.Out, "printing %s colors\n", typ)
	return nil
}
