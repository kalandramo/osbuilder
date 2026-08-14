package options

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// NewCmdOptions 对应 cmd/options/options.go: NewCmdOptions
// 实现 `osbuilder options` 子命令：打印所有命令继承的持久 flag 列表。
// 注意：此包名为 options 是因为 "options" 表示命令行选项(flags)，
// 与 util/options（全局配置项）同名但职责完全不同，不可混淆。
func NewCmdOptions(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Print the list of flags inherited by all commands",
		Long:  "Print the list of flags inherited by all commands",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Usage()
		},
	}

	// Usage() 默认输出到 stderr，这里重定向到 out（stdout）
	cmd.SetOutput(out)
	cmd.SetErr(out)

	templates.UseOptionsTemplates(cmd)
	return cmd
}
