package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/util"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/color"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/create"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/es"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/options"
	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd/version"
	clioptions "github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/util/options"
)

const (
	defaultConfigName = "osbuilder.yaml"
	defaultHomeDir    = ".onexstack"
)

// NewDefaultOSCtlCommand 对应 osbuilder cmd.go: NewDefaultOSCtlCommand（简化掉 plugin 逻辑）
func NewDefaultOSCtlCommand() *cobra.Command {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return NewOSCtlCommand(ioStreams)
}

// searchDirs 对应 cmd.go: searchDirs —— 返回配置搜索目录
func searchDirs() []string {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}

// NewOSCtlCommand 对应 osbuilder cmd.go: NewOSCtlCommand —— 建根 + 分组 + 配置 + AddCommand
func NewOSCtlCommand(ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := clioptions.NewOptions() // 对应 cmd.go: clioptions.NewOptions()

	cmds := &cobra.Command{
		Use:   "osbuilder",
		Short: i18n.T("osbuilder is a command-line tool for the onex technology stack scaffold"),
		Run: func(c *cobra.Command, _ []string) { _ = c.Help() },
		PersistentPreRunE: func(*cobra.Command, []string) error {
			// 对应 cmd.go:160 opts.Complete() —— 把 viper 值反序列化回填 Options
			opts.Complete()
			return nil
		},
	}

	// 注册配置 flag —— 对应 cmd.go:192 opts.AddFlags(flags)
	flags := cmds.PersistentFlags()
	opts.AddFlags(flags)

	// 持久 flag 镜像进 viper —— 对应 cmd.go:200 viper.BindPFlags
	_ = viper.BindPFlags(cmds.PersistentFlags())

	// 对应 cmd.go:201 —— 注册"执行前钩子"读配置文件+环境变量
	// 注意：真实代码用 cobra.OnInitialize(core.OnInitialize(...))，viper.OnInitialize 已废弃。
	// 这里用延迟求值 func() *string，在命令执行前（flag 已解析）才读取 --config，
	// 避免构造期 viper.GetString 取到空值导致 --config 失效。
	cobra.OnInitialize(cmdutil.OnInitialize(
		func() *string { return getConfigPtr(viper.GetString(clioptions.FlagConfig)) },
		"OSCTL",
		searchDirs(),
		defaultConfigName,
	))

	// 配置注入 Factory —— 对应 cmd.go:204 cmdutil.NewFactory(opts)
	f := cmdutil.NewFactory(opts)

	// 分组（仅在 help 中展示，不影响命令树）—— 对应 cmd.go CommandGroups
	groups := templates.CommandGroups{
		{
			Message: "Basic Commands (Beginner):",
			Commands: []*cobra.Command{
				color.NewCmdColor(f, ioStreams),
			},
		},
		{
			Message: "Project Commands:",
			Commands: []*cobra.Command{
				create.NewCmdCreate(f, ioStreams),
			},
		},
		{
			Message: "Resource Commands (Elasticsearch):",
			Commands: []*cobra.Command{
				es.NewCmdES(f, ioStreams),
			},
		},
	}
	groups.Add(cmds)
	templates.ActsAsRootCommand(cmds, []string{}, groups...)

	// 一级命令直接挂载 —— 对应 cmd.go AddCommand
	cmds.AddCommand(version.NewCmdVersion(f, ioStreams))
	cmds.AddCommand(options.NewCmdOptions(ioStreams.Out)) // `osbuilder options` 帮助子命令
	return cmds
}

// getConfigPtr 把 string 转 *string（对应 cmd.go: ptr.To(viper.GetString(...))）
func getConfigPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
