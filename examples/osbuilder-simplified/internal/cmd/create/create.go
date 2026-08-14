package create

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/create/project"
)

// NewCmdCreate 对应 create.go: NewCmdCreate —— 只路由，挂载子命令
func NewCmdCreate(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "create [command]",
		Short:            "Create a new project",
		TraverseChildren: true,
		SilenceUsage:     true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	// 二级命令 —— 对应 create.go AddCommand(NewCmdProject)
	cmd.AddCommand(project.NewCmdProject(f, ioStreams))
	return cmd
}
