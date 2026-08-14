package es

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/es/get"
)

// NewCmdES 是 es 子命令的中间节点（对应 create.NewCmdCreate 的角色）。
// 分组定位："Resource Commands (Elasticsearch)"。
func NewCmdES(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "es",
		Short: "Manage Elasticsearch resources",
		Long:  "Manage Elasticsearch resources such as indices and documents.",
	}
	cmd.AddCommand(get.NewCmdGet(f, ioStreams))
	return cmd
}
