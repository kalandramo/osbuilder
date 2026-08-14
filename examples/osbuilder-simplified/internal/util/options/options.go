package options

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// 对应 osbuilder util/options/options.go 的 FlagConfig
const FlagConfig = "config"

// Options 对应 osbuilder 的 Options（简化掉 usercenter/gateway，仅保留 user 演示）
type Options struct {
	Config string
	User   *UserOptions `json:"user" mapstructure:"user"`
}

// UserOptions 演示业务配置分层
type UserOptions struct {
	Name  string `json:"name" mapstructure:"name"`
	Email string `json:"email" mapstructure:"email"`
}

// NewOptions 对应 NewOptions —— 构造带默认值的配置结构体
func NewOptions() *Options {
	return &Options{
		User: &UserOptions{},
	}
}

// AddFlags 对应 Options.AddFlags —— 把配置项注册成持久 flag（--config / --user.name / --user.email）
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Config, FlagConfig, o.Config, "Path to the osbuilder.yaml file to use for CLI.")
	fs.StringVar(&o.User.Name, "user.name", o.User.Name, "user name")
	fs.StringVar(&o.User.Email, "user.email", o.User.Email, "user email")
}

// Complete 对应 Options.Complete —— 把 viper 已加载的值反序列化回填结构体
func (o *Options) Complete() {
	if err := viper.Unmarshal(o); err != nil {
		panic(err)
	}
}
