package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
	esoptions "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/es/options"
)

// Options 是 es get 的领域选项：内嵌 IOStreams + 领域配置 ESOptions。
type Options struct {
	esoptions.ESOptions
	genericiooptions.IOStreams
}

func NewCmdGet(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := &Options{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:              "get",
		Short:            "Get resources from Elasticsearch",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, _ []string) {
			cmdutil.CheckErr(o.Complete(cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	// 领域配置 flag 挂在命令自身（非 persist），不污染根命令。
	o.AddFlags(cmd.Flags())
	// 关键：领域 flag 镜像进 viper（与全局 BindPFlags(PersistentFlags) 时机一致，
	// 在构造期绑定，使 viper 能拿到 flag 默认值并参与「文件 < 环境变量 < flag」合并）。
	_ = viper.BindPFlags(cmd.Flags())
	return cmd
}

func (o *Options) Complete(cmd *cobra.Command) error {
	// 从 viper + flag 合并回填领域配置。
	o.ESOptions.Complete(cmd)
	return nil
}

func (o *Options) Validate() error {
	return o.ESOptions.Validate()
}

func (o *Options) Run() error {
	// 演示：仅打印连接信息，不真正连接 ES。
	fmt.Fprintf(o.Out, "connecting to ES addr=%s index=%s user=%s\n",
		o.Addr, o.Index, o.Username)
	return nil
}
