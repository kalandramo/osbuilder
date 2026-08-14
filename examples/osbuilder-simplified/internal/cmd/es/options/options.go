package options

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// errEmpty 校验失败时返回的可读错误。
func errEmpty(flag string) error {
	return fmt.Errorf("required flag %q not set", flag)
}

// ESOptions 是 es 子命令的【领域配置】（与全局 Options 解耦）。
// 仅服务于管理 ES 资源的子命令，不挂到根命令的 PersistentFlags，
// 因此不会污染其他子命令（create/version/color）的 flag 表。
type ESOptions struct {
	Addr     string `json:"addr" mapstructure:"addr"`
	Index    string `json:"index" mapstructure:"index"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

// NewESOptions 构造带默认值的领域配置。
func NewESOptions() *ESOptions {
	return &ESOptions{
		Addr:  "http://localhost:9200",
		Index: "default",
	}
}

// AddFlags 注册 es 子命令专用的局部 flag。
// 注意：这些 flag 只绑定到 es 命令自身的 cmd.Flags()（非 persist），
// 并经 viper.BindPFlags 镜像进 viper 的局部作用域。
func (o *ESOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Addr, "es.addr", o.Addr, "Elasticsearch address, e.g. http://localhost:9200")
	fs.StringVar(&o.Index, "es.index", o.Index, "Elasticsearch index name")
	fs.StringVar(&o.Username, "es.username", o.Username, "Elasticsearch username")
	fs.StringVar(&o.Password, "es.password", o.Password, "Elasticsearch password")
}

// Complete 把配置回填结构体，优先级严格对齐全局规则：
//
//	flag 默认值（NewESOptions） < osbuilder.yaml 的 es.* < 环境变量 OSCTL_ES_* < flag 显式值
//
// 实现要点：viper.UnmarshalKey 不会读取 BindPFlags 注册的 flag 默认值（viper v1.19 行为），
// 因此这里手动合并：先以「解析后的 flag 值（含默认值）」为底，若 viper 中该 key 被文件或
// 环境变量显式设置（IsSet），则用 viper 值覆盖，从而保证高优先级来源生效。
func (o *ESOptions) Complete(cmd *cobra.Command) {
	// 底：flag 解析后的当前值（未传则为 NewESOptions 默认值）
	if f := cmd.Flags().Lookup("es.addr"); f != nil {
		o.Addr = f.Value.String()
	}
	if f := cmd.Flags().Lookup("es.index"); f != nil {
		o.Index = f.Value.String()
	}
	if f := cmd.Flags().Lookup("es.username"); f != nil {
		o.Username = f.Value.String()
	}
	if f := cmd.Flags().Lookup("es.password"); f != nil {
		o.Password = f.Value.String()
	}
	// 高优先级覆盖：文件 / 环境变量（viper 中显式 IsSet 的来源）
	if viper.IsSet("es.addr") {
		o.Addr = viper.GetString("es.addr")
	}
	if viper.IsSet("es.index") {
		o.Index = viper.GetString("es.index")
	}
	if viper.IsSet("es.username") {
		o.Username = viper.GetString("es.username")
	}
	if viper.IsSet("es.password") {
		o.Password = viper.GetString("es.password")
	}
}

// Validate 校验领域配置必填项。
func (o *ESOptions) Validate() error {
	if o.Addr == "" {
		return errEmpty("es.addr")
	}
	if o.Index == "" {
		return errEmpty("es.index")
	}
	return nil
}
