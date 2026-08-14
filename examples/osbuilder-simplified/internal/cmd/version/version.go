package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
)

type Options struct {
	Short bool
	genericiooptions.IOStreams
}

func NewCmdVersion(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := &Options{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print the client version information",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, _ []string) {
			cmdutil.CheckErr(o.Complete())
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(f))
		},
	}
	cmd.Flags().BoolVar(&o.Short, "short", false, "print just the version number")
	return cmd
}

func (o *Options) Complete() error { return nil }
func (o *Options) Validate() error { return nil }
func (o *Options) Run(f cmdutil.Factory) error {
	if o.Short {
		fmt.Fprintln(o.Out, "v0.1.0")
		return nil
	}
	fmt.Fprintln(o.Out, "Client Version: v0.1.0")
	// 演示从 Factory 读取配置（对应 f.GetOptions()）
	cfg := f.GetOptions()
	if cfg.User != nil {
		fmt.Fprintf(o.Out, "Configured user: name=%q email=%q\n", cfg.User.Name, cfg.User.Email)
	}
	return nil
}
